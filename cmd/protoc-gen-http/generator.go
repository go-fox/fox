// Package http
// MIT License
//
// # Copyright (c) 2024 go-fox
// Author https://github.com/go-fox/fox
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/go-fox/fox/api/annotations"
	"github.com/go-fox/fox/transport/http"
)

const (
	contextPackage     = protogen.GoImportPath("context")
	httpPackage        = protogen.GoImportPath("github.com/go-fox/fox/transport/http")
	annotationsPackage = protogen.GoImportPath("github.com/go-fox/fox/api/annotations")
)

var methodSets = make(map[string]int)

// Generator 定义生成器
type Generator struct {
	gen         *protogen.Plugin
	file        *protogen.File
	writer      *protogen.GeneratedFile
	httpPackage string
	omitempty   bool // 是否忽略掉没有配置路由的方法
	version     string
}

// GenerateFile 生成http文件
func GenerateFile(gen *protogen.Plugin, file *protogen.File, omitempty *bool, v string) {
	filename := file.GeneratedFilenamePrefix + "_http.pb.go"
	w := gen.NewGeneratedFile(filename, file.GoImportPath)
	g := Generator{
		gen:       gen,
		file:      file,
		omitempty: *omitempty,
		writer:    w,
		version:   v,
	}
	g.Run()
}

// Run 开始生成
func (g *Generator) Run() {
	g.P("// Code generated by protoc-gen-fox. DO NOT EDIT.")
	g.P("// versions:")
	g.P(fmt.Sprintf("// - protoc-gen-fox %s", g.version))
	g.P("// - protoc             ", g.getProtocVersion())
	if g.file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", g.file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", g.file.Desc.Path())
	}
	g.P()
	g.P("package ", g.file.GoPackageName)
	g.P()
	g.genFileContent()
}

// writeFileContent 写出文件内容
func (g *Generator) genFileContent() {
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the ceres package it is being compiled against.")
	g.P("var _ = new(", contextPackage.Ident("Context"), ")")
	g.P("const _ = ", httpPackage.Ident("SupportPackageIsVersion1"))
	g.P()
	str := g.writer.QualifiedGoIdent(protogen.GoIdent{
		GoName:       "ceres-go",
		GoImportPath: httpPackage,
	})
	g.httpPackage = strings.TrimSuffix(str, ".ceres-go")
	for _, service := range g.file.Services {
		g.genService(service)
	}
}

// genService 生成
func (g *Generator) genService(service *protogen.Service) {

	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	// HTTP Server.
	sd := &serviceDesc{
		HttpPath:    g.httpPackage,
		ServiceType: service.GoName,
		ServiceName: string(service.Desc.FullName()),
		Comments:    service.Comments.Leading.String(),
		Metadata:    g.file.Desc.Path(),
	}
	for _, method := range service.Methods {
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue
		}
		rule, ok := proto.GetExtension(method.Desc.Options(), annotations.E_Method).(*annotations.MethodRule)
		if rule != nil && ok {
			for _, bind := range rule.Http.AdditionalBindings {
				httpRule := g.buildHTTPRule(method, bind)
				if httpRule != nil {
					sd.Methods = append(sd.Methods)
				}
			}
			httpRule := g.buildHTTPRule(method, rule.Http)
			if httpRule != nil {
				sd.Methods = append(sd.Methods, g.buildHTTPRule(method, rule.Http))
			}
		} else if !g.omitempty {
			path := fmt.Sprintf("/%s/%s", service.Desc.FullName(), method.Desc.Name())
			sd.Methods = append(sd.Methods, g.buildMethodDesc(method, http.MethodPost, path))
		}
	}
	if len(sd.Methods) != 0 {
		g.P(sd.execute())
	}
}

func (g *Generator) buildHTTPRule(m *protogen.Method, rule *annotations.HttpRule) *methodDesc {
	var (
		path   string
		method string
	)
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = "Get"
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = "Put"
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = "Post"
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = "Delete"
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = "Patch"
	case *annotations.HttpRule_Options:
		path = pattern.Options
		method = "Options"
	case *annotations.HttpRule_Head:
		path = pattern.Head
		method = "Head"
	case *annotations.HttpRule_Trace:
		path = pattern.Trace
		method = "Trace"
	case *annotations.HttpRule_Connect:
		path = pattern.Connect
		method = "Connect"
	default:
		return nil
	}
	md := g.buildMethodDesc(m, method, path)
	return md
}

func (g *Generator) buildMethodDesc(m *protogen.Method, method, path string) *methodDesc {
	defer func() { methodSets[m.GoName]++ }()
	return &methodDesc{
		Upload:       m.Input.Desc.FullName() == "fox.api.UploadRequest",
		Name:         m.GoName,
		OriginalName: string(m.Desc.Name()),
		Num:          methodSets[m.GoName],
		Request:      g.writer.QualifiedGoIdent(m.Input.GoIdent),
		Reply:        g.writer.QualifiedGoIdent(m.Output.GoIdent),
		Path:         path,
		Comments:     m.Comments.Leading.String(),
		Method:       method,
	}
}

// printf 打印日志
func (g *Generator) printf(msg string, arg ...interface{}) {
	fmt.Fprintf(os.Stderr, msg, arg...)
}

// getProtocVersion 获取protoc版本
func (g *Generator) getProtocVersion() string {
	v := g.gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

// P 写入信息到文件
func (g *Generator) P(v ...interface{}) {
	g.writer.P(v...)
}

// hasHTTPRule 判断是否包含http路由
func hasHTTPRule(file *protogen.File) bool {
	for _, service := range file.Services {
		for _, method := range service.Methods {
			if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
				continue
			}
			rule, ok := proto.GetExtension(method.Desc.Options(), annotations.E_Method).(*annotations.MethodRule)
			if rule != nil && ok {
				switch rule.Http.Pattern.(type) {
				case *annotations.HttpRule_Get:
					return true
				case *annotations.HttpRule_Put:
					return true
				case *annotations.HttpRule_Post:
					return true
				case *annotations.HttpRule_Delete:
					return true
				case *annotations.HttpRule_Patch:
					return true
				case *annotations.HttpRule_Options:
					return true
				case *annotations.HttpRule_Trace:
					return true
				case *annotations.HttpRule_Connect:
					return true
				case *annotations.HttpRule_Head:
					return true
				}
			}
		}
	}
	return false
}

func replacePath(name string, value string, path string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?i){([\s]*%s[\s]*)=?([^{}]*)}`, name))
	idx := pattern.FindStringIndex(path)
	if len(idx) > 0 {
		path = fmt.Sprintf("%s{%s:%s}%s",
			path[:idx[0]], // The start of the match
			name,
			strings.ReplaceAll(value, "*", ".*"),
			path[idx[1]:],
		)
	}
	return path
}

func camelCaseVars(s string) string {
	subs := strings.Split(s, ".")
	vars := make([]string, 0, len(subs))
	for _, sub := range subs {
		vars = append(vars, camelCase(sub))
	}
	return strings.Join(vars, ".")
}

// camelCase returns the CamelCased name.
// If there is an interior underscore followed by a lower case letter,
// drop the underscore and convert the letter to upper case.
// There is a remote possibility of this rewrite causing a name collision,
// but it's so remote we're prepared to pretend it's nonexistent - since the
// C++ generator lowercase names, it's extremely unlikely to have two fields
// with different capitalization.
// In short, _my_field_name_2 becomes XMyFieldName_2.
func camelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

const deprecationComment = "// Deprecated: Do not use."
