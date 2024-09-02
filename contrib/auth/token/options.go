// Package token
// MIT License
//
// # Copyright (c) 2024 golang-token
// Author https://github.com/golang-token/token
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
package token

import (
	"log/slog"
)

///  ==============创建参数相关=====================

// Style Token值样式
type Style string

const (
	StyleUUID       Style = "uuid"        // StyleUUID uuid样式
	StyleSimpleUUID Style = "simple-uuid" // StyleSimpleUUID uuid不带下划线
	StyleRandom32   Style = "random-32"   // StyleRandom32 随机32位字符串
	StyleRandom64   Style = "random-64"   // StyleRandom64 随机64位字符串
)

// Option 创建参数
type Option func(o *Config)

// WithLoginType 设置登录类型
//
//	@param loginType string
//	@return Option
func WithLoginType(loginType string) Option {
	return func(o *Config) {
		o.LoginType = loginType
	}
}

// WithTokenName 设置token名称
//
//	@param tokenName string
//	@return Option
func WithTokenName(tokenName string) Option {
	return func(o *Config) {
		o.TokenName = tokenName
	}
}

// WithIsConcurrent 是否允许同一账号多地同时登录 （为 true 时允许一起登录, 为 false 时新登录挤掉旧登录）
//
//	@param isConcurrent bool
//	@return Option
func WithIsConcurrent(isConcurrent bool) Option {
	return func(o *Config) {
		o.IsConcurrent = isConcurrent
	}
}

// WithTimeout 设置此次登录的有效期（单位：秒）
//
//	@param timeout int64
//	@return Option
func WithTimeout(timeout int64) Option {
	return func(o *Config) {
		o.Timeout = timeout
	}
}

// WithActiveTimeout 设置最低活跃时间（单位：秒）
//
//	@param timeout int64
//	@return Option
func WithActiveTimeout(timeout int64) Option {
	return func(o *Config) {
		o.ActiveTimeout = timeout
	}
}

// WithDynamicActiveTimeout 设置是否支持动态活跃时间
//
//	@param dynamicActiveTimeout bool
//	@return Option
func WithDynamicActiveTimeout(dynamicActiveTimeout bool) Option {
	return func(o *Config) {
		o.DynamicActiveTimeout = dynamicActiveTimeout
	}
}

// WithMaxTryCount 设置每次创建 token 时的最高循环次数，用于保证token的唯一性
//
//	@param maxTryCount int
//	@return Option
func WithMaxTryCount(maxTryCount int) Option {
	return func(o *Config) {
		o.MaxTryCount = maxTryCount
	}
}

// WithMaxLoginCount 设置同一个账号最大的登录数量，-1代表不限 （只有在 isConcurrent=true, isShare=false 时此配置项才有意义）
//
//	@param maxLoginCount int
//	@return Option
func WithMaxLoginCount(maxLoginCount int) Option {
	return func(o *Config) {
		o.MaxLoginCount = maxLoginCount
	}
}

// WithStyle 设置token值样式
//
//	@param style Style token值类型
//	@return Option
func WithStyle(style Style) Option {
	return func(o *Config) {
		o.Style = style
	}
}

// WithAutoRenew 设置是否自动更新
//
//	@param autoRenew bool 是否自动更新token
//	@return Option
func WithAutoRenew(autoRenew bool) Option {
	return func(o *Config) {
		o.AutoRenew = autoRenew
	}
}

// WithCreateTokenFunction 设置token创建方法
//
//	@param createTokenFunction CreateTokenFunction 创建token的方法
//	@return Option
func WithCreateTokenFunction(createTokenFunction CreateTokenFunction) Option {
	return func(o *Config) {
		o.createTokenFunction = createTokenFunction
	}
}

// WithGenerateUniqueToken 设置生成唯一token的方法
//
//	@param generateUniqueToken GenerateUniqueTokenFunction
//	@return Option
func WithGenerateUniqueToken(generateUniqueToken GenerateUniqueTokenFunction) Option {
	return func(o *Config) {
		o.generateUniqueToken = generateUniqueToken
	}
}

