package src

import (
	"fmt"
	"go/types"
	"maps"
	"os"
	"path/filepath"
	"strings"
)

// GenerateCode generates implementation code for an interface
func GenerateCode(iface InterfaceInfo, outputDir string) error {
	// Check if interface has methods
	if len(iface.Methods) == 0 {
		return fmt.Errorf("interface %s has no methods", iface.Name)
	}

	// Generate implementation structure
	implStruct, err := generateImplStruct(iface)
	if err != nil {
		return fmt.Errorf("failed to generate implementation structure: %v", err)
	}

	// Generate implementations for each method
	for _, method := range iface.Methods {
		implStruct, err = generateMethodImplementation(iface, method, implStruct)
		if err != nil {
			return fmt.Errorf("failed to generate implementation for method %s: %v", method.Name, err)
		}
	}

	// Write implementation to file
	if err := writeImplStructToFile(implStruct, outputDir); err != nil {
		return fmt.Errorf("failed to write implementation to file: %v", err)
	}

	return nil
}

// generateImplStruct creates the basic structure for the implementation class
func generateImplStruct(iface InterfaceInfo) (string, error) {
	// Create implementation name (interface name + Impl)
	implName := iface.Name + "Impl"

	// Create basic structure with package declaration
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("package %s\n\n", iface.PackageName))

	// Add import statements
	if len(iface.Depends) > 0 {
		sb.WriteString("import (\n")
		for _, importPath := range iface.Depends {
			sb.WriteString(fmt.Sprintf("\t\"%s\"\n", importPath))
		}
		sb.WriteString(")\n\n")
	}

	// Add type definition
	sb.WriteString(fmt.Sprintf("// Auto-generated implementation of %s interface\ntype %s struct {\n}\n\n",
		iface.Name, implName))

	return sb.String(), nil
}

// generateMethodImplementation creates the implementation for a single method
func generateMethodImplementation(iface InterfaceInfo, method MethodInfo, implStruct string) (string, error) {
	if len(method.Params) == 0 || len(method.Results) == 0 {
		return "", fmt.Errorf("method %s must have at least one parameter and one return value", method.Name)
	}

	// Get parameter and return types
	sourceType := method.Params[0].Type
	targetType := method.Results[0].Type

	// Create method implementation template
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf(`
// %s implements conversion logic
func (a *%sImpl) %s(%s %s) %s {
	target := %s{}

`, method.Name, iface.Name, method.Name, method.Params[0].Name, sourceType, targetType, targetType))

	// Process type information for mapping
	dependenceMap, err := buildDependencyMap(iface, sourceType, targetType)
	if err != nil {
		return "", err
	}

	// Get struct information
	targetStruct, err := getStruct(targetType, dependenceMap)
	if err != nil {
		return "", fmt.Errorf("failed to get target struct info: %v", err)
	}

	sourceStruct, err := getStruct(sourceType, dependenceMap)
	if err != nil {
		return "", fmt.Errorf("failed to get source struct info: %v", err)
	}

	// Add field mapping logic for matching field names
	generateFieldMappings(&sb, method.Params[0].Name, method.Comment, targetStruct, sourceStruct)

	sb.WriteString(`
	return target
}
`)

	// Add method implementation to the structure
	return implStruct + sb.String(), nil
}

// getStruct 获取结构体信息
func getStruct(targetType string, dependenceMap map[string]string) (*types.Struct, error) {
	targetTypeParts := strings.Split(targetType, ".")
	targetStruct, err := getStructInfo(dependenceMap[targetTypeParts[1]], targetTypeParts[1])
	if err != nil {
		return nil, err
	}

	return targetStruct, nil
}

// buildDependencyMap creates a map of type names to their package paths
func buildDependencyMap(iface InterfaceInfo, sourceType, targetType string) (map[string]string, error) {
	targetTypeParts := strings.Split(targetType, ".")
	sourceTypeParts := strings.Split(sourceType, ".")
	dependenceMap := make(map[string]string)

	// Find the corresponding packages for source and target types
	for _, dependencePath := range iface.Depends {
		if len(dependenceMap) == 2 {
			break
		} else if strings.LastIndex(dependencePath, targetTypeParts[0]) != -1 {
			dependenceMap[targetTypeParts[1]] = dependencePath
			continue
		} else if strings.LastIndex(dependencePath, sourceTypeParts[0]) != -1 {
			dependenceMap[sourceTypeParts[1]] = dependencePath
			continue
		}
	}

	return dependenceMap, nil
}

// generateFieldMappings generates code to map fields with matching names
func generateFieldMappings(sb *strings.Builder, paramName string, methodComment []string, targetStruct, sourceStruct *types.Struct) {
	// get targetFildName and sourceFieldName
	targetSourceMap := make(map[string]string)

	for _, comment := range methodComment {
		targetSourceMap2, err := parseMethodComment(comment)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		maps.Copy(targetSourceMap, targetSourceMap2)
	}

	for i := range targetStruct.NumFields() {
		if _, ok := targetSourceMap[targetStruct.Field(i).Name()]; ok {
			continue
		}

		targetFieldName := targetStruct.Field(i).Name()

		for j := range sourceStruct.NumFields() {
			sourceFieldName := sourceStruct.Field(j).Name()
			if targetFieldName == sourceFieldName {
				targetSourceMap[targetFieldName] = sourceFieldName
				break
			}
		}
	}

	// map the fields
	for targetField, sourceField := range targetSourceMap {
		sb.WriteString(fmt.Sprintf("\ttarget.%s = %s.%s\n",
			targetField, paramName, sourceField))
	}
}

// writeImplStructToFile writes the implementation to a file
func writeImplStructToFile(implStruct string, outputDir string) error {
	// Extract package name and struct name from content
	_, structName, err := extractPackageAndStructName(implStruct)
	if err != nil {
		return err
	}

	// Create file path
	fileName := strings.ToLower(structName) + ".go"
	filePath := filepath.Join(outputDir, fileName)

	// Write to file
	if err := os.WriteFile(filePath, []byte(implStruct), 0644); err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	fmt.Printf("Generated implementation file: %s\n", filePath)
	return nil
}

// extractPackageAndStructName extracts package and struct names from generated code
func extractPackageAndStructName(implStruct string) (packageName string, structName string, err error) {
	lines := strings.Split(implStruct, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "package ") {
			packageName = strings.TrimPrefix(line, "package ")
		} else if strings.Contains(line, "type ") && strings.Contains(line, "Impl struct") {
			parts := strings.Split(line, " ")
			for i, part := range parts {
				if part == "type" && i+1 < len(parts) {
					structName = parts[i+1]
					break
				}
			}
		}
		if packageName != "" && structName != "" {
			return packageName, structName, nil
		}
	}

	return "", "", fmt.Errorf("could not extract package or struct name from generated code")
}

// ProcessInterface handles processing of an interface for code generation
func ProcessInterface(iface InterfaceInfo, outputDir string) error {
	fmt.Printf("Processing interface: %s\n", iface.Name)

	// Generate code
	if err := GenerateCode(iface, outputDir); err != nil {
		return fmt.Errorf("code generation failed: %v", err)
	}

	return nil
}
