package nacos

import (
	"context"
	"github.com/go-fox/fox/config"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	configClient "github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/file"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type nacos struct {
	opts   *options
	client configClient.IConfigClient
}

// NewSource  构造函数
func NewSource(opts ...Option) config.Source {
	// 创建clientConfig
	o := &options{
		clientConfig: constant.ClientConfig{
			TimeoutMs:            10 * 1000,
			BeatInterval:         5 * 1000,
			OpenKMS:              false,
			CacheDir:             file.GetCurrentPath() + string(os.PathSeparator) + "cache",
			UpdateThreadNum:      20,
			NotLoadCacheAtStart:  false,
			UpdateCacheWhenEmpty: false,
			LogDir:               file.GetCurrentPath() + string(os.PathSeparator) + "log",
			LogLevel:             "info",
		},
		configParam:   vo.ConfigParam{},
		serverConfigs: make([]constant.ServerConfig, 0),
	}
	for _, opt := range opts {
		opt(o)
	}
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &o.clientConfig,
			ServerConfigs: o.serverConfigs,
		},
	)
	if err != nil {
		panic(err)
	}
	return &nacos{
		opts:   o,
		client: client,
	}
}

// Load 加载数据
func (n *nacos) Load() ([]*config.DataSet, error) {
	content, err := n.client.GetConfig(vo.ConfigParam{
		DataId: n.opts.configParam.DataId,
		Group:  n.opts.configParam.Group,
	})
	if err != nil {
		return nil, err
	}
	k := n.opts.configParam.DataId
	return []*config.DataSet{
		{
			Key:       k,
			Value:     []byte(content),
			Format:    strings.TrimPrefix(filepath.Ext(k), "."),
			Timestamp: time.Now(),
		},
	}, nil
}

// Watch 开启监控
func (n *nacos) Watch() (config.Watcher, error) {
	watcher := newWatcher(context.Background(), n.opts.configParam.DataId, n.opts.configParam.Group, n.client.CancelListenConfig)
	err := n.client.ListenConfig(vo.ConfigParam{
		DataId: n.opts.configParam.DataId,
		Group:  n.opts.configParam.Group,
		OnChange: func(_, group, dataId, data string) {
			if dataId == watcher.dataID && group == watcher.group {
				watcher.content <- data
			}
		},
	})
	if err != nil {
		return nil, err
	}
	return watcher, nil
}
