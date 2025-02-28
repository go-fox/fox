package generator

import (
	"bytes"
	"strings"
	"text/template"
	"time"

	"google.golang.org/protobuf/compiler/protogen"
)

const versionFormat = "20060102150405"

// NewVersion generates a new migration version.
func NewVersion() string {
	return time.Now().UTC().Format(versionFormat)
}

const tpl = "INSERT INTO `{{.Table}}` (create_at, update_at, delete_at, title, operation) VALUES (CURRENT_TIMESTAMP(),CURRENT_TIMESTAMP(),null,'{{.Title}}','{{.Operation}}') on duplicate key update  update_at = CURRENT_TIMESTAMP(), name = VALUES(title), operation = VALUES(operation);"

// Configuration 配置
type Configuration struct {
	OutputMode *string
	Table      *string
	Version    *string
	Filename   *string
}

// FoxRouteSQLGenerator 路由生成器
type FoxRouteSQLGenerator struct {
	config Configuration
	files  []*protogen.File
	plugin *protogen.Plugin
}

// FoxRoute 路由信息
type FoxRoute struct {
	Table     string // 表名
	Title     string // 标题
	Operation string // 操作
}

// NewFoxRouteSQLGenerator 创建路由生成器
func NewFoxRouteSQLGenerator(plugin *protogen.Plugin, conf Configuration, files []*protogen.File) *FoxRouteSQLGenerator {
	return &FoxRouteSQLGenerator{
		plugin: plugin,
		config: conf,
		files:  files,
	}
}

// Run 执行生成
func (f *FoxRouteSQLGenerator) Run(outputFile *protogen.GeneratedFile) error {
	parse, err := template.New("").Parse(tpl)
	if err != nil {
		return err
	}
	for _, file := range f.files {
		for _, service := range file.Services {
			for _, method := range service.Methods {
				Route := &FoxRoute{
					Table:     *f.config.Table,
					Title:     strings.TrimSuffix(strings.ReplaceAll(method.Comments.Leading.String(), "//", ""), "\n"),
					Operation: "/" + string(service.Desc.FullName()) + "/" + string(method.Desc.Name()),
				}
				buffer := bytes.NewBuffer([]byte{})
				if err := parse.Execute(buffer, Route); err != nil {
					return err
				}
				buffer.WriteString("\n")
				_, err := outputFile.Write(buffer.Bytes())
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
