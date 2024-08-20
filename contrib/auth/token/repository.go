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

// Repository 存储接口
type Repository interface {
	// Set 设置数据
	//
	//  @param key string 存储key
	//  @param value string 存储值
	//  @param timeout int64 过期时间
	//  @return error
	Set(key string, value string, timeout int64) error
	// Get 获取字符串值
	//
	//  @param key string 存储键
	//  @return string 返回值
	//  @return error 是否有错
	Get(key string) (string, error)
	// Delete
	//
	//  @param key string 存储键
	//  @return error 错误信息
	Delete(key string) error
	// Update 修改字符串值
	//
	//  @param key string 存储key
	//  @param val string 存储值
	//  @return error
	Update(key string, val string) error
	// setSession
	//
	//  @param s Session 要保存的 SaSession 对象
	//  @param timeout int64 过期时间，单位：秒
	//  @return error
	setSession(s Session, timeout int64) error
	// GetSession 获取session
	//
	//  @param sessionId string
	//  @return Session
	GetSession(sessionId string) (Session, error)
	// UpdateObject 修改值
	//
	//  @param sessionId string
	//  @param obj any
	//  @return error
	UpdateObject(sessionId string, obj any) error
	// UpdateObjectTimeout 修改 Object 的剩余存活时间（单位: 秒）
	//
	//  @param key string 指定 key
	//  @param timeout int64 剩余活跃时间
	//  @return error
	UpdateObjectTimeout(key string, timeout int64) error
	// GetObjectTimeout 获取指定对象的剩余活跃时间（单位：秒）
	//
	//  @param key string 指定key
	//  @return int64 剩余存活时间
	//  @return error
	GetObjectTimeout(key string) (int64, error)
}
