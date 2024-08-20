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
	"bytes"
	"encoding/json"
	"math"
	"sync"
)

var _ Session = (*session)(nil)

// SessionType session的类型
type SessionType string

const (
	// AccountSessionType 账号类型
	AccountSessionType SessionType = "account-session"
	// TokenSessionType token类型
	TokenSessionType SessionType = "token-session"
	// CustomSessionType 自定义类型
	CustomSessionType SessionType = "custom-session"
)

// SerializeData 序列化数据
type SerializeData struct {
	Id          string         `json:"id"`                 // session唯一标识
	SessionType SessionType    `json:"session_type"`       // session类型
	Data        map[string]any `json:"data"`               // 挂载数据
	LoginType   string         `json:"login_type"`         // 所属loginType
	LoginId     any            `json:"login_id,omitempty"` // 当session为account的时候有效
	Token       string         `json:"token,omitempty"`    // token值，当sessionType为SessionTypeToken时值有效
	SignList    SignList       `json:"sign_list"`          // 签名列表
	CreatedAt   int64          `json:"created_at"`         // 创建时间
}

// Session 接口
type Session interface {
	GetSessionId() string
	SetSessionId(id string)
	SetSessionType(sessionType SessionType)
	GetSessionType() SessionType
	AddTokenSign(sign *Sign) error
	SetLoginType(loginType string)
	GetLoginType() string
	SetLoginId(loginId any)
	GetLoginId() any
	SetToken(tokenValue string)
	GetToken() string
	Decode(data []byte) error
	Encode() ([]byte, error)
	UpdateMinTimeout(timeout int64) error
	GetTokenSignListByDevice(device string) SignList
	GetTokenValueListByDevice(device string) []string
	RemoveTokenSign(value string) error
	LogoutByTokenSignCountIsZero() error
}

// session 会话作用域读取值对象
type session struct {
	serializeData *SerializeData // 序列化数据
	repository    Repository     // 底层存储
	lock          sync.RWMutex   // 锁
}

// NewSession 创建session
//
//	@param storage
//	@return Session
func NewSession(repository Repository) Session {
	return &session{
		repository:    repository,
		lock:          sync.RWMutex{},
		serializeData: &SerializeData{},
	}
}

// SetSessionId 设置sessionId
//
//	@receiver s
//	@param sessionId string
func (s *session) SetSessionId(sessionId string) {
	s.serializeData.Id = sessionId
}

// GetSessionId 获取
//
//	@receiver s
//	@return string
func (s *session) GetSessionId() string {
	return s.serializeData.Id
}

// SetLoginType 设置登录类型
//
//	@receiver s
//	@param loginType string 登录类型：例如多个用户体系，admin, user
func (s *session) SetLoginType(loginType string) {
	s.serializeData.LoginType = loginType
}

// GetLoginType 获取登录类型
//
//	@receiver s
//	@return string
func (s *session) GetLoginType() string {
	return s.serializeData.LoginType
}

// SetLoginId 设置登录账号
//
//	@receiver s
//	@param loginId string
func (s *session) SetLoginId(loginId any) {
	s.serializeData.LoginId = loginId
}

// GetLoginId 获取登录账号
//
//	@receiver s
//	@return any
func (s *session) GetLoginId() any {
	return s.serializeData.LoginId
}

// SetToken 设置tokenValue，只有当session为token-session时有效
//
//	@receiver s
//	@param tokenValue string
func (s *session) SetToken(tokenValue string) {
	s.serializeData.Token = tokenValue
}

// GetToken 获取tokenValue，只有当session为Token-Session时有效
//
//	@receiver s
//	@return string
func (s *session) GetToken() string {
	return s.serializeData.Token
}

// Encode 编码
//
//	@receiver s
//	@return []byte 编码后的数据
//	@return error 编码是否有错
func (s *session) Encode() ([]byte, error) {
	return json.Marshal(s.serializeData)
}

// Decode 解码
//
//	@receiver s
//	@param data []byte 编码后的数据
//	@return error 是否有错
func (s *session) Decode(data []byte) error {
	decoder := json.NewDecoder(bytes.NewBuffer(data))
	decoder.UseNumber()
	return decoder.Decode(&s.serializeData)
}

// SetSessionType 设置session类型
//
//	@receiver s
//	@param sessionType SessionType session类型
func (s *session) SetSessionType(sessionType SessionType) {
	s.serializeData.SessionType = sessionType
}

