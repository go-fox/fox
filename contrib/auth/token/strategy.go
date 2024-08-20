package token

import (
	"errors"
	"strings"

	"github.com/duke-git/lancet/v2/random"
	"github.com/google/uuid"
)

// CreateTokenFunction 创建token的方法
type CreateTokenFunction func(loginId any, loginType string, style Style) string

// CheckTokenFunction 检测token是否唯一
type CheckTokenFunction func(tokenValue string) (bool, error)

// CreateSessionFunction 创建session策略
type CreateSessionFunction func(sessionId string, repository Repository) (Session, error)

// GenerateUniqueTokenFunction 生成唯一token
type GenerateUniqueTokenFunction func(element string, maxTryCount int, createTokenFunction func() string, checkTokenFunction func(value string) (bool, error)) (string, error)

// defaultCreateTokenFunction 默认创建token的方法
func defaultCreateTokenFunction(loginId any, loginType string, style Style) string {
	switch style {
	case StyleUUID:
		return uuid.New().String()
	case StyleSimpleUUID:
		return strings.ReplaceAll(uuid.New().String(), "_", "")
	case StyleRandom32:
		return random.RandString(32)
	case StyleRandom64:
		return random.RandString(64)
	default:
		return uuid.New().String()
	}
}

// defaultCreateSessionFunction 默认创建session的方法
func defaultCreateSessionFunction(sessionId string, repository Repository) (Session, error) {
	s := NewSession(repository)
	s.SetSessionId(sessionId)
	return s, nil
}

// defaultGenerateUniqueToken 默认生成唯一token的方法
func defaultGenerateUniqueToken(element string, maxTryCount int, createTokenFunction func() string, checkTokenFunction func(value string) (bool, error)) (string, error) {
	i := 1
	for {
		// 生成token
		value := createTokenFunction()
		// 如果 maxTryCount == -1，表示不做唯一性验证，直接返回
		if maxTryCount == -1 {
			return value, nil
		}
		v, err := checkTokenFunction(value)
		// 如果检查token不存在则直接返回
		if err == nil && v {
			return value, nil
		}
		if maxTryCount > 0 && i >= maxTryCount {
			return "", errors.New("maxTryCount exceeded")
		}
		i++
	}
}
