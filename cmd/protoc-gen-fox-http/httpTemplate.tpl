{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

{{- range .MethodSets}}
const Operation{{$svrType}}{{.OriginalName}} = "/{{$svrName}}/{{.OriginalName}}"
{{- end}}

type {{.ServiceType}}HTTPServer interface {
{{- range .MethodSets}}
	{{- if ne .Comment ""}}
	{{.Comment}}
	{{- end}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{- end}}
}

func Register{{.ServiceType}}HTTPServer(s http.Router, srv {{.ServiceType}}HTTPServer) {
	r := s.Group("/")
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv))
	{{- end}}
}

{{range .Methods}}
{{if .Upload }}
{{$FileQualifiedGoIdent:=.FileQualifiedGoIdent}}
{{$FileQualifiedGoIdent:=.FileQualifiedGoIdent}}
func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv {{$svrType}}HTTPServer) func(ctx *http.Context) error {
    return func(ctx *http.Context) error {
        var in {{.Request}}
        form, err := ctx.MultipartForm()
        if err != nil {
            return err
        }{{ range .UploadFields }}
        if fileheaders, ok := form.File["{{.JSONName}}"]; ok {
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
                in.{{.GoName}} = append(in.{{.Name}}, &{{$FileQualifiedGoIdent}}{
                    Name: fileheader.Filename,
                    Size: fileheader.Size,
                    Content:filebuf,
                })
                {{else}}
                in.{{.GoName}} = &{{$FileQualifiedGoIdent}}{
                    Name: fileheader.Filename,
                    Size: fileheader.Size,
                    Content:filebuf,
                }{{ end }}
            }
        }{{ end }}
        http.SetOperation(ctx,Operation{{$svrType}}{{.OriginalName}})
        h := ctx.Middleware(func(c context.Context, req interface{}) (interface{}, error) {
            defer ctx.SetContext(c)
            return srv.{{.Name}}(ctx, req.(*{{.Request}}))
        })
        out, err := h(ctx.Context(), &in)
        if err != nil {
            return err
        }
        reply := out.(*{{.Reply}})
        return ctx.Result(200, reply{{.ResponseBody}})
    }
}
{{else}}
func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv {{$svrType}}HTTPServer) func(ctx *http.Context) error {
	return func(ctx *http.Context) error {
		var in {{.Request}}
		{{- if .HasBody}}
		if err := ctx.BindBody(&in{{.Body}}); err != nil {
			return err
		}
		{{- end}}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		{{- if .HasVars}}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		{{- end}}
		http.SetOperation(ctx,Operation{{$svrType}}{{.OriginalName}})
		h := ctx.Middleware(func(c context.Context, req interface{}) (interface{}, error) {
            defer ctx.SetContext(c)
            return srv.{{.Name}}(ctx, req.(*{{.Request}}))
        })
        out, err := h(ctx.Context(), &in)
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
	{{.Name}}(ctx context.Context, req *{{.Request}}, opts ...http.CallOption) (rsp *{{.Reply}}, err error)
{{- end}}
}

type {{.ServiceType}}HTTPClientImpl struct{
	cc *http.Client
}

func New{{.ServiceType}}HTTPClient (client *http.Client) {{.ServiceType}}HTTPClient {
	return &{{.ServiceType}}HTTPClientImpl{client}
}

{{range .MethodSets}}
func (c *{{$svrType}}HTTPClientImpl) {{.Name}}(ctx context.Context, in *{{.Request}}, opts ...http.CallOption) (*{{.Reply}}, error) {
	var out {{.Reply}}
	pattern := "{{.Path}}"
    path := binding.EncodeURL(pattern, in, {{not .HasBody}})
	opts = append(opts, http.WithCallOperation(Operation{{$svrType}}{{.OriginalName}}))
    opts = append(opts, http.WithCallPathTemplate(path))
	{{if .HasBody -}}
	err := c.cc.Invoke(ctx, "{{.Method}}", path, in{{.Body}}, &out{{.ResponseBody}}, opts...)
	{{else -}}
	err := c.cc.Invoke(ctx, "{{.Method}}", path, nil, &out{{.ResponseBody}}, opts...)
	{{end -}}
	if err != nil {
		return nil, err
	}
	return &out, nil
}
{{end}}
