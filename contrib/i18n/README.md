# i18n
golang i18n, golang实现的多语言解析使用
# 安装
- go get
```
go get github.com/go-fox/contrib/i18n
```
# 使用
```shell
# 创建文件夹
mkdir -p language

# 编写中文语言文件
cat >>demo/language/zh_cn.toml<<EOF
{
  "user": {
    "test": "test",
    "username":"username",
    "user_not_found": "user not found",
    "parameter_error":"parameter {{.Parameter}} error"
  }
}
EOF

# 编写中文语言文件
cat >>demo/language/en_us.json<<EOF
[user]
test="测试"
username="用户名"
user_not_found="用户未找到"
parameter_error="参数：{{.Parameter}}错误"
EOF
```
编写go代码文件demoi18n/main.go
```go
package main

import (
	"fmt"

	"github.com/go-fox/contrilb/i18n"
)

func main() {
	n, err := i18n.New(
		i18n.WithSources(i18n.NewFileSource(
			"./language",
		)),
	)
	if err != nil {
		panic(err)
	}
	// 使用原始方法输出zh_cn的test
	fmt.Printf("user.test zh_cn is: %s\n", n.T("zh_cn", "user.test"))
	// 使用原始方法输出en_us的test
	fmt.Printf("user.test en_us is: %s\n", n.T("en_us", "user.test"))
	// 使用原始方法输出en_us的parameter_error,带参数模板
	fmt.Printf("user.parameter_error en_us is: %s\n", n.T("en_us", "user.parameter_error", map[string]any{
		"Parameter": "cat",
	}))
	// 使用原始方法输出en_us的parameter_error,带参数模板
	fmt.Printf("user.parameter_error zh_cn is: %s\n", n.T("zh_cn", "user.parameter_error", map[string]any{
		"Parameter": "cat",
	}))
	// 构建作用域的translate
	scope := n.Locale("zh_cn").Scope("user")
	fmt.Printf("test zh_cn is: %s\n", scope.T("test"))
	fmt.Printf("parameter_error zh_cn is: %s\n", scope.T("parameter_error", map[string]any{
		"Parameter": "cat",
	}))
}
```
