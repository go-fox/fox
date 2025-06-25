package main

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"
)

//go:embed httpTemplate.tpl
var httpTemplate string

type serviceDesc struct {
	ServiceType    string // Greeter
	ServiceName    string // helloworld.Greeter
	ServiceComment string // 注册服务
	Metadata       string // api/helloworld/helloworld.proto
	Methods        []*methodDesc
	MethodSets     map[string]*methodDesc
}

// UploadFields 上传字段
type UploadFields struct {
	GoName   string // 上传字段的 Go 名称
	JSONName string // 上传字段的 JSON 名称
	Name     string // 上传字段的名称
	IsList   bool
}

type methodDesc struct {
	Upload               bool // 是否是上传方法
	UploadFields         []UploadFields
	FileQualifiedGoIdent string
	// method
	Name         string
	OriginalName string // The parsed original name
	Num          int
	Request      string
	Reply        string
	Comment      string
	// http_rule
	Path                string
	Method              string
	HasVars             bool
	HasBody             bool
	Body                string
	ResponseBody        string
	HttpFuncComment     string
	NoMethodNameComment string
	RouteInfoComment    string
}

func (s *serviceDesc) execute() string {
	s.MethodSets = make(map[string]*methodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(httpTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(buf.String(), "\r\n")
}
