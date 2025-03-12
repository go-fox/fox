package token

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-fox/sugar/container/satomic"
)

type activeTimeValue struct {
	currentTime   *time.Time // 存储值的当前时间
	activeTimeout *int64     // 设置的过期时间
}

// newActiveTimeValue 构造函数
//
//	@param activeTimeStr string
//	@return *activeTimeValue
func newActiveTimeValue(activeTimeStr string) *activeTimeValue {
	res := &activeTimeValue{}
	if activeTimeStr == "" {
		return res
	}
	split := strings.Split(activeTimeStr, ",")
	if len(split) == 1 {
		timeMilli, _ := strconv.ParseInt(split[0], 10, 64)
		t := time.UnixMilli(timeMilli)
		res.currentTime = &t
	} else if len(split) == 2 {
		timeMilli, _ := strconv.ParseInt(split[0], 10, 64)
		v1 := time.UnixMilli(timeMilli)
		res.currentTime = &v1
		v2, _ := strconv.ParseInt(split[1], 10, 64)
		res.activeTimeout = &v2
	}
	return res
}

// newActiveTimeValueWithValue 构造函数
//
//	@param currentTime *time.Time
//	@param activeTimeout *int64
//	@return *activeTimeValue
func newActiveTimeValueWithValue(currentTime *time.Time, activeTimeout *int64) *activeTimeValue {
	return &activeTimeValue{
		currentTime:   currentTime,
		activeTimeout: activeTimeout,
	}
}

// getCurrentTime 获取存储时间
//
//	@receiver a
//	@return *time.Time
func (a *activeTimeValue) getCurrentTime() *time.Time {
	return a.currentTime
}

// getActiveTimeout 获取过期时间
//
//	@receiver a
//	@return int64
func (a *activeTimeValue) getActiveTimeout() *int64 {
	return a.activeTimeout
}

// Fmt 格式化为字符串
//
//	@receiver a
//	@return string
func (a *activeTimeValue) Fmt() string {
	if a.activeTimeout != nil {
		return strconv.FormatInt(a.currentTime.UnixMilli(), 10) + strconv.FormatInt(*a.activeTimeout, 10)
	}
	return strconv.FormatInt(a.currentTime.UnixMilli(), 10)
}

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

// Scan scan value to struct
func (a *atomicValue) Scan(val interface{}) error {
	data, err := json.Marshal(a.Load())
	if err != nil {
		return err
	}
	return json.Unmarshal(data, val)
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
