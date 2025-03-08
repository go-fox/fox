package nacos

import (
	"net/url"
	"testing"
)

func TestName(t *testing.T) {
	configServers := []string{
		"http://127.0.0.1:8848/nacos?grpc=9848",
	}
	env := "dev"
	var opts []Option
	for _, server := range configServers {
		parse, err := url.Parse(server)
		if err != nil {
			panic(err)
		}
		opts = append(opts, WithServer(*parse))
	}
	opts = append(opts, WithTimeoutMs(5000))
	opts = append(opts, WithNamespaceID("de1782ee-c664-4045-a9e9-c3fd089d4d0c"))
	opts = append(opts, WithCacheDir("./cache"))
	opts = append(opts, WithLogDir("./log"))
	opts = append(opts, WithNotLoadCacheAtStart())
	opts = append(opts, WithGroup("meixiaoguan"))
	opts = append(opts, WithDataID(env+".toml"))
	opts = append(opts, WithLogLevel("warn"))
	opts = append(opts, WithUsername("meixiaoguan"))
	opts = append(opts, WithPassword("Liuqin@76624291"))
	source := NewSource(opts...)
	if _, err := source.Load(); err != nil {
		t.Error(err)
		return
	}
}
