// Package discovery
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
package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-fox/sugar/util/surl"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"

	"github.com/go-fox/fox/registry"
)

var _ resolver.Resolver = (*discovery)(nil)

type discovery struct {
	w  registry.Watcher
	cc resolver.ClientConn

	ctx    context.Context
	cancel context.CancelFunc

	insecure bool
	debug    bool
}

func (d *discovery) ResolveNow(_ resolver.ResolveNowOptions) {

}

func (d *discovery) Close() {
	d.cancel()
	err := d.w.Stop()
	if err != nil {
		slog.Error(fmt.Sprintf("[resolver] failed to watch top: %s", err))
	}
}

func (d *discovery) watch() {
	for {
		select {
		case <-d.ctx.Done():
			return
		default:
		}
		ins, err := d.w.Next()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			slog.Error(fmt.Sprintf("[resolver] Failed to watch discovery endpoint: %v", err))
			time.Sleep(time.Second)
			continue
		}
		d.update(ins)
	}
}

func (d *discovery) update(ins []*registry.ServiceInstance) {
	var (
		endpoints = make(map[string]struct{})
		filtered  = make([]*registry.ServiceInstance, 0, len(ins))
	)
	for _, in := range ins {
		ept, err := surl.PickURL(in.Endpoints, surl.Scheme("grpc", !d.insecure))
		if err != nil {
			slog.Error(fmt.Sprintf("[resolver] Failed to pick endpoint: %v", err))
			continue
		}
		if ept == "" {
			continue
		}
		// filter redundant endpoints
		if _, ok := endpoints[ept]; ok {
			continue
		}
		filtered = append(filtered, in)
	}

	addrs := make([]resolver.Address, 0, len(filtered))
	for _, in := range filtered {
		ept, _ := surl.PickURL(in.Endpoints, surl.Scheme("grpc", !d.insecure))
		endpoints[ept] = struct{}{}
		addr := resolver.Address{
			ServerName: in.Name,
			Attributes: parseAttributes(in.Metadata).WithValue("rawServiceInstance", in),
			Addr:       ept,
		}
		addrs = append(addrs, addr)
	}
	if len(addrs) == 0 {
		slog.Warn(fmt.Sprintf("[resolver] No endpoints found in discovery endpoints: %v", ins))
		return
	}
	err := d.cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		slog.Error(fmt.Sprintf("[resolver] failed to update state: %v", err))
	}
	if d.debug {
		b, _ := json.Marshal(filtered)
		slog.Debug(fmt.Sprintf("[resolver] discovery endpoints: %s", b))
	}
}

func parseAttributes(md map[string]string) *attributes.Attributes {
	a := new(attributes.Attributes)
	for k, v := range md {
		a = a.WithValue(k, v)
	}
	return a
}
