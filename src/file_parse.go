package src

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// 表示解析到的接口信息
type InterfaceInfo struct {
	Name        string       // 接口名称
	PackageName string       // 包名
	Depends     []string     // 依赖的包
	Methods     []MethodInfo // 接口方法
	FilePath    string       // 文件路径
	Comment     string       // 注释
}

// 表示接口方法信息
type MethodInfo struct {
	Name    string      // 方法名称
	Params  []ParamInfo // 参数信息
	Results []ParamInfo // 返回值信息
	Comment []string    // 方法注释
}

// 表示参数或返回值信息
type ParamInfo struct {
	Name string // 参数名称
	Type string // 参数类型
}

// ParseFile 解析单个Go文件，查找带有特定注释的接口定义
func ParseFile(filePath string) ([]InterfaceInfo, error) {
	// 创建文件集
	fset := token.NewFileSet()

	// 解析Go文件
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("解析文件失败: %v", err)
	}

	// 获取包名
	packageName := file.Name.Name

	// 获取依赖的包
	depends := []string{}
	for _, spec := range file.Imports {
		depends = append(depends, strings.Trim(spec.Path.Value, "\""))
	}

	// 用于存储找到的接口信息
	var interfaces []InterfaceInfo

	// 遍历所有顶级声明
	for _, decl := range file.Decls {
		// 查找类型声明（如接口定义）
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		// 检查注释
		if genDecl.Doc == nil {
			continue
		}

		// 查找特定注释
		hasMapMapComment := false
		for _, comment := range genDecl.Doc.List {
			if strings.Contains(comment.Text, "mapmap:assembler") {
				hasMapMapComment = true
				break
			}
		}

		if !hasMapMapComment {
			continue
		}

		// 处理类型规范
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			// 验证是否为接口类型
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}

			// 创建接口信息对象
			ifaceInfo := InterfaceInfo{
				Name:        typeSpec.Name.Name,
				PackageName: packageName,
				FilePath:    filePath,
				Comment:     "mapmap:assembler",
				Depends:     depends,
			}

			// 解析接口方法
			if interfaceType.Methods != nil {
				for _, method := range interfaceType.Methods.List {
					if len(method.Names) == 0 {
						continue // 嵌入接口
					}

					methodName := method.Names[0].Name
					methodType, ok := method.Type.(*ast.FuncType)
					if !ok {
						continue
					}

					methodInfo := MethodInfo{
						Name: methodName,
					}

					// 解析方法参数
					if methodType.Params != nil {
						methodInfo.Params = parseFieldList(methodType.Params)
					}

					// 解析方法返回值
					if methodType.Results != nil {
						methodInfo.Results = parseFieldList(methodType.Results)
					}

					// 获取方法注释
					if method.Doc != nil && len(method.Doc.List) > 0 {
						for _, comment := range method.Doc.List {
							methodInfo.Comment = append(methodInfo.Comment, comment.Text)
						}
					}

					ifaceInfo.Methods = append(ifaceInfo.Methods, methodInfo)
				}
			}

			interfaces = append(interfaces, ifaceInfo)
		}
	}

	return interfaces, nil
}

// 解析字段列表（参数或返回值）
func parseFieldList(fieldList *ast.FieldList) []ParamInfo {
	var params []ParamInfo

	for _, field := range fieldList.List {
		// 获取类型表示和包路径
		typeExpr := field.Type
		typeStr := exprToString(typeExpr)

		// 处理可能的多个名称
		if len(field.Names) == 0 {
			// 匿名参数
			params = append(params, ParamInfo{
				Name: "",
				Type: typeStr,
			})
		} else {
			for _, name := range field.Names {
				params = append(params, ParamInfo{
					Name: name.Name,
					Type: typeStr,
				})
			}
		}
	}

	return params
}

// 将类型表达式转换为字符串
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + exprToString(t.Elt)
		}
		// 固定长度数组暂不处理
		return "array"
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	default:
		return "未知类型"
	}
}

// 创建一个用于生成代码的结构体信息
type StructInfo struct {
	Name   string
	Fields []FieldInfo
}

// 字段信息
type FieldInfo struct {
	Name      string
	Type      string
	MapSource string // mapsource 标签值
}

// 查找并解析结构体信息
func FindStructInfo(filePath string, structName string) (*StructInfo, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != structName {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structInfo := &StructInfo{
				Name: structName,
			}

			// 解析结构体字段
			if structType.Fields != nil {
				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						continue // 匿名字段
					}

					fieldName := field.Names[0].Name
					fieldType := exprToString(field.Type)

					fieldInfo := FieldInfo{
						Name: fieldName,
						Type: fieldType,
					}

					// 解析 mapsource 标签
					if field.Tag != nil {
						tag := field.Tag.Value
						mapSourceTag := extractTag(tag, "mapsource")
						fieldInfo.MapSource = mapSourceTag
					}

					structInfo.Fields = append(structInfo.Fields, fieldInfo)
				}
			}

			return structInfo, nil
		}
	}

	return nil, fmt.Errorf("未找到结构体: %s", structName)
}

// 从标签字符串中提取特定标签
func extractTag(tagStr string, tagName string) string {
	// 移除首尾的反引号
	tagStr = strings.Trim(tagStr, "`")

	// 寻找目标标签
	tagPrefix := tagName + ":\""
	start := strings.Index(tagStr, tagPrefix)
	if start == -1 {
		return ""
	}

	// 移动到标签值的开始
	start += len(tagPrefix)
	end := strings.Index(tagStr[start:], "\"")
	if end == -1 {
		return ""
	}

	return tagStr[start : start+end]
}
