package generator

import (
	_ "embed"
	"errors"
	"github.com/emicklei/proto"
	"github.com/go-fox/fox/cmd/protoc-gen-fox-migrate-route/generator/model"
	"github.com/go-fox/fox/cmd/protoc-gen-fox-migrate-route/generator/templatex"
	"google.golang.org/protobuf/compiler/protogen"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const versionFormat = "20060102150405"

//go:embed template.go.tpl
var httpTemplate string

// NewVersion generates a new migration version.
func NewVersion() string {
	return time.Now().UTC().Format(versionFormat)
}

// Configuration 配置
type Configuration struct {
	Src        *string // 源文件目录
	Table      *string // 表名
	Package    *string // 生成文件目录
	Filename   *string // 文件名
	Version    *bool   // 版本号
	EntPackage *string // entgo
}

// FoxRouteSQLGenerator 路由生成器
type FoxRouteSQLGenerator struct {
	plugin    *protogen.Plugin
	config    Configuration
	protos    []*proto.Proto
	srcDir    string
	goImport  string
	goPackage string
	funcName  string
	module    *Module
	enums     []*proto.Enum
}

// FoxRoute 路由信息
type FoxRoute struct {
	ID    string // id
	Title string // 标题
}

// NewFoxRouteSQLGenerator 创建路由生成器
func NewFoxRouteSQLGenerator(conf Configuration, plugin *protogen.Plugin) (*FoxRouteSQLGenerator, error) {
	protos, err := readParser(*conf.Src)
	if err != nil {
		return nil, err
	}
	return &FoxRouteSQLGenerator{
		config: conf,
		protos: protos,
		plugin: plugin,
	}, nil
}

// Run 执行生成
func (f *FoxRouteSQLGenerator) Run() error {
	routes, err := f.parserRoute()
	if err != nil {
		return err
	}
	var entImport string
	// 读取当前模块的ent所在的目录
	if len(*f.config.EntPackage) > 0 {
		entImport = *f.config.EntPackage
	} else {
		entImport = f.getEntImport()
		if len(entImport) == 0 {
			return errors.New("ent path not found")
		}
	}

	f.goPackage = *f.config.Package
	f.funcName = ToCamelCase(*f.config.Filename, true)
	var goImports []string
	goImports = append(goImports, `"`+entImport+`"`)

	parse := templatex.With("insert_router").Parse(httpTemplate).GoFmt(true)

	// 创建文件
	file := f.plugin.NewGeneratedFile(*f.config.Filename+".go", "")

	execute, err := parse.Execute(map[string]any{
		"Imports":     strings.Join(goImports, "\n"),
		"PackageName": f.goPackage,
		"TableName":   ToCamelCase(*f.config.Table, true),
		"FuncName":    f.funcName,
		"Routes":      routes,
	})
	if err != nil {
		return err
	}
	if _, err := file.Write(execute.Bytes()); err != nil {
		return err
	}
	return nil
}

func (f *FoxRouteSQLGenerator) parserRoute() ([]*FoxRoute, error) {
	res := make([]*FoxRoute, 0)
	protos := f.parser()
	for _, m := range protos {
		pkgName := m.Package.Name
		for _, s := range m.Service {
			for _, r := range s.RPC {
				for _, option := range r.Options {
					var title string
					var has bool
					for _, line := range r.Comment.Lines {
						if len(line) > 0 {
							// 去掉注释中的空格
							title = strings.Trim(line, " ")
						}
					}
					if option.Name == "(fox.migrate.route)" {
						parseBool, _ := strconv.ParseBool(option.Constant.Source)
						if !parseBool {
							continue
						}
						route := &FoxRoute{
							ID:    pkgName + "." + s.Name + "/" + r.Name,
							Title: title,
						}
						res = append(res, route)
						has = true
					}
					if option.Name == "(google.api.http)" && has != true {
						route := &FoxRoute{
							ID:    pkgName + "." + s.Name + "/" + r.Name,
							Title: title,
						}
						res = append(res, route)
					}

				}
			}
		}
	}
	return res, nil
}

func (f *FoxRouteSQLGenerator) parser() []model.Proto {
	var ret []model.Proto
	for _, p := range f.protos {
		var pro model.Proto
		var serviceList model.Services
		proto.Walk(p, proto.WithImport(func(i *proto.Import) {
			pro.Import = append(pro.Import, model.Import{Import: i})
		}),
			proto.WithEnum(func(enum *proto.Enum) {
				f.enums = append(f.enums, enum)
			}),
			proto.WithMessage(func(message *proto.Message) {
				pro.Message = append(pro.Message, model.Message{Message: message})
			}),
			proto.WithPackage(func(p *proto.Package) {
				pro.Package = model.Package{Package: p}
			}),
			proto.WithService(func(service *proto.Service) {
				serv := model.Service{Service: service}
				elements := service.Elements
				for _, el := range elements {
					v, _ := el.(*proto.RPC)
					if v == nil {
						continue
					}
					serv.RPC = append(serv.RPC, &model.RPC{RPC: v})
				}

				serviceList = append(serviceList, serv)
			}),
			proto.WithOption(func(option *proto.Option) {
				if option.Name == "go_package" {
					pro.GoPackage = option.Constant.Source
				}
			}))
		pro.Service = serviceList
		ret = append(ret, pro)
	}
	return ret
}

func readParser(src string) ([]*proto.Proto, error) {
	var res []*proto.Proto
	abs, err := filepath.Abs(src)
	if err != nil {
		return nil, err
	}
	dir, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	for _, entry := range dir {
		if entry.IsDir() {
			protos, err := readParser(filepath.Join(abs, entry.Name()))
			if err != nil {
				return nil, err
			}
			res = append(res, protos...)
		}
		open, err := os.Open(filepath.Join(abs, entry.Name()))
		if err != nil {
			return nil, err
		}
		if path.Ext(entry.Name()) == ".proto" {
			parser := proto.NewParser(open)
			protoData, err := parser.Parse()
			if err != nil {
				return nil, err
			}
			res = append(res, protoData)
		}
	}
	return res, nil
}

func (f *FoxRouteSQLGenerator) getEntPath(src string) string {
	dir, _ := os.ReadDir(src)
	for _, entry := range dir {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), "ent") {
				if f.hasSchema(filepath.Join(src, entry.Name())) {
					return filepath.Join(src, entry.Name())
				}
			}
			entPath := f.getEntPath(filepath.Join(src, entry.Name()))
			if len(entPath) > 0 {
				return entPath
			}
		}
	}
	return ""
}

