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

const (
	InvalidToken        = "-1"          // InvalidToken 表示 token 无效
	InvalidTokenMessage = "token 无效"    // InvalidTokenMessage ...
	TimeoutToken        = "-2"          // TimeoutToken 表示 token 已过期
	TimeoutTokenMessage = "token 已过期"   // TimeoutTokenMessage 表示 token 已过期
	BeReplaced          = "-3"          // BeReplaced 表示 token 已被顶下线
	BeReplacedMessage   = "token 已被顶下线" // BeReplacedMessage 表示 token 已被顶下线
	KickOut             = "-4"          // KickOut 表示 token 已被踢下线
	KickOutMessage      = "token 已被踢下线" // KickOutMessage 表示 token 已被踢下线
	FreezeToken         = "-5"          // FreezeToken 表示 token 已被冻结
	FreezeTokenMessage  = "token 已被冻结"  // FreezeTokenMessage 表示 token 已被冻结
)

const (
	NeverExpire    int64 = -1 // NeverExpire 常量，表示一个 key 永不过期 （在一个 key 被标注为永远不过期时返回此值）
	NotValueExpire int64 = -2 // NotValueExpire 常量，表示系统中不存在这个缓存（在对不存在的 key 获取剩余存活时间时返回此值）
)

var (
	AbnormalList = []string{InvalidToken, TimeoutToken, BeReplaced, KickOut, FreezeToken} // AbnormalList token 异常标识集合
)

const (
	MinDisableLevel       int    = 1       // MinDisableLevel 常量 key 标记: 在封禁账号时，可使用的最小封禁级别
	DefaultDisableLevel   int    = 1       // DefaultDisableLevel 常量 key 标记: 在封禁账号时，默认封禁的等级
	DefaultDisableService string = "login" // DefaultDisableService 常量 key 标记，默认的禁用服务
)
