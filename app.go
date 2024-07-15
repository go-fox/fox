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
	"runtime"
	"sync"
	"time"

	"github.com/fatih/color"

	"go.uber.org/automaxprocs/maxprocs"

	"github.com/go-fox/fox/internal/cycle"
	"github.com/go-fox/fox/internal/signals"
	"github.com/go-fox/fox/registry"
	"github.com/go-fox/fox/transport"
)

// AppInfo 应用信息
type AppInfo interface {
	ID() string
	Name() string
	Version() string
	Metadata() map[string]string
	Endpoint() []string
}

// Application application
type Application struct {
	ctx           context.Context           // 应用上下文
	options       *options                  // 应用配置信息
	locker        *sync.RWMutex             // 读写锁
	cycle         *cycle.Cycle              // 生命周期管理
	startupOnce   sync.Once                 // 启动执行函数
	stopOnce      sync.Once                 // 停止执行函数
	serviceInfo   *registry.ServiceInstance // 服务信息
	maxprocsClean func()                    // max procs clean
	logger        *slog.Logger              // 日志
}

// New create application
func New(opts ...Option) *Application {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}
	return &Application{
		options: options,
		ctx:     options.ctx,
		cycle:   cycle.NewCycle(),
		locker:  &sync.RWMutex{},
		logger:  options.logger,
	}
}

// ID app id
func (app *Application) ID() string {
	return app.serviceInfo.ID
}

// Name app name
func (app *Application) Name() string {
	return app.serviceInfo.Name
}

// Version app version
func (app *Application) Version() string {
	return app.serviceInfo.Version
}

// Metadata app metadata
func (app *Application) Metadata() map[string]string {
	return app.serviceInfo.Metadata
}

// Endpoint registry endpoint
func (app *Application) Endpoint() []string {
	return app.serviceInfo.Endpoints
}

// initialize 初始化
func (app *Application) startup() (err error) {
	app.startupOnce.Do(func() {
		err = app.serialRunner(
			app.printBanner,
			app.initMaxProcs,
		)
	})
	return
}

// printBanner 打印banner
func (app *Application) printBanner() error {
	if app.options.hideBanner {
		return nil
	}
	const banner = `
				(  __)/  \( \/ )
				 ) _)(  O ))  ( 
				(__)  \__/(_/\_)
			
			fox@` + version + `    https://github.com/go-fox/fox


`
	color.Green(banner)
	return nil
}

// InitMaxProcs 设置处理器调用
func (app *Application) initMaxProcs() error {
	if maxProcs := app.options.maxProc; maxProcs != 0 {
		runtime.GOMAXPROCS(int(maxProcs))
	} else {
		if clean, err := maxprocs.Set(); err != nil {
			app.logger.With(slog.Any("error", err)).Error("max procs failed")
		} else {
			app.maxprocsClean = clean
		}
	}
	return nil
}

// serialRunner 串行执行器
func (app *Application) serialRunner(fns ...func() error) error {
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

// Run 启动
func (app *Application) Run() error {
	// 启动前初始化
	if err := app.startup(); err != nil {
		return err
	}
	app.waitSignals()
	defer app.clear()

	// 启动服务
	app.cycle.Run(app.startServers)

	if err := <-app.cycle.Wait(); err != nil {
		return err
	}
	app.logger.Info("shutdown ceres, bye!")
	return nil
}

// waitSignals 等待退出命令
func (app *Application) waitSignals() {
	signals.Shutdown(func(grace bool) {
		_ = app.Stop()
	})
}

// startServers 启动服务
func (app *Application) startServers() error {
	info, err := app.buildServerInfo()
	if err != nil {
		return err
	}
	app.locker.Lock()
	app.serviceInfo = info
	app.locker.Unlock()

	appCtx := NewContext(app.ctx, app)
	wg := sync.WaitGroup{}
	// 启动前钩子
	app.runHook(BeforeStart)
	for _, srv := range app.options.servers {
		srv := srv
		wg.Add(1)
		app.cycle.Run(func() error {
			wg.Done()
			return srv.Start(appCtx)
		})
	}
	wg.Wait()

	// 注册服务
	if app.options.registry != nil {
		ctx, cancel := context.WithTimeout(appCtx, 3*time.Second)
		defer cancel()
		if err = app.options.registry.Register(ctx, info); err != nil {
			return err
		}
	}

	// 启动后钩子
	app.runHook(AfterStart)

	return nil
}

// buildServerInfo builder service instance
func (app *Application) buildServerInfo() (*registry.ServiceInstance, error) {
	endpoints := make([]string, 0)
	if len(endpoints) == 0 {
		for _, srv := range app.options.servers {
			if r, ok := srv.(transport.Endpointer); ok {
				e, err := r.Endpoint()
				if err != nil {
					return nil, err
				}
				endpoints = append(endpoints, e.String())
			}
		}
	}
	info := &registry.ServiceInstance{
		ID:      app.options.id,
		Name:    app.options.name,
		Version: app.options.version,
		Metadata: map[string]string{
			"region": app.options.region,
			"zone":   app.options.zone,
		},
		Endpoints: endpoints,
	}
	for key, value := range app.options.metadata {
		info.Metadata[key] = value
	}
	return info, nil
}

// Stop 停止应用
func (app *Application) Stop() (err error) {
	app.stopOnce.Do(func() {
		// 执行钩子
		app.runHook(BeforeStop)
		// 服务信息
		serverInfo := app.serviceInfo
		// 注销服务
		stopCtx, cancel := context.WithTimeout(NewContext(app.ctx, app), app.options.stopTimeout)
		defer cancel()
		if app.options.registry != nil && serverInfo != nil {
			if err = app.options.registry.Deregister(stopCtx, serverInfo); err != nil {
				app.logger.With(slog.Any("error", err)).Error("stop server error")
			}
		}
		// 停止服务
		app.locker.RLock()
		for _, s := range app.options.servers {
			s := s
			app.cycle.Run(func() error {
				return s.Stop(stopCtx)
			})
		}
		app.locker.RUnlock()
		<-app.cycle.Done()
		app.runHook(AfterStop)
		app.cycle.Close()
	})
	return err
}

// clear 清除
func (app *Application) clear() {
	app.maxprocsClean()
}

// runHook 运行钩子
func (app *Application) runHook(k HookType) {
	hooks, ok := app.options.hooks[k]
	if ok {
		ctx := NewContext(app.ctx, app)
		for _, hook := range hooks {
			hook(ctx)
		}
	}
}

type appKey struct{}

// NewContext 创建附带服务信息的上下文
func NewContext(ctx context.Context, info AppInfo) context.Context {
	return context.WithValue(ctx, appKey{}, info)
}

// FromContext 从上下文中获取服务信息
func FromContext(ctx context.Context) (info AppInfo, ok bool) {
	info, ok = ctx.Value(appKey{}).(AppInfo)
	return
}
