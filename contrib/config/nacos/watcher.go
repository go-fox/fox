package nacos

import (
	"context"
	"github.com/go-fox/fox/config"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"path/filepath"
	"strings"
	"time"
)

type watcher struct {
	dataID             string
	group              string
	content            chan string
	cancelListenConfig cancelListenConfigFunc

	ctx    context.Context
	cancel context.CancelFunc
}
type cancelListenConfigFunc func(params vo.ConfigParam) (err error)

func newWatcher(ctx context.Context, dataID, group string, cancelListen cancelListenConfigFunc) *watcher {
	ctx, cancel := context.WithCancel(ctx)
	return &watcher{
		dataID:             dataID,
		group:              group,
		cancelListenConfig: cancelListen,
		content:            make(chan string, 100),

		ctx:    ctx,
		cancel: cancel,
	}
}

func (w *watcher) Next() ([]*config.DataSet, error) {
	select {
	case <-w.ctx.Done():
		return nil, w.ctx.Err()
	case content := <-w.content:
		k := w.dataID
		return []*config.DataSet{
			{
				Key:       k,
				Value:     []byte(content),
				Format:    strings.TrimPrefix(filepath.Ext(k), "."),
				Timestamp: time.Now(),
			},
		}, nil
	}
}

// Close 关闭
func (w *watcher) Close() error {
	err := w.cancelListenConfig(vo.ConfigParam{
		DataId: w.dataID,
		Group:  w.group,
	})
	w.cancel()
	return err
}

// Stop 停止
func (w *watcher) Stop() error {
	return w.Close()
}