// WithCreateSessionFunction 设置创建session的策略方法
//
//	@param createSessionFunction CreateTokenFunction
//	@return Option
func WithCreateSessionFunction(createSessionFunction CreateSessionFunction) Option {
	return func(o *Config) {
		o.createSessionFunction = createSessionFunction
	}
}

// WithAppendListener 添加监听器
//
//	@param listener Listener
//	@return Option
func WithAppendListener(listener ...Listener) Option {
	return func(o *Config) {
		o.listener.RegisterListener(listener...)
	}
}

// WithSetListener 重置监听器
//
//	@param listener ...Listener
//	@return Option
func WithSetListener(listener ...Listener) Option {
	return func(o *Config) {
		o.listener.SetListener(listener)
	}
}

// WithRepository 设置数据持久化
//
//	@param repository Repository 数据持久化接口
//	@return Option
func WithRepository(repository Repository) Option {
	return func(o *Config) {
		o.repository = repository
	}
}

// WithLogger 设置日志组件
//
//	@param logger Logger 日志组件实例
//	@return Option
func WithLogger(logger *slog.Logger) Option {
	return func(o *Config) {
		o.logger = logger
	}
}

/// ================登录参数相关====================

// LoginOptions 登录参数
type LoginOptions struct {
	logiId        string                 // 登录用户编号
	device        string                 // 登录设备
	token         string                 // 预定的token值
	timeout       int64                  // 指定此次登录 token 有效期，单位：秒 （如未指定，自动取全局配置的 timeout 值）
	activeTimeout int64                  // 指定此次登录 token 最低活跃频率，单位：秒（如未指定，则使用全局配置的 activeTimeout 值）
	extraData     map[string]interface{} // 额外参数
}

// GetExtraData 获取额外参数
func (l *LoginOptions) GetExtraData() map[string]interface{} {
	return l.extraData
}

// SetDevice 设置登录设备
//
//	@receiver l
//	@param device string
func (l *LoginOptions) SetDevice(device string) {
	l.device = device
}

// GetDevice 获取登录设备
//
//	@receiver l
//	@return string
func (l *LoginOptions) GetDevice() string {
	return l.device
}

// SetLoginId 设置登录用户
//
//	@receiver l
//	@param userId string
func (l *LoginOptions) SetLoginId(logiId string) {
	l.logiId = logiId
}

// GetLoginId 获取用户编号
//
//	@receiver l
//	@return string
func (l *LoginOptions) GetLoginId() string {
	return l.logiId
}

// SetActiveTimeout 设置最低活跃频率（单位：秒）
//
//	@receiver l
//	@param activeTimeout int64
func (l *LoginOptions) SetActiveTimeout(activeTimeout int64) {
	l.activeTimeout = activeTimeout
}

// GetActiveTimeout 获取此次登录的有效期
//
//	@receiver l
//	@return int64
func (l *LoginOptions) GetActiveTimeout() int64 {
	return l.activeTimeout
}

// SetTimeout 设置此次登录的有效期
//
//	@receiver l
//	@param timeout int64
func (l *LoginOptions) SetTimeout(timeout int64) {
	l.timeout = timeout
}

// GetTimeout 获取此次登录的有效时间
//
//	@receiver l
//	@return int64
func (l *LoginOptions) GetTimeout() int64 {
	return l.timeout
}

// SetToken 设置登录类型
//
//	@receiver l
//	@param loginType string
func (l *LoginOptions) SetToken(token string) {
	l.token = token
}

// GetToken 获取token值
//
//	@receiver l
//	@return string
func (l *LoginOptions) GetToken() string {
	return l.token
}

// Apply 合并参数
//
//	@receiver l
//	@param o *Config
//	@return string
func (l *LoginOptions) Apply(o *Config) {
	if l.timeout == 0 {
		l.timeout = o.Timeout
	}
	if l.activeTimeout == 0 {
		l.activeTimeout = o.ActiveTimeout
	}
}

// LoginOption 登录参数
type LoginOption func(o *LoginOptions)

// LoginWithDevice 设置登录参数
//
//	@param device string 设备信息
//	@return LoginOption 登录参数
func LoginWithDevice(device string) LoginOption {
	return func(o *LoginOptions) {
		o.device = device
	}
}
