# mapmap

## 功能
通过命令的方式通过反射生成转换类
允许使用注释控制执行结果

注释例子
```
// mapmap:assembler
type UserAsm interface {
	// mapmap:source:"name",target:"name" 指定字段来源字段
	// mapmap:target:"age",ignore 忽略字段
	Convert(dto UserDTO) User
}
```

1. 首先遍历所有文件，找到所有有注释的接口
如:
```
// mapmap:assembler
type UserAsm interface {
	Convert(dto UserDTO) User
}
```

1. 在根据注释生成转换类
获取接口方法的参数类和返回类的导入路径
使用反射获取返回类与参数类的所有字段
根据方法注释生成转换代码, 没有注释的则使用同名字段

## 命令

### generate

**参数**
1. -f 指定文件
2. -d 指定目录
3. -o 输出目录(必填)


### help
提示如何使用generate