// Package nacos
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
package nacos

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/go-fox/fox/registry"
)

var (
	_ registry.Registry  = (*Registry)(nil)
	_ registry.Discovery = (*Registry)(nil)
)

// Registry is registry impl
type Registry struct {
	cli     naming_client.INamingClient
	prefix  string
	weight  float64
	cluster string
	group   string
	kind    string
}

// New create a nacos registry
func New(opts ...Option) *Registry {
	r := &Registry{
		prefix:  "/fox",
		cluster: "DEFAULT",
		group:   constant.DEFAULT_GROUP,
		weight:  100,
		kind:    "grpc",
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.cli == nil {
		panic("nacos client is nil")
	}
	return r
}

// GetService get service list
func (r *Registry) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	res, err := r.cli.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		GroupName:   r.group,
		HealthyOnly: true,
	})
	if err != nil {
		return nil, err
	}
	items := make([]*registry.ServiceInstance, 0, len(res))
	for _, in := range res {
		kind := r.kind
		if k, ok := in.Metadata["kind"]; ok {
			kind = k
		}
		items = append(items, &registry.ServiceInstance{
			ID:        in.InstanceId,
			Name:      in.ServiceName,
			Version:   in.Metadata["version"],
			Metadata:  in.Metadata,
			State:     registry.State(in.Metadata["state"]),
			Endpoints: []string{fmt.Sprintf("%s://%s:%d", kind, in.Ip, in.Port)},
		})
	}
	return items, nil
}

// Watch creates a watcher according to the service name.
func (r *Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	return newWatcher(ctx, r.cli, serviceName, r.group, r.kind, []string{r.cluster})
}

// Register register a sever
func (r *Registry) Register(_ context.Context, si *registry.ServiceInstance) error {
	if si.Name == "" {
		return registry.ErrServiceInstanceNameEmpty
	}
	for _, endpoint := range si.Endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		var rmd map[string]string
		if si.Metadata == nil {
			rmd = map[string]string{
				"kind":    u.Scheme,
				"version": si.Version,
				"state":   string(si.State),
			}
		} else {
			rmd = make(map[string]string, len(si.Metadata)+3)
			for k, v := range si.Metadata {
				rmd[k] = v
			}
			rmd["kind"] = u.Scheme
			rmd["version"] = si.Version
			rmd["state"] = string(si.State)
		}
		_, e := r.cli.RegisterInstance(vo.RegisterInstanceParam{
			Ip:          host,
			Port:        uint64(p),
			ServiceName: si.Name + "." + u.Scheme,
			Weight:      r.weight,
			Enable:      true,
			Healthy:     true,
			Ephemeral:   true,
			Metadata:    rmd,
			ClusterName: r.cluster,
			GroupName:   r.group,
		})
		if e != nil {
			return fmt.Errorf("RegisterInstance err %v,%v", e, endpoint)
		}
	}
	return nil
}

// Update update sever info
func (r *Registry) Update(_ context.Context, si *registry.ServiceInstance) error {
	for _, endpoint := range si.Endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		var rmd map[string]string
		if si.Metadata == nil {
			rmd = map[string]string{
				"kind":    u.Scheme,
				"version": si.Version,
				"state":   string(si.State),
			}
		} else {
			rmd = make(map[string]string, len(si.Metadata)+3)
			for k, v := range si.Metadata {
				rmd[k] = v
			}
			rmd["kind"] = u.Scheme
			rmd["version"] = si.Version
			rmd["state"] = string(si.State)
		}
		_, e := r.cli.UpdateInstance(vo.UpdateInstanceParam{
			Ip:          host,
			Port:        uint64(p),
			ServiceName: si.Name + "." + u.Scheme,
			Weight:      r.weight,
			Enable:      true,
			Ephemeral:   true,
			Metadata:    rmd,
			ClusterName: r.cluster,
			GroupName:   r.group,
		})
		if e != nil {
			return fmt.Errorf("RegisterInstance err %v,%v", e, endpoint)
		}
	}
	return nil
}

// Deregister deregister server
func (r *Registry) Deregister(_ context.Context, service *registry.ServiceInstance) error {
	for _, endpoint := range service.Endpoints {
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		host, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			return err
		}
		p, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		if _, err = r.cli.DeregisterInstance(vo.DeregisterInstanceParam{
			Ip:          host,
			Port:        uint64(p),
			ServiceName: service.Name + "." + u.Scheme,
			GroupName:   r.group,
			Cluster:     r.cluster,
			Ephemeral:   true,
		}); err != nil {
			return err
		}
	}
	return nil
}