// AddTokenSign 添加登录凭证
//
//	@receiver s
//	@param sign Sign
//	@return SessionType
func (s *session) AddTokenSign(sign *Sign) error {
	oldToken := s.getTokenSign(sign.Value)
	if oldToken == nil {
		s.serializeData.SignList = append(s.serializeData.SignList, sign)
		return s.update()
	}
	oldToken.Value = sign.Value
	oldToken.Device = sign.Device
	oldToken.Extra = sign.Extra
	return s.update()
}

// getTokenSignListCopy 复制数据
//
//	@receiver s
//	@return SignList
func (s *session) getTokenSignListCopy() SignList {
	s2 := make([]*Sign, len(s.serializeData.SignList))
	copy(s2, s.serializeData.SignList)
	return s2
}

// getTokenSign 根据token值获取sign
//
//	@receiver s
//	@param tokenValue string token值
//	@return *Sign
func (s *session) getTokenSign(tokenValue string) *Sign {
	for _, sign := range s.getTokenSignListCopy() {
		if sign.Value == tokenValue {
			return sign
		}
	}
	return nil
}

// GetSessionType 获取当前session类型
//
//	@receiver s
//	@return SessionType
func (s *session) GetSessionType() SessionType {
	return s.serializeData.SessionType
}

// LogoutByTokenSignCountIsZero 如果token签名数量为0，则直接注销account-session
//
//	@receiver s
//	@return error
func (s *session) LogoutByTokenSignCountIsZero() error {
	if len(s.serializeData.SignList) == 0 {
		return s.logout()
	}
	return nil
}

// update 更新
//
//	@receiver s
//	@return error
func (s *session) update() error {
	return s.repository.UpdateObject(s.serializeData.Id, s.serializeData)
}

// logout 退出
//
//	@receiver s
//	@return error 错误信息
func (s *session) logout() error {
	return s.repository.Delete(s.serializeData.Id)
}

// UpdateMinTimeout 修改此Session的最小剩余存活时间 (只有在 Session 的过期时间低于指定的 minTimeout 时才会进行修改)
//
//	@receiver s
//	@param ttl time.Duration 过期时间
//	@return error
func (s *session) UpdateMinTimeout(minTimeout int64) error {
	min := s.trans(minTimeout)
	timeout, err := s.GetTimeout()
	if err != nil {
		return err
	}
	curr := s.trans(timeout)
	if curr < min {
		return s.updateTimeout(minTimeout)
	}
	return nil
}

// GetTokenSignListByDevice 根据设备获取签名列表
//
//	@receiver s
//	@param device string 指定设备
//	@return SignList 签名列表
//	@return error 是否有错
func (s *session) GetTokenSignListByDevice(device string) SignList {
	if device == "" {
		return s.getTokenSignListCopy()
	}
	signList := s.getTokenSignListCopy()
	res := make(SignList, 0)
	for _, sign := range signList {
		if sign.Device == device {
			res = append(res, sign)
		}
	}
	return res
}

// GetTokenValueListByDevice 获取token值列表
//
//	@receiver s
//	@param device string 设备信息
//	@return []string token值列表
func (s *session) GetTokenValueListByDevice(device string) []string {
	tokenList := s.getTokenSignListCopy()
	res := make([]string, 0)
	for _, token := range tokenList {
		if device == "" || token.Device == device {
			res = append(res, token.Value)
		}
	}
	return res
}

// RemoveTokenSign 移除一个token签名
//
//	@receiver s
//	@param value string token值
//	@return error 是否有错
func (s *session) RemoveTokenSign(value string) error {
	tokenSign := s.getTokenSign(value)
	type nullInt struct {
		value   int
		isValid bool
	}
	var index = nullInt{}
	for i, sign := range s.serializeData.SignList {
		if sign.Value == tokenSign.Value {
			index.value = i
			index.isValid = true
			break
		}
	}
	if index.isValid {
		s.serializeData.SignList = append(s.serializeData.SignList[:index.value], s.serializeData.SignList[index.value+1:]...)
		return s.update()
	}
	return nil
}

// trans 修复时间
//
//	@receiver s
//	@param value int64
//	@return int64
func (s *session) trans(value int64) int64 {
	if value == NeverExpire {
		return math.MaxInt32
	}
	return value
}

// updateTimeout 修改存活时间（单位：秒）
//
//	@receiver s
//	@param timeout int64 剩余存活时间
//	@return error
func (s *session) updateTimeout(timeout int64) error {
	return s.repository.UpdateObjectTimeout(s.serializeData.Id, timeout)
}

// GetTimeout 获取当前session的有效存活时间（单位：秒）
//
//	@receiver s
//	@return int64
func (s *session) GetTimeout() (int64, error) {
	return s.repository.GetObjectTimeout(s.GetSessionId())
}
