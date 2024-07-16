package config

import (
	"errors"
	"log/slog"
	"testing"
	"time"
)

const (
	_testJSON = `{
	"username":"admin",
	"password":"123456"
}`
	_testToml = `username="admin"
password="123456"`
)

type testZhCnJSONSource struct {
	sig chan struct{}
	err chan struct{}
}

func (t *testZhCnJSONSource) Watch() (Watcher, error) {
	return newTestWatcher(t.sig, t.err), nil
}

func (t *testZhCnJSONSource) Load() ([]*DataSet, error) {
	return []*DataSet{
		{
			Key:       "redis",
			Value:     []byte(_testJSON),
			Format:    "json",
			Timestamp: time.Now(),
		},
	}, nil
}

type testEnToml struct {
	sig chan struct{}
	err chan struct{}
}

func (t *testEnToml) Watch() (Watcher, error) {
	return newTestWatcher(t.sig, t.err), nil
}

func (t *testEnToml) Load() ([]*DataSet, error) {
	return []*DataSet{
		{
			Key:    "mysql",
			Value:  []byte(_testToml),
			Format: "toml",
		},
	}, nil
}

type testWatcher struct {
	sig  chan struct{}
	err  chan struct{}
	exit chan struct{}
}

func newTestWatcher(sig, err chan struct{}) Watcher {
	return &testWatcher{sig: sig, err: err, exit: make(chan struct{})}
}

func (w *testWatcher) Next() ([]*DataSet, error) {
	select {
	case <-w.sig:
		return nil, nil
	case <-w.err:
		return nil, errors.New("error")
	case <-w.exit:
		return nil, nil
	}
}

func (w *testWatcher) Stop() error {
	close(w.exit)
	return nil
}

func TestConfig(t *testing.T) {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	c := New(
		WithSources(&testZhCnJSONSource{
			sig: make(chan struct{}),
			err: make(chan struct{}),
		}, &testEnToml{
			sig: make(chan struct{}),
			err: make(chan struct{}),
		}),
	)
	err := c.Load()
	if err != nil {
		t.Fatal(err)
	}
	load := c.Get("redis").Load()
	t.Logf("%+v", load)
}
