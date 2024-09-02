// Package redis
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
package redis

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/go-fox/fox/config"
)

// Config redis config
type Config struct {
	Address    []string `json:"address"`
	ClientName string   `json:"client_name"`
	// Database to be selected after connecting to the server.
	// Only single-node and failover clients.
	DB int `json:"db"`
	// Common options.

	Dialer    func(ctx context.Context, network, addr string) (net.Conn, error)
	OnConnect func(ctx context.Context, cn *redis.Conn) error

	Protocol         int    `json:"protocol"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	SentinelUsername string `json:"sentinel_username"`
	SentinelPassword string `json:"sentinel_password"`

	MaxRetries      int           `json:"max_retries"`
	MinRetryBackoff time.Duration `json:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `json:"max_retry_backoff"`

	DialTimeout           config.Duration `json:"dial_timeout"`
	ReadTimeout           config.Duration `json:"read_timeout"`
	WriteTimeout          config.Duration `json:"write_timeout"`
	ContextTimeoutEnabled bool            `json:"context_timeout_enabled"`

	// PoolFIFO uses FIFO mode for each node connection pool GET/PUT (default LIFO).
	PoolFIFO bool `json:"pool_fifo"`

	PoolSize        int             `json:"pool_size"`
	PoolTimeout     config.Duration `json:"pool_timeout"`
	MinIdleConns    int             `json:"min_idle_conns"`
	MaxIdleConns    int             `json:"max_idle_conns"`
	MaxActiveConns  int             `json:"max_active_conns"`
	ConnMaxIdleTime config.Duration `json:"conn_max_idle_time"`
	ConnMaxLifetime config.Duration `json:"conn_max_lifetime"`

	// TLS config
	TLSConfig *tls.Config
	CertFile  string `json:"cert_file"`
	KeyFile   string `json:"key_file"`

	// Only cluster clients.

	MaxRedirects   int  `json:"max_redirects"`
	ReadOnly       bool `json:"read_only"`
	RouteByLatency bool `json:"route_by_latency"`
	RouteRandomly  bool `json:"route_randomly"`

	// The sentinel master name.
	// Only failover clients.

	MasterName string `json:"master_name"`

	DisableIndentity bool   `json:"disable_indentity"`
	IdentitySuffix   string `json:"identity_suffix"`
}

// DefaultConfig default config
func DefaultConfig() *Config {
	return &Config{
		Address:         []string{"127.0.0.1:6379"},
		ClientName:      "fox",
		DB:              0,
		PoolSize:        10,
		MaxRetries:      5,
		MinIdleConns:    100,
		DialTimeout:     config.Duration{Duration: 3 * time.Second},
		ReadTimeout:     config.Duration{Duration: 3 * time.Second},
		WriteTimeout:    config.Duration{Duration: 3 * time.Second},
		ConnMaxIdleTime: config.Duration{Duration: 3 * time.Second},
	}
}

// WithOption apply option
func (c *Config) WithOption(opts ...Option) *Config {
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Config) toRedisConf() *redis.UniversalOptions {
	return &redis.UniversalOptions{
		Addrs: c.Address,

		// ClientName will execute the `CLIENT SETNAME ClientName` command for each conn.
		ClientName: c.ClientName,

		// Database to be selected after connecting to the server.
		// Only single-node and failover clients.
		DB: c.DB,

		// Common options.

		Dialer:    c.Dialer,
		OnConnect: c.OnConnect,

		Protocol:         c.Protocol,
		Username:         c.Username,
		Password:         c.Password,
		SentinelUsername: c.SentinelUsername,
		SentinelPassword: c.SentinelPassword,

		MaxRetries:      c.MaxRetries,
		MinRetryBackoff: c.MinRetryBackoff,
		MaxRetryBackoff: c.MaxRetryBackoff,

		DialTimeout:           c.DialTimeout.Duration,
		ReadTimeout:           c.ReadTimeout.Duration,
		WriteTimeout:          c.WriteTimeout.Duration,
		ContextTimeoutEnabled: c.ContextTimeoutEnabled,

		// PoolFIFO uses FIFO mode for each node connection pool GET/PUT (default LIFO).
		PoolFIFO: c.PoolFIFO,

		PoolSize:        c.PoolSize,
		PoolTimeout:     c.PoolTimeout.Duration,
		MinIdleConns:    c.MinIdleConns,
		MaxIdleConns:    c.MaxIdleConns,
		MaxActiveConns:  c.MaxActiveConns,
		ConnMaxIdleTime: c.ConnMaxIdleTime.Duration,
		ConnMaxLifetime: c.ConnMaxLifetime.Duration,

		TLSConfig: c.TLSConfig,

		// Only cluster clients.

		MaxRedirects:   c.MaxRedirects,
		ReadOnly:       c.ReadOnly,
		RouteByLatency: c.RouteByLatency,
		RouteRandomly:  c.RouteRandomly,

		// The sentinel master name.
		// Only failover clients.

		MasterName: c.MasterName,

		DisableIndentity: c.DisableIndentity,
		IdentitySuffix:   c.IdentitySuffix,
	}
}

// Build create a redis client with this config
func (c *Config) Build() Client {
	return NewWithConfig(c)
}

// RawConfig scan config key to Config value
func RawConfig(key string) *Config {
	conf := DefaultConfig()
	if err := config.Get(key).Scan(conf); err != nil {
		panic(err)
	}
	return conf
}

// ScanConfig scan config name to Config value
func ScanConfig(names ...string) *Config {
	key := "application.clients.redis"
	if len(names) > 0 {
		key = key + "." + names[0]
	}
	return RawConfig(key)
}
