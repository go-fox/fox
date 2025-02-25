// Package websocket
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
package websocket

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/go-fox/sugar/container/satomic"
	"github.com/go-fox/sugar/container/spool"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"

	"github.com/go-fox/fox/api/gen/go/protocol"
	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/errors"
)

var sessionPool = spool.New[*Session](func() *Session {
	return &Session{}
}, func(ss *Session) {
	ss.Id = ""
	ss.baseCtx = nil
	ss.conn = nil
	ss.md = nil
	ss.lastActiveTime = nil
	ss.storeMu = nil
	ss.codec = nil
})

// Session websocket client session
type Session struct {
	Id             string `json:"id"`
	baseCtx        context.Context
	conn           *websocket.Conn // 连接
	md             metadata.MD     // 元数据
	lastActiveTime *satomic.Value[time.Time]
	storeMu        *sync.RWMutex
	codec          codec.Codec
	closeOnce      sync.Once
}

func acquireSession(
	ctx context.Context,
	codec codec.Codec,
	conn *websocket.Conn,
) *Session {
	ss := sessionPool.Get()
	ss.Id = uuid.New().String()
	ss.baseCtx = ctx
	ss.conn = conn
	ss.md = metadata.Pairs()
	ss.lastActiveTime = satomic.New[time.Time]()
	ss.storeMu = &sync.RWMutex{}
	ss.closeOnce = sync.Once{}
	ss.codec = codec
	return ss
}

func releaseSession(ss *Session) {
	sessionPool.Put(ss)
}

// ID is client session unique id
func (s *Session) ID() string {
	return s.Id
}

// Send 发送
func (s *Session) Send(reply *protocol.Reply) error {
	// 编码
	message, err := s.codec.Marshal(reply)
	if err != nil {
		return err
	}
	//  发送bytes
	if err = s.conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
		return err
	}
	s.lastActiveTime.Store(time.Now())
	return nil
}

// Receive 接收
func (s *Session) Receive(request *protocol.Request) error {
	for {
		select {
		case <-s.baseCtx.Done():
			return io.EOF
		default:
		}
		messageType, message, err := s.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(
				err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
				websocket.CloseNormalClosure,
				websocket.CloseNoStatusReceived,
			) {
				return errors.ClientClosed("WEB_SOCKET_CLIENT_CLOSE", err.Error())
			}
			return err
		}
		// 更新最新活跃时间
		s.lastActiveTime.Store(time.Now())
		// 解码数据
		if messageType == websocket.BinaryMessage {
			if err = s.codec.Unmarshal(message, request); err != nil {
				return errors.FromError(err)
			}
		}
		return nil
	}
}

// Store 存储元数据
func (s *Session) Store(key, value string) {
	s.storeMu.Lock()
	defer s.storeMu.Unlock()
	s.md.Set(key, value)
}

// Load 加载数据
func (s *Session) Load(key string) string {
	s.storeMu.RLock()
	defer s.storeMu.RUnlock()
	vals := s.md.Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// Close 关闭
func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		s.conn.Close()
	})
	return nil
}
