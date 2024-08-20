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

// Listener 监听器接口
type Listener interface {
	// DoLogin 事件发布：xx 账号登录
	//
	//  @param loginType string 登录类型
	//  @param loginId string 登录用户
	//  @param loginId string 登录的tokenValue值
	//  @param loginOptions LoginOptions 额外参数
	DoLogin(loginType string, loginId any, tokenValue string, loginOptions LoginOptions)

	// DoLogout 事件发布：xx 账号注销
	//
	//  @param logoutType string 登录类型
	//  @param logoutId any 登录用户
	//  @param tokenValue string token值
	DoLogout(logoutType string, logoutId any, tokenValue string)

	// DoReplaced 事件发布：xx 账号被顶下线
	//
	//  @param logoutType string 登录类型
	//  @param logoutId any 登录用户
	//  @param tokenValue string token值
	DoReplaced(logoutType string, logoutId any, tokenValue string)

	// DoDisable 事件发布：xx 账号被封禁
	//
	//  @param logoutType string 登录类型
	//  @param loginId any 登录账号id
	//  @param service string 封禁服务
	//  @param level int 封禁等级
	//  @param timeout int64 过期时间
	DoDisable(logoutType string, loginId any, service string, level int, timeout int64)
}
