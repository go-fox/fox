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

// Option create a redis client option
type Option func(c *Config)

// WithAddress with a redis address option
func WithAddress(address []string) Option {
	return func(c *Config) {
		c.Address = address
	}
}

// WithClientName with a redis client name option
func WithClientName(clientName string) Option {
	return func(c *Config) {
		c.ClientName = clientName
	}
}

// WithDB Database to be selected after connecting to the server.
// Only single-node and failover clients.
func WithDB(db int) Option {
	return func(c *Config) {
		c.DB = db
	}
}

// WithDialer with a client Dialer option
func WithDialer(dialer func(ctx context.Context, network, addr string) (net.Conn, error)) Option {
	return func(c *Config) {
		c.Dialer = dialer
	}
}

// WithOnConnect with a client connect callback option
func WithOnConnect(onConnect func(ctx context.Context, cn *redis.Conn) error) Option {
	return func(c *Config) {
		c.OnConnect = onConnect
	}
}

// WithProtocol with a client protocol option
func WithProtocol(protocol int) Option {
	return func(c *Config) {
		c.Protocol = protocol
	}
}

// WithUsername with a redis client username option
func WithUsername(username string) Option {
	return func(c *Config) {
		c.Username = username
	}
}

// WithPassword with a redis password option
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithSentinelUsername with a redis SentinelUsername option
func WithSentinelUsername(sentinelUsername string) Option {
	return func(c *Config) {
		c.SentinelUsername = sentinelUsername
	}
}

// WithSentinelPassword with a redis SentinelPassword option
func WithSentinelPassword(sentinelUsername string) Option {
	return func(c *Config) {
		c.SentinelUsername = sentinelUsername
	}
}

// WithPoolSize with a redis client pool size option
func WithPoolSize(poolSize int) Option {
	return func(c *Config) {
		c.PoolSize = poolSize
	}
}

// WithMaxRetries with a redis client max retries
func WithMaxRetries(maxRetries int) Option {
	return func(c *Config) {
		c.MaxRetries = maxRetries
	}
}

// WithMinRetryBackoff with a redis client MinRetryBackoff option
func WithMinRetryBackoff(minRetryBackoff time.Duration) Option {
	return func(c *Config) {
		c.MinRetryBackoff = minRetryBackoff
	}
}

// WithMaxRetryBackoff with a redis client MaxRetryBackoff option
func WithMaxRetryBackoff(maxRetryBackoff time.Duration) Option {
	return func(c *Config) {
		c.MaxRetryBackoff = maxRetryBackoff
	}
}

// WithDialTimeout with a redis clint dial timeout option
func WithDialTimeout(dialTimeout time.Duration) Option {
	return func(c *Config) {
		c.DialTimeout = config.Duration{
			Duration: dialTimeout,
		}
	}
}

// WithReadTimeout with a redis client read timeout option
func WithReadTimeout(readTimeout time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = config.Duration{
			Duration: readTimeout,
		}
	}
}

// WithWriteTimeout with a redis client writeTimeout option
func WithWriteTimeout(writeTimeout time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = config.Duration{
			Duration: writeTimeout,
		}
	}
}

// WithContextTimeoutEnabled with a context timout enable
func WithContextTimeoutEnabled(contextTimeoutEnabled bool) Option {
	return func(c *Config) {
		c.ContextTimeoutEnabled = contextTimeoutEnabled
	}
}

// WithReadOnly with a redis client readOnly option
func WithReadOnly(readOnly bool) Option {
	return func(c *Config) {
		c.ReadOnly = readOnly
	}
}

// WithMinIdleConns with a redis client minIdleConns
func WithMinIdleConns(minIdleConns int) Option {
	return func(c *Config) {
		c.MinIdleConns = minIdleConns
	}
}

// WithMMaxIdleConns with a redis client MaxIdleConns option
func WithMMaxIdleConns(maxIdleConns int) Option {
	return func(c *Config) {
		c.MaxIdleConns = maxIdleConns
	}
}

// WithMaxActiveConns with a redis client MaxActiveConns option
func WithMaxActiveConns(maxActiveConns int) Option {
	return func(c *Config) {
		c.MaxActiveConns = maxActiveConns
	}
}

// WithConnMaxIdleTime with a redis conn max idle time option
func WithConnMaxIdleTime(connMaxIdleTime time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxIdleTime = config.Duration{
			Duration: connMaxIdleTime,
		}
	}
}

// WithConnMaxLifetime with a redis conn max lifet time option
func WithConnMaxLifetime(connMaxLifetime time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxLifetime = config.Duration{
			Duration: connMaxLifetime,
		}
	}
}

// WithMaxRedirects with a redis client MaxRedirects option
func WithMaxRedirects(maxRedirects int) Option {
	return func(c *Config) {
		c.MaxRedirects = maxRedirects
	}
}

// WithRouteByLatency with a redis client RouteByLatency option
func WithRouteByLatency(routeByLatency bool) Option {
	return func(c *Config) {
		c.RouteByLatency = routeByLatency
	}
}

// WithRouteRandomly with a redis client RouteByLatency option
func WithRouteRandomly(routeRandomly bool) Option {
	return func(c *Config) {
		c.RouteRandomly = routeRandomly
	}
}

// WithMasterName with MasterName option
func WithMasterName(masterName string) Option {
	return func(c *Config) {
		c.MasterName = masterName
	}
}

// WithTLSFile with TLS file path option
func WithTLSFile(certFile, keyFile string) Option {
	return func(c *Config) {
		c.KeyFile = keyFile
		c.CertFile = certFile
	}
}

// WithTLSConfig with tls.Config option
func WithTLSConfig(tLSConfig *tls.Config) Option {
	return func(c *Config) {
		c.TLSConfig = tLSConfig
	}
}
