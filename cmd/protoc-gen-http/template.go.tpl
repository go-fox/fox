{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}
{{$serverPath := .HttpPath}}
{{$clientPath := .HttpPath}}

{{- range .MethodSets}}
const Operation{{$svrType}}{{.OriginalName}} = "/{{$svrName}}/{{.OriginalName}}"
{{- end}}

type {{.ServiceType}}HTTPServer interface {
{{- range .MethodSets}}
	{{.Comments}}{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}HTTPServer(r {{$serverPath}}.Router, srv {{.ServiceType}}HTTPServer) {
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", _{{$svrType}}_{{.Name}}_{{.Num}}_HTTP_Handler(srv))
	{{- end}}
}

{{range .Methods}}
{{if .Upload }}
{{$FileQualifiedGoIdent:=.FileQualifiedGoIdent}}
func _{{$svrType}}_{{.Name}}_{{.Num}}_HTTP_Handler(srv {{$svrType}}HTTPServer) func(ctx *{{$serverPath}}.Context) error {
	return func(ctx *{{$serverPath}}.Context) error {
		var in {{.Request}}
		form, err := ctx.MultipartForm()
		if err != nil {
			return err
		}{{ range .UploadFields }}
		if fileheaders, ok := form.File["{{.TagName}}"]; ok {
			for _, fileheader := range fileheaders {
				f, err := fileheader.Open()
				if err != nil {
					return err
				}
				filebuf := make([]byte, fileheader.Size)
				_, err = f.Read(filebuf)
				if err != nil {
					return err
				}{{ if .IsList }}
				in.{{.Name}} = append(in.{{.Name}}, &{{$FileQualifiedGoIdent}}{
					Name: fileheader.Filename,
					Size: fileheader.Size,
					Content:filebuf,
				})
				{{else}}
				in.{{.Name}} = &{{$FileQualifiedGoIdent}}{
					Name: fileheader.Filename,
					Size: fileheader.Size,
					Content:filebuf,
				}{{ end }}
			}
		}{{ end }}
		{{$serverPath}}.SetOperation(ctx,Operation{{$svrType}}{{.OriginalName}})
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.{{.Name}}(ctx, req.(*{{.Request}}))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*{{.Reply}})
		return ctx.Result(200, reply{{.ResponseBody}})
	}
}
{{else}}
func _{{$svrType}}_{{.Name}}_{{.Num}}_HTTP_Handler(srv {{$svrType}}HTTPServer) func(ctx *{{$serverPath}}.Context) error {
	return func(ctx *{{$serverPath}}.Context) error {
		var in {{.Request}}
		if err := ctx.ShouldBind(&in{{.Body}}); err != nil {
			return err
		}
		{{$serverPath}}.SetOperation(ctx,Operation{{$svrType}}{{.OriginalName}})
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.{{.Name}}(ctx, req.(*{{.Request}}))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*{{.Reply}})
		return ctx.Result(200, reply{{.ResponseBody}})
	}
}
{{end}}
{{end}}

type {{.ServiceType}}HTTPClient interface {
{{- range .MethodSets}}
	{{.Name}}(ctx context.Context, req *{{.Request}}, opts ...{{$clientPath}}.CallOption) (rsp *{{.Reply}}, err error)
{{- end}}
}

type {{.ServiceType}}HTTPClientImpl struct{
	cc *{{$clientPath}}.Client
}

func New{{.ServiceType}}HTTPClient (client *{{$clientPath}}.Client) {{.ServiceType}}HTTPClient {
	return &{{.ServiceType}}HTTPClientImpl{client}
}

{{range .MethodSets}}
func (c *{{$svrType}}HTTPClientImpl) {{.Name}}(ctx context.Context, in *{{.Request}}, opts ...{{$clientPath}}.CallOption) (*{{.Reply}}, error) {
	var out {{.Reply}}
	path := "{{.Path}}"
	opts = append(opts, {{$clientPath}}.WithCallOperation(Operation{{$svrType}}{{.OriginalName}}))
	opts = append(opts, {{$clientPath}}.WithCallPathTemplate(path))
	err := c.cc.Invoke(ctx, "{{.Method}}", path, in{{.Body}}, &out{{.ResponseBody}}, opts...)
	if err != nil {
		return nil, err
	}
	return &out, err
}
{{end}}
