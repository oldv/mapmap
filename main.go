package main

import (
	"flag"
	"fmt"
	"github.com/oldv/mapmap/src"
	"os"
	"path/filepath"
)

func main() {
	// 定义命令行子命令
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// 解析子命令
	switch os.Args[1] {
	case "generate":
		generateCmd()
	case "help":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

// 打印使用帮助
func printUsage() {
	fmt.Println("mapmap - 根据注释生成结构体转换代码")
	fmt.Println("\n命令:")
	fmt.Println("  generate  生成转换类代码")
	fmt.Println("    -f      指定文件")
	fmt.Println("    -d      指定目录")
	fmt.Println("    -o      输出目录(必填)")
	fmt.Println("  help      显示此帮助信息")
}

// 处理 generate 命令
func generateCmd() {
	// 为 generate 命令创建一个新的 FlagSet
	genCmd := flag.NewFlagSet("generate", flag.ExitOnError)

	// 定义命令参数
	filePath := genCmd.String("f", "", "指定文件路径")
	dirPath := genCmd.String("d", "", "指定目录路径")
	outputDir := genCmd.String("o", "", "输出目录(必填)")

	// 解析参数，注意要跳过子命令本身
	genCmd.Parse(os.Args[2:])

	// 验证参数
	if *outputDir == "" {
		fmt.Println("错误: 必须指定输出目录 (-o)")
		genCmd.Usage()
		os.Exit(1)
	}

	// 检查文件或目录是否存在
	if *filePath != "" && *dirPath != "" {
		fmt.Println("错误: -f 和 -d 不能同时使用")
		os.Exit(1)
	}

	if *filePath == "" && *dirPath == "" {
		fmt.Println("错误: 必须指定文件 (-f) 或目录 (-d)")
		os.Exit(1)
	}

	// 创建输出目录（如果不存在）
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("错误: 创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	// 开始扫描和生成代码
	if *filePath != "" {
		processFile(*filePath, *outputDir)
	} else {
		processDirectory(*dirPath, *outputDir)
	}
}

// 处理单个文件
func processFile(filePath string, outputDir string) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("错误: 文件不存在: %s\n", filePath)
		os.Exit(1)
	}

	fmt.Printf("处理文件: %s\n", filePath)

	// 解析文件
	interfaces, err := src.ParseFile(filePath)
	if err != nil {
		fmt.Printf("解析文件失败: %v\n", err)
		os.Exit(1)
	}

	// 处理找到的接口
	for _, iface := range interfaces {
		fmt.Printf("找到接口: %s 在包 %s 中\n", iface.Name, iface.PackageName)

		// 生成转换代码
		if err := src.ProcessInterface(iface, outputDir); err != nil {
			fmt.Printf("处理接口 %s 失败: %v\n", iface.Name, err)
			continue
		}

		// 输出找到的方法
		for _, method := range iface.Methods {
			fmt.Printf("  方法: %s\n", method.Name)
			fmt.Printf("    参数: ")
			for _, param := range method.Params {
				fmt.Printf("%s %s, ", param.Name, param.Type)
			}
			fmt.Println()

			fmt.Printf("    返回: ")
			for _, result := range method.Results {
				fmt.Printf("%s %s, ", result.Name, result.Type)
			}
			fmt.Println()
		}
	}
}

// 处理目录
func processDirectory(dirPath string, outputDir string) {
	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		fmt.Printf("错误: 目录不存在: %s\n", dirPath)
		os.Exit(1)
	}

	fmt.Printf("处理目录: %s\n", dirPath)

	// 遍历目录中的所有 .go 文件
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			fmt.Printf("扫描文件: %s\n", path)

			// 解析文件
			interfaces, err := src.ParseFile(path)
			if err != nil {
				fmt.Printf("解析文件失败 %s: %v\n", path, err)
				return nil // 继续处理其他文件
			}

			// 处理找到的接口
			for _, iface := range interfaces {
				fmt.Printf("找到接口: %s 在包 %s 中\n", iface.Name, iface.PackageName)

				// 生成转换代码
				if err := src.ProcessInterface(iface, outputDir); err != nil {
					fmt.Printf("处理接口 %s 失败: %v\n", iface.Name, err)
					continue
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("错误: 遍历目录失败: %v\n", err)
		os.Exit(1)
	}
}
