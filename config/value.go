// Package config
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
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/go-fox/sugar/container/satomic"
	"google.golang.org/protobuf/proto"

	internalJson "github.com/go-fox/fox/codec/json"
)

var _ Value = (*atomicValue)(nil)
var _ Value = (*errValue)(nil)

func newAtomicValue(val any) Value {
	v := &atomicValue{}
	v.Store(val)
	return v
}

// Value 值接口
type Value interface {
	IsEmpty() bool
	Bool() (bool, error)
	Int() (int64, error)
	String() (string, error)
	Float() (float64, error)
	Duration() (time.Duration, error)
	Slice() ([]Value, error)
	Map() (map[string]Value, error)
	Scan(val interface{}) error
	Bytes() ([]byte, error)
	Store(interface{})
	Load() interface{}
}

type atomicValue struct {
	satomic.Value[any]
}

// Slice implements the interface Slice for Value
//
//	@receiver a
//	@return []Value
//	@return error
//	@player
func (a *atomicValue) Slice() ([]Value, error) {
	vals, ok := a.Load().([]interface{})
	if !ok {
		return nil, fmt.Errorf("type assert to %v failed", reflect.TypeOf(vals))
	}
	slices := make([]Value, 0, len(vals))
	for _, val := range vals {
		slices = append(slices, newAtomicValue(val))
	}
	return slices, nil
}

// Map implements the interface Map for Value
//
//	@receiver a
//	@return map[string]Value
//	@return error
//	@player
func (a *atomicValue) Map() (map[string]Value, error) {
	vals, ok := a.Load().(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("type assert to %v failed", reflect.TypeOf(vals))
	}
	m := make(map[string]Value, len(vals))
	for key, val := range vals {
		a := new(atomicValue)
		a.Store(val)
		m[key] = a
	}
	return m, nil
}

// Scan
//
//	@receiver a
//	@param val interface{}
//	@return error
//	@player
func (a *atomicValue) Scan(val interface{}) error {
	data, err := json.Marshal(a.Load())
	if err != nil {
		return err
	}
	if pb, ok := val.(proto.Message); ok {
		return internalJson.UnmarshalOptions.Unmarshal(data, pb)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	return decoder.Decode(val)
}

type errValue struct {
	err error
}

func (v *errValue) IsEmpty() bool {
	return true
}
func (v *errValue) Bool() (bool, error)              { return false, v.err }
func (v *errValue) Int() (int64, error)              { return 0, v.err }
func (v *errValue) Float() (float64, error)          { return 0.0, v.err }
func (v *errValue) Duration() (time.Duration, error) { return 0, v.err }
func (v *errValue) String() (string, error)          { return "", v.err }
func (v *errValue) Bytes() ([]byte, error) {
	return []byte{}, v.err
}
func (v *errValue) Scan(interface{}) error         { return v.err }
func (v *errValue) Load() interface{}              { return nil }
func (v *errValue) Store(interface{})              {}
func (v *errValue) Slice() ([]Value, error)        { return nil, v.err }
func (v *errValue) Map() (map[string]Value, error) { return nil, v.err }
