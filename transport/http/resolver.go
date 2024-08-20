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

package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/go-fox/sugar/util/surl"

	"github.com/go-fox/fox/internal/cycle"
	"github.com/go-fox/fox/registry"
	"github.com/go-fox/fox/selector"
	"github.com/go-fox/fox/selector/base"
)

// Target endpoint parse data
type Target struct {
	Scheme    string
	Authority string
	Endpoint  string
}

// parseTarget parse endpoint
func parseTarget(endpoint string, insecure bool) (*Target, error) {
	if !strings.Contains(endpoint, "://") {
		if insecure {
			endpoint = "http://" + endpoint
		} else {
			endpoint = "https://" + endpoint
		}
	}
	parse, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	target := &Target{
		Scheme:    parse.Scheme,
		Authority: parse.Host,
	}
	if len(parse.Path) > 1 {
		target.Endpoint = parse.Path[1:]
	}
	return target, nil
}

// resolver 服务发现客户端
type resolver struct {
	discovery registry.Discovery //
	selector  selector.Selector  // 节点平衡器
	target    *Target            // 目标地址
	watcher   registry.Watcher   // 服务监听者
	insecure  bool               // 是否是不安全的
	logger    *slog.Logger       // 日志
	cycle     *cycle.Cycle
}

// newResolver create resolver
func newResolver(ctx context.Context, log *slog.Logger, discovery registry.Discovery, target *Target, selector selector.Selector, block, insecure bool) (*resolver, error) {
	watcher, err := discovery.Watch(ctx, target.Endpoint)
	if err != nil {
		return nil, err
	}
	r := &resolver{
		logger:    log,
		target:    target,
		watcher:   watcher,
		selector:  selector,
		insecure:  insecure,
		discovery: discovery,
		cycle:     cycle.NewCycle(),
	}
	if block {
		r.cycle.Run(func() error {
			for {
				services, err := watcher.Next()
				if err != nil {
					return err
				}
				if r.update(services) {
					return nil
				}
			}
		})
		select {
		case err := <-r.cycle.Wait():
			if err != nil {
				stopErr := watcher.Stop()
				if stopErr != nil {
					log.Error(fmt.Sprintf("failed to http client watch stop: %v, error: %+v", target, stopErr))
				}
				return nil, err
			}
		case <-ctx.Done():
			log.Error(fmt.Sprintf("http client watch service %v reaching context deadline!", target))
			stopErr := watcher.Stop()
			if stopErr != nil {
				log.Error(fmt.Sprintf("failed to http client watch stop: %v, error: %+v", target, stopErr))
			}
			return nil, ctx.Err()
		}

	}
	go func() {
		_ = r.run()
	}()
	return r, nil
}

// run watcher node
func (r *resolver) run() error {
	for {
		services, err := r.watcher.Next()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			r.logger.Error(fmt.Sprintf("http client watch service %v got unexpected error:=%v", r.target, err))
			time.Sleep(time.Second)
			continue
		}
		r.update(services)
	}
}

// update update node
func (r *resolver) update(services []*registry.ServiceInstance) bool {
	nodes := make([]selector.Node, 0)
	for _, ins := range services {
		ept, err := surl.PickURL(ins.Endpoints, surl.Scheme("http", !r.insecure))
		if err != nil {
			r.logger.Error(fmt.Sprintf("Failed to parse (%v) discovery endpoint: %v error %v", r.target, ins.Endpoints, err))
			continue
		}
		if ept == "" {
			continue
		}
		nodes = append(nodes, base.NewNode("http", ept, ins))
	}
	if len(nodes) == 0 {
		r.logger.Warn(fmt.Sprintf("[http resolver]Zero endpoint found,refused to write,set: %s ins: %v", r.target.Endpoint, nodes))
		return false
	}
	r.selector.Store(nodes)
	return true
}

// Close is stop watcher
func (r *resolver) Close() error {
	return r.watcher.Stop()
}
