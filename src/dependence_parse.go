package src

import (
	"fmt"
	"go/importer"
	"go/token"
	"go/types"
)

// 获取结构体信息
// - dependencePath: 依赖路径
// - structName: 结构体名称
// - 返回: 结构体信息
func getStructInfo(dependencePath string, structName string) (*types.Struct, error) {

	// 获取依赖的包
	pkg, err := getLocalPackageInfo(dependencePath)
	if err != nil {
		fmt.Println("Error getting package info:", err)
		return nil, err
	}

	// 获取结构体信息
	structInfo, err := doGetStructInfo(pkg, structName)
	if err != nil {
		fmt.Println("Error getting struct info:", err)
		return nil, err
	}

	return structInfo, nil
}

// 获取包信息
// - pkgPath: 包路径
// - 返回: 包信息
func getLocalPackageInfo(pkgPath string) (*types.Package, error) {
	// 使用 ForCompiler 创建一个本地包导入器
	imp := importer.ForCompiler(token.NewFileSet(), "source", nil)
	// 导入包
	pkg, err := imp.Import(pkgPath)
	if err != nil {
		return nil, err
	}
	return pkg, nil
}

// 获取结构体信息
// - pkg: 包信息
// - structName: 结构体名称
// - 返回: 结构体信息
func doGetStructInfo(pkg *types.Package, structName string) (*types.Struct, error) {
	// 遍历包中的所有对象
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		// 检查对象是否为类型声明
		if typeName, ok := obj.(*types.TypeName); ok {
			// 检查类型是否为结构体
			if structType, ok := typeName.Type().Underlying().(*types.Struct); ok {
				if typeName.Name() == structName {
					return structType, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("struct %s not found in package %s", structName, pkg.Path())
}
