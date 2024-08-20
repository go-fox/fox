// Package etcd
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
package etcd

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-fox/sugar/container/smap"
	"go.etcd.io/etcd/client/v3"

	"github.com/go-fox/fox/codec/json"

	"github.com/go-fox/fox/registry"
)

var (
	_        registry.Registry  = (*Registry)(nil)
	_        registry.Discovery = (*Registry)(nil)
	encoding                    = json.Codec{}
)

// Registry is registry impl
type Registry struct {
	ctx       context.Context
	kv        clientv3.KV
	client    *clientv3.Client
	lease     clientv3.Lease
	config    Config
	cancelMap *smap.Map[string, context.CancelFunc]
}

// New creating Registry with options
func New(opts ...Option) *Registry {
	conf := DefaultConfig()
	for _, opt := range opts {
		opt(conf)
	}
	return NewWithConfig(conf)
}

// NewWithConfig creating Registry with config
func NewWithConfig(configs ...*Config) *Registry {
	conf := DefaultConfig()
	if len(configs) > 0 {
		conf = configs[0]
	}
	r := &Registry{
		ctx:       context.Background(),
		config:    *conf,
		cancelMap: smap.New[string, context.CancelFunc](true),
	}

	if r.client == nil {
		panic("etcd client is nil")
	}
	r.kv = clientv3.NewKV(r.client)
	return r
}

// GetService get service list
func (r *Registry) GetService(ctx context.Context, serviceName string) ([]*registry.ServiceInstance, error) {
	key := fmt.Sprintf("%s/%s", r.config.Prefix, serviceName)
	resp, err := r.kv.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	items := make([]*registry.ServiceInstance, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		si := &registry.ServiceInstance{}
		err := encoding.Unmarshal(kv.Value, si)
		if err != nil {
			return nil, err
		}
		if si.Name != serviceName {
			continue
		}
		items = append(items, si)
	}
	return items, nil
}

// Watch creates a watcher according to the service name.
func (r *Registry) Watch(ctx context.Context, serviceName string) (registry.Watcher, error) {
	key := fmt.Sprintf("%s/%s", r.config.Prefix, serviceName)
	return newWatcher(ctx, key, serviceName, r.client)
}

// Register register a sever
func (r *Registry) Register(ctx context.Context, service *registry.ServiceInstance) error {
	key := fmt.Sprintf("%s/%s/%s", r.config.Prefix, service.Name, service.ID)
	value, err := encoding.Marshal(service)
	if err != nil {
		return err
	}
	if r.lease != nil {
		_ = r.lease.Close()
	}
	r.lease = clientv3.NewLease(r.client)
	leaseID, err := r.register(ctx, key, value)
	if err != nil {
		return err
	}

	hctx, cancel := context.WithCancel(r.ctx)
	r.cancelMap.Set(service.ID, cancel)
	go r.heartBeat(hctx, leaseID, key, value)
	return nil
}

// Update update sever info
func (r *Registry) Update(ctx context.Context, service *registry.ServiceInstance) error {
	key := fmt.Sprintf("%s/%s/%s", r.config.Prefix, service.Name, service.ID)
	value, err := encoding.Marshal(service)
	if err != nil {
		return err
	}
	_, err = r.client.Put(ctx, key, string(value))
	if err != nil {
		return err
	}
	return nil
}

// Deregister deregister server
func (r *Registry) Deregister(ctx context.Context, service *registry.ServiceInstance) error {
	defer func() {
		if r.lease != nil {
			_ = r.lease.Close()
		}
	}()
	// cancel heartbeat
	if cancel, ok := r.cancelMap.Get(service.ID); ok {
		cancel()
		r.cancelMap.Del(service.ID)
	}
	key := fmt.Sprintf("%s/%s/%s", r.config.Prefix, service.Name, service.ID)
	_, err := r.client.Delete(ctx, key)
	return err
}

func (r *Registry) register(ctx context.Context, key string, value []byte) (clientv3.LeaseID, error) {
	grant, err := r.lease.Grant(ctx, int64(r.config.TTL.Seconds()))
	if err != nil {
		return 0, err
	}
	_, err = r.client.Put(ctx, key, string(value), clientv3.WithLease(grant.ID))
	if err != nil {
		return 0, err
	}
	return grant.ID, nil
}

func (r *Registry) heartBeat(ctx context.Context, leaseID clientv3.LeaseID, key string, value []byte) {
	curLeaseID := leaseID
	kac, err := r.client.KeepAlive(ctx, leaseID)
	if err != nil {
		curLeaseID = 0
	}
	for {
		if curLeaseID == 0 {
			// try to registerWithKV
			var retreat []int
			for retryCnt := 0; retryCnt < r.config.MaxRetry; retryCnt++ {
				if ctx.Err() != nil {
					return
				}
				// prevent infinite blocking
				idChan := make(chan clientv3.LeaseID, 1)
				errChan := make(chan error, 1)
				cancelCtx, cancel := context.WithCancel(ctx)
				go func() {
					defer cancel()
					id, registerErr := r.register(cancelCtx, key, value)
					if registerErr != nil {
						errChan <- registerErr
					} else {
						idChan <- id
					}
				}()

				select {
				case <-time.After(3 * time.Second):
					cancel()
					continue
				case <-errChan:
					continue
				case curLeaseID = <-idChan:
				}

				kac, err = r.client.KeepAlive(ctx, curLeaseID)
				if err == nil {
					break
				}
				retreat = append(retreat, 1<<retryCnt)
				time.Sleep(time.Duration(retreat[rand.Intn(len(retreat))]) * time.Second)
			}
			if _, ok := <-kac; !ok {
				// retry failed
				return
			}
		}

		select {
		case _, ok := <-kac:
			if !ok {
				if ctx.Err() != nil {
					// channel closed due to context cancel
					return
				}
				// need to retry registration
				curLeaseID = 0
				continue
			}
		case <-r.ctx.Done():
			return
		}
	}
}
