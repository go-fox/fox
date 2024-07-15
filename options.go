// Package fox
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
package fox

import (
	"context"
	"log/slog"
	"net/url"
	"time"

	"github.com/go-fox/fox/registry"
	"github.com/go-fox/fox/transport"
)

// HookType 定义 hook类型
type HookType uint

// HookFunc 钩子回调方法
type HookFunc func(ctx context.Context)

const (
	ModName              = "app" // ModName mod name
	BeforeStart HookType = iota  // BeforeStart hook enum
	BeforeStop                   // BeforeStop hook enum
	AfterStart                   // AfterStart hook enum
	AfterStop                    // AfterStop hook enum
)

// Option app option
type Option func(o *options)

// Options options
type options struct {
	ctx              context.Context         // 应用上下文
	id               string                  // 应用唯一标识
	name             string                  // 应用名称
	version          string                  // 应用版本
	metadata         map[string]string       // 附加信息
	endpoints        []*url.URL              // 服务地址
	region           string                  // 服务所属地域
	zone             string                  // 服务所属分区
	hideBanner       bool                    // 隐藏打印横幅
	maxProc          int64                   // 处理器内核优化
	registrarTimeout time.Duration           // 服务注册超时时间
	stopTimeout      time.Duration           // 注销服务超时时间
	hooks            map[HookType][]HookFunc // 启动钩子
	servers          []transport.Server      // 服务集合
	registry         registry.Registry       // 注册中心
	logger           *slog.Logger            // 日志组件
}

// defaultOptions default options
func defaultOptions() *options {
	return &options{
		ctx:              context.Background(),
		id:               AppId(),
		name:             AppName(),
		version:          AppVersion(),
		region:           AppRegion(),
		zone:             AppZone(),
		hideBanner:       false,
		registrarTimeout: 10 * time.Second,
		stopTimeout:      3 * time.Second,
		hooks:            map[HookType][]HookFunc{},
		servers:          []transport.Server{},
		logger:           slog.With(slog.String("mod", ModName)),
	}
}

// Context app context
func Context(ctx context.Context) Option {
	return func(o *options) {
		o.ctx = ctx
	}
}

// Id app id
func Id(id string) Option {
	return func(o *options) {
		o.id = id
	}
}

// Name app name
func Name(name string) Option {
	return func(o *options) {
		o.name = name
	}
}

// Version with an app version
func Version(version string) Option {
	return func(o *options) {
		o.version = version
	}
}

// Metadata with app metadata
func Metadata(metadata map[string]string) Option {
	return func(o *options) {
		o.metadata = metadata
	}
}

// Region with app Region
func Region(region string) Option {
	return func(o *options) {
		o.region = region
	}
}

// Zone with app zone
func Zone(zone string) Option {
	return func(o *options) {
		o.zone = zone
	}
}

// HideBanner hide print banner
func HideBanner() Option {
	return func(o *options) {
		o.hideBanner = true
	}
}

// RegistrarTimeout with registry.Registrar timeout
func RegistrarTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.registrarTimeout = timeout
	}
}

// StopTimeout with stop timeout
func StopTimeout(timeout time.Duration) Option {
	return func(o *options) {
		o.stopTimeout = timeout
	}
}

// Hooks with hooks options
func Hooks(hookType HookType, hooks ...HookFunc) Option {
	return func(o *options) {
		if o.hooks[hookType] == nil {
			o.hooks[hookType] = make([]HookFunc, 0)
		}
		o.hooks[hookType] = hooks
	}
}

// AddHooks add hooks
func AddHooks(hookType HookType, hooks ...HookFunc) Option {
	return func(o *options) {
		if o.hooks[hookType] == nil {
			o.hooks[hookType] = make([]HookFunc, 0)
		}
		o.hooks[hookType] = append(o.hooks[hookType], hooks...)
	}
}

// Registry with registry
func Registry(reg registry.Registry) Option {
	return func(o *options) {
		o.registry = reg
	}
}

// Server with server
func Server(servers ...transport.Server) Option {
	return func(o *options) {
		o.servers = servers
	}
}

// Logger with logger
func Logger(log *slog.Logger) Option {
	return func(o *options) {
		o.logger = log
	}
}

// MaxProc with max procs
func MaxProc(maxProc int64) Option {
	return func(o *options) {
		o.maxProc = maxProc
	}
}
