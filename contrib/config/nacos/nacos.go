package nacos

import (
	"context"
	"github.com/go-fox/fox/config"
	configClient "github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	nacosClient "github.com/nacos-group/nacos-sdk-go/v2/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"path/filepath"
	"strings"
	"time"
)

type nacos struct {
	opts   *vo.ConfigParam
	client configClient.IConfigClient
}

// NewSource  构造函数
func NewSource(nacosClient nacosClient.INacosClient, opts ...Option) config.Source {
	client, err := configClient.NewConfigClient(
		nacosClient,
	)
	if err != nil {
		panic(err)
	}
	options := &vo.ConfigParam{}
	for _, opt := range opts {
		opt(options)
	}
	return &nacos{
		opts:   options,
		client: client,
	}
}

// Load 加载数据
func (n *nacos) Load() ([]*config.DataSet, error) {
	content, err := n.client.GetConfig(vo.ConfigParam{
		DataId: n.opts.DataId,
		Group:  n.opts.Group,
	})
	if err != nil {
		return nil, err
	}
	k := n.opts.DataId
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
	watcher := newWatcher(context.Background(), n.opts.DataId, n.opts.Group, n.client.CancelListenConfig)
	err := n.client.ListenConfig(vo.ConfigParam{
		DataId: n.opts.DataId,
		Group:  n.opts.Group,
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
