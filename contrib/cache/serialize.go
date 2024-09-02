// Package cache
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
package cache

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/codec/json"
)

var (
	defaultCodec      = codec.GetCodec(json.Name)
	defaultSerializer = &Serializer{}
)

// DefaultSerializer 默认的序列化
func DefaultSerializer() *Serializer {
	return defaultSerializer
}

// Serializer 序列化
type Serializer struct{}

// Marshal 序列化
func (s *Serializer) Marshal(v any) ([]byte, error) {
	if v == nil {
		return nil, errors.New("serializer can not be nil")
	}
	switch v := v.(type) {
	case string:
		return s.stringToBytes(v), nil
	case *string:
		return s.stringToBytes(*v), nil
	case []byte:
		return v, nil
	case *[]byte:
		return *v, nil
	case int:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case *int:
		return strconv.AppendInt(nil, int64(*v), 10), nil
	case int8:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case *int8:
		return strconv.AppendInt(nil, int64(*v), 10), nil
	case int16:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case *int16:
		return strconv.AppendInt(nil, int64(*v), 10), nil
	case int32:
		return strconv.AppendInt(nil, int64(v), 10), nil
	case *int32:
		return strconv.AppendInt(nil, int64(*v), 10), nil
	case int64:
		return strconv.AppendInt(nil, v, 10), nil
	case *int64:
		return strconv.AppendInt(nil, int64(*v), 10), nil
	case uint:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case *uint:
		return strconv.AppendUint(nil, uint64(*v), 10), nil
	case uint8:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case *uint8:
		return strconv.AppendUint(nil, uint64(*v), 10), nil
	case uint16:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case *uint16:
		return strconv.AppendUint(nil, uint64(*v), 10), nil
	case uint32:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case uint64:
		return strconv.AppendUint(nil, v, 10), nil
	case *uint64:
		return strconv.AppendUint(nil, *v, 10), nil
	case uintptr:
		return strconv.AppendUint(nil, uint64(v), 10), nil
	case *uintptr:
		return strconv.AppendUint(nil, uint64(*v), 10), nil
	case float32:
		return strconv.AppendFloat(nil, float64(v), 'g', -1, 32), nil
	case *float32:
		return strconv.AppendFloat(nil, float64(*v), 'g', -1, 32), nil
	case float64:
		return strconv.AppendFloat(nil, v, 'g', -1, 64), nil
	case *float64:
		return strconv.AppendFloat(nil, *v, 'g', -1, 64), nil
	case bool:
		return strconv.AppendBool(nil, v), nil
	case *bool:
		return strconv.AppendBool(nil, *v), nil
	case time.Time:
		return strconv.AppendInt(nil, v.Unix(), 10), nil
	case *time.Time:
		return strconv.AppendInt(nil, v.Unix(), 10), nil
	default:
		return defaultCodec.Marshal(v)
	}
}

// Unmarshal 反序列化
func (s *Serializer) Unmarshal(b []byte, v any) error {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return errors.New("serializer can only be a pointer to a struct")
	}
	switch v := v.(type) {
	case nil:
		return fmt.Errorf("redis: Scan(nil)")
	case *string:
		*v = s.bytesToString(b)
		return nil
	case *[]byte:
		*v = b
		return nil
	case *int:
		var err error
		*v, err = s.atoi(b)
		return err
	case *int8:
		n, err := s.parseInt(b, 10, 8)
		if err != nil {
			return err
		}
		*v = int8(n)
		return nil
	case *int16:
		n, err := s.parseInt(b, 10, 16)
		if err != nil {
			return err
		}
		*v = int16(n)
		return nil
	case *int32:
		n, err := s.parseInt(b, 10, 32)
		if err != nil {
			return err
		}
		*v = int32(n)
		return nil
	case *int64:
		n, err := s.parseInt(b, 10, 64)
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *uint:
		n, err := s.parseUint(b, 10, 64)
		if err != nil {
			return err
		}
		*v = uint(n)
		return nil
	case *uint8:
		n, err := s.parseUint(b, 10, 8)
		if err != nil {
			return err
		}
		*v = uint8(n)
		return nil
	case *uint16:
		n, err := s.parseUint(b, 10, 16)
		if err != nil {
			return err
		}
		*v = uint16(n)
		return nil
	case *uint32:
		n, err := s.parseUint(b, 10, 32)
		if err != nil {
			return err
		}
		*v = uint32(n)
		return nil
	case *uint64:
		n, err := s.parseUint(b, 10, 64)
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *float32:
		n, err := s.parseFloat(b, 32)
		if err != nil {
			return err
		}
		*v = float32(n)
		return err
	case *float64:
		var err error
		*v, err = s.parseFloat(b, 64)
		return err
	case *bool:
		*v = s.bytesToString(b) == "true"
		return nil
	case *time.Time:
		var err error
		*v, err = time.Parse(time.RFC3339Nano, s.bytesToString(b))
		return err
	case *time.Duration:
		n, err := s.parseInt(b, 10, 64)
		if err != nil {
			return err
		}
		*v = time.Duration(n)
		return nil
	case *net.IP:
		*v = b
		return nil
	default:
		return defaultCodec.Unmarshal(b, v)
	}
}

func (s *Serializer) atoi(b []byte) (int, error) {
	return strconv.Atoi(s.bytesToString(b))
}

func (s *Serializer) parseInt(b []byte, base int, bitSize int) (int64, error) {
	return strconv.ParseInt(s.bytesToString(b), base, bitSize)
}

func (s *Serializer) parseUint(b []byte, base int, bitSize int) (uint64, error) {
	return strconv.ParseUint(s.bytesToString(b), base, bitSize)
}

func (s *Serializer) parseFloat(b []byte, bitSize int) (float64, error) {
	return strconv.ParseFloat(s.bytesToString(b), bitSize)
}

// BytesToString converts byte slice to string.
func (s *Serializer) bytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// StringToBytes converts string to byte slice.
func (s *Serializer) stringToBytes(v string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{v, len(v)},
	))
}
