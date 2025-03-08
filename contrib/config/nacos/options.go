package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"net/url"
	"strconv"
)

type options struct {
	clientConfig  constant.ClientConfig
	serverConfigs []constant.ServerConfig
	configParam   vo.ConfigParam
}

// Option 参数
type Option func(o *options)

// WithServer 设置nacos服务地址
func WithServer(url url.URL) Option {
	portStr := url.Port()
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 80
	}
	host := url.Hostname()
	grpcPort := uint64(9848)
	grpcPortStr := url.Query().Get("grpc")
	if grpcPortStr != "" {
		grpcPort, err = strconv.ParseUint(grpcPortStr, 10, 64)
		if err != nil {
			grpcPort = 9848
		}
	}
	return func(o *options) {
		o.serverConfigs = append(o.serverConfigs, constant.ServerConfig{
			Scheme:      url.Scheme,
			ContextPath: url.Path,
			Port:        uint64(port),
			IpAddr:      host,
			GrpcPort:    grpcPort,
		})
	}
}

// WithNamespaceID 设置namespace
func WithNamespaceID(namespaceID string) Option {
	return func(o *options) {
		o.clientConfig.NamespaceId = namespaceID
	}
}

// WithTimeoutMs 设置超时时间
func WithTimeoutMs(timeout uint64) Option {
	return func(o *options) {
		o.clientConfig.TimeoutMs = timeout
	}
}

// WithNotLoadCacheAtStart 启动时不加载缓存
func WithNotLoadCacheAtStart() Option {
	return func(o *options) {
		o.clientConfig.NotLoadCacheAtStart = true
	}
}

// WithLogDir 日志目录
func WithLogDir(dir string) Option {
	return func(o *options) {
		o.clientConfig.LogDir = dir
	}
}

// WithCacheDir 缓存目录
func WithCacheDir(dir string) Option {
	return func(o *options) {
		o.clientConfig.CacheDir = dir
	}
}

// WithLogLevel 设置日志等级
func WithLogLevel(dir string) Option {
	return func(o *options) {
		o.clientConfig.LogLevel = dir
	}
}

// WithGroup 设置分组
func WithGroup(group string) Option {
	return func(o *options) {
		o.configParam.Group = group
	}
}

// WithDataID 设置配置ID
func WithDataID(dataID string) Option {
	return func(o *options) {
		o.configParam.DataId = dataID
	}
}

// WithUsername 设置用户名
func WithUsername(username string) Option {
	return func(o *options) {
		o.clientConfig.Username = username
	}
}

// WithPassword 设置密码
func WithPassword(password string) Option {
	return func(o *options) {
		o.clientConfig.Password = password
	}
}
