package token

import (
	"log/slog"

	"github.com/go-fox/fox/config"
)

// Config 创建参数
type Config struct {
	LoginType             string                      `json:"login_type"`             // 登录类型
	TokenName             string                      `json:"token_name"`             // token名称
	IsConcurrent          bool                        `json:"is_concurrent"`          // 是否允许同一账号多地同时登录 （为 true 时允许一起登录, 为 false 时新登录挤掉旧登录）
	IsShare               bool                        `json:"is_share"`               // 在多人登录同一账号时，是否共用一个 token （为 true 时所有登录共用一个 token, 为 false 时每次登录新建一个 token）
	Timeout               int64                       `json:"timeout"`                // token 有效期（单位：秒） 默认30天，-1 代表永久有效
	ActiveTimeout         int64                       `json:"active_timeout"`         // token 最低活跃频率（单位：秒），如果 token 超过此时间没有访问系统就会被冻结，默认-1 代表不限制，永不冻结,例如（设置1800秒，则30分钟内无操作就冻结）
	DynamicActiveTimeout  bool                        `json:"dynamic_active_timeout"` // 是否启用动态 ActiveTimeout 功能，如不需要请设置为 false，节省缓存请求次数
	MaxTryCount           int                         `json:"max_try_count"`          // 在每次创建 token 时的最高循环次数，用于保证 token 唯一性（-1=不循环尝试，直接使用）
	MaxLoginCount         int                         `json:"max_login_count"`        // 同一账号最大登录数量，-1代表不限 （只有在 IsConcurrent=true, IsShare=false 时此配置项才有意义）
	Style                 Style                       `json:"style"`                  // token样式
	AutoRenew             bool                        `json:"auto_renew"`             // 是否自动续签
	createTokenFunction   CreateTokenFunction         // 创建token的方法
	generateUniqueToken   GenerateUniqueTokenFunction // 生成唯一token的方法
	createSessionFunction CreateSessionFunction       // 创建session的策略
	listener              *ListenerManager            // 事件监听器
	repository            Repository                  // 存储仓
	logger                *slog.Logger                // 日志实例

}

// DefaultConfig 默认参数
func DefaultConfig() *Config {
	c := &Config{
		LoginType:             "login",
		TokenName:             "authorization",
		IsConcurrent:          true,
		IsShare:               true,
		Timeout:               60 * 60 * 24 * 30,
		AutoRenew:             true,
		ActiveTimeout:         -1,
		MaxTryCount:           3,
		createTokenFunction:   defaultCreateTokenFunction,
		generateUniqueToken:   defaultGenerateUniqueToken,
		createSessionFunction: defaultCreateSessionFunction,
		logger:                slog.With("mod", "auth.token"),
	}
	c.listener = newListenerCenter(c)
	return c
}

// RawConfig scan config
func RawConfig(key string) *Config {
	conf := DefaultConfig()
	if err := config.Get(key).Scan(conf); err != nil {
		panic(err)
	}
	return conf
}

// ScanConfig scan config
func ScanConfig(names ...string) *Config {
	key := "application.auth.token"
	if len(names) > 0 {
		key = key + "." + names[0]
	}
	return RawConfig(key)
}

// WithOptions apply options
func (c *Config) WithOptions(opts ...Option) *Config {
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Build create a token
func (c *Config) Build() Token {
	return NewWithConfig(c)
}