func (f *FoxRouteSQLGenerator) getEntImport() string {
	entPath := f.getEntPath("")
	if len(entPath) == 0 {
		return ""
	}
	extendPath := strings.TrimPrefix(entPath, f.module.Dir)
	extendPath = strings.ReplaceAll(extendPath, "\\", "/") // 防止windows路径
	return f.module.Path + extendPath
}

func (f *FoxRouteSQLGenerator) hasSchema(path string) bool {
	dir, _ := os.ReadDir(path)
	for _, entry := range dir {
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), "schema") {
				return true
			}
		}
	}
	return false
}

// ToCamelCase 将下划线分隔的字符串转换为驼峰命名。
// 如果首字母大写为 true，则结果的首字母是大写的（CamelCase），否则是小写的（camelCase）。
func ToCamelCase(s string, firstUpper bool) string {
	words := strings.Split(s, "_")
	var builder strings.Builder

	for i, word := range words {
		if len(word) == 0 {
			continue // 跳过空字符串
		}
		if i == 0 && !firstUpper {
			builder.WriteString(strings.ToLower(word))
		} else {
			first := true
			for _, r := range word {
				if first {
					if firstUpper || i != 0 {
						builder.WriteRune(unicode.ToUpper(r))
					} else {
						builder.WriteRune(unicode.ToLower(r))
					}
					first = false
				} else {
					builder.WriteRune(unicode.ToLower(r))
				}
			}
		}
	}

	return builder.String()
}
