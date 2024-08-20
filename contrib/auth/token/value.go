package token

import (
	"strconv"
	"strings"
	"time"
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
