package registry

import (
	"context"
	"errors"
)

const (
	Disallow State = "disallow" // Disallow is service disallow request
	Down     State = "down"     // Down is service stop
	Up       State = "up"       // Up is service start
)

// ErrServiceInstanceNameEmpty no instance name
var ErrServiceInstanceNameEmpty = errors.New("service instance name is empty")

// Registry 注册中心
type Registry interface {
	// Register 注册服务
	//
	//  @param ctx context.Context
	//  @param service *ServiceInstance 服务信息
	//  @return error
	Register(ctx context.Context, service *ServiceInstance) error
	// Update 修改服务
	//
	//  @param ctx context.Context
	//  @param service *ServiceInstance
	//  @return error
	Update(ctx context.Context, service *ServiceInstance) error
	// Deregister 注销服务
	//
	//  @param ctx context.Context
	//  @param service *ServiceInstance 服务信息
	//  @return error
	Deregister(ctx context.Context, service *ServiceInstance) error
}

// Discovery 服务发现
type Discovery interface {
	// GetService 获取服务器列表
	//
	//  @param ctx context.Context
	//  @param serviceName string
	//  @return []*ServiceInstance
	//  @return error
	GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	// Watch 监听服务
	//
	//  @param ctx context.Context
	//  @param serviceName string
	//  @return Watcher
	//  @return error
	Watch(ctx context.Context, serviceName string) (Watcher, error)
}

// Watcher is service watcher.
type Watcher interface {
	// Next returns services in the following two cases:
	//	1.第一次获取服务列表部位空.
	//	2.发现任何服务实例发生更改.
	//	否则会一直阻塞
	//
	//  @return []*ServiceInstance
	//  @return error
	Next() ([]*ServiceInstance, error)
	// Stop 停止监控
	//
	//  @return error
	Stop() error
}

// State is a service state type
type State string

// ServiceInstance 服务信息
type ServiceInstance struct {
	// ID 当前服务的唯一ID
	ID string `json:"id"`
	// Name 当前服务注册的唯一ID.
	Name string `json:"name"`
	// Version is the version of the compiled.
	Version string `json:"version"`
	// State is the state of the service instance
	State State `json:"state"`
	// Metadata is the kv pair metadata associated with the service instance.
	Metadata map[string]string `json:"metadata"`
	// Endpoints are endpoint addresses of the service instance.
	// Schema:
	//   http://127.0.0.1:8000?isSecure=false
	//   grpc://127.0.0.1:9000?isSecure=false
	//   ws://127.0.0.1:7000?isSecure=false
	//   tcp://127.0.0.1:6000?isSecure=false
	Endpoints []string `json:"endpoints"`
}

// ServiceInstanceList 服务实例列表
type ServiceInstanceList []*ServiceInstance

// Find 查找服务
//
//	@receiver s
//	@param id string
//	@return *ServiceInstance
func (s ServiceInstanceList) Find(id string) *ServiceInstance {
	for _, instance := range s {
		if instance.ID == id {
			return instance
		}
	}
	return nil
}
