// Package token
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
package token

// ListenerManager 监听器管理
type ListenerManager struct {
	listeners []Listener
}

// newListenerCenter 构造函数
//
//	@param config *Config
//	@return *ListenerCenter
func newListenerCenter(opts *Config) *ListenerManager {
	listeners := make([]Listener, 0)
	if opts.logger != nil {
		listeners = append(listeners, newLoggerListener(opts.logger))
	}
	return &ListenerManager{
		listeners: listeners,
	}
}

// RegisterListener 添加监听器
//
//	@receiver l
//	@param listener Listener
func (l *ListenerManager) RegisterListener(listener ...Listener) {
	l.listeners = append(l.listeners, listener...)
}

// SetListener 重置监听器
//
//	@receiver l
//	@param listener []Listener
func (l *ListenerManager) SetListener(listener []Listener) {
	l.listeners = listener
}

// GetListener 获取监听器
//
//	@receiver l
//	@return []Listener
func (l *ListenerManager) GetListener() []Listener {
	return l.listeners
}

// DoLogin 执行登录钩子
//
//	@receiver l
//	@param loginType string
//	@param loginId any
//	@param loginOptions LoginOptions登录参数
func (l *ListenerManager) DoLogin(loginType string, loginId any, tokenValue string, loginOptions LoginOptions) {
	for _, listener := range l.listeners {
		go listener.DoLogin(loginType, loginId, tokenValue, loginOptions)
	}
}

// DoLogout 执行退出钩子
//
//	@receiver l
//	@param logoutType string
//	@param logoutId any
//	@param tokenValue string
func (l *ListenerManager) DoLogout(logoutType string, logoutId any, tokenValue string) {
	for _, listener := range l.listeners {
		go listener.DoLogout(logoutType, logoutId, tokenValue)
	}
}

// DoReplaced 执行被顶钩子
//
//	@receiver l
//	@param logoutType string
//	@param logoutId any
//	@param tokenValue string
func (l *ListenerManager) DoReplaced(logoutType string, logoutId any, tokenValue string) {
	for _, listener := range l.listeners {
		go listener.DoReplaced(logoutType, logoutId, tokenValue)
	}
}

// DoDisable 执行被封禁钩子
//
//	@receiver l
//	@param logoutType string
//	@param logoutId any
//	@param service string
//	@param level int
//	@param timeout int64
func (l *ListenerManager) DoDisable(logoutType string, logoutId any, service string, level int, timeout int64) {
	for _, listener := range l.listeners {
		listener.DoDisable(logoutType, logoutId, service, level, timeout)
	}
}
