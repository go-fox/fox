package websocket

import (
	"context"
	"errors"
	"testing"

	v1 "github.com/go-fox/fox/internal/testdata/helloword/v1"
)

func TestServer(t *testing.T) {
	server := NewServer(func(s *Server) {
		s.address = "127.0.0.1:8989"
	})
	server.Handler("test", func(ctx Context) error {
		req := &v1.SayHiRequest{}
		if err := ctx.Bind(req); err != nil {
			return err
		}
		return ctx.Result(&v1.SayHiResponse{
			Content: "服务端回复：" + req.Name,
		})
	})
	server.Handler("test2", func(ctx Context) error {
		return errors.New("错误信息")
	})
	err := server.Start(context.Background())
	if err != nil {
		t.Error(err)
	}
}
