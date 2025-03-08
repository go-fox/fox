package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

type Option func(o *vo.ConfigParam)

// WithDataID 设置数据ID
func WithDataID(dataID string) Option {
	return func(o *vo.ConfigParam) {
		o.DataId = dataID
	}
}

// WithGroup 设置分组
func WithGroup(group string) Option {
	return func(o *vo.ConfigParam) {
		o.Group = group
	}
}
