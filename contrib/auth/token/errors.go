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

// NotLoginError 没有登录错误
type NotLoginError struct {
	errType   string
	loginType string
	message   string
	token     string
}

// Error 实现错误接口
//
//	@receiver e
//	@return string
func (e NotLoginError) Error() string {
	if len(e.token) > 0 {
		return e.message + ": " + e.token
	}
	return e.message
}

// GetType 获取错误类型
//
//	@receiver e
//	@return string
func (e NotLoginError) GetType() string {
	return e.errType
}

// GetLoginType 获取登录类型
//
//	@receiver e
//	@return string
func (e NotLoginError) GetLoginType() string {
	return e.loginType
}

// GetToken 获取token
//
//	@receiver e
//	@return string
func (e NotLoginError) GetToken() string {
	return e.token
}

// NewNotLoginError 构建一个未登录错误
//
//	@param errorType string
//	@param loginType string
//	@param message string
//	@param token ...string
//	@param
//	@return *NotLoginError
func NewNotLoginError(
	errorType string,
	loginType string,
	message string,
	token ...string,
) *NotLoginError {
	tk := ""
	if len(token) > 0 {
		tk = token[0]
	}
	return &NotLoginError{
		errType:   errorType,
		loginType: loginType,
		message:   message,
		token:     tk,
	}
}
