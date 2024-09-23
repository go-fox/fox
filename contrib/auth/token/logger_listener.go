package token

import (
	"log/slog"
	"time"
)

var _ Listener = (*LoggerListener)(nil)

func newLoggerListener(log *slog.Logger) Listener {
	return &LoggerListener{
		logger: &logger{
			Logger: log,
		},
	}
}

// LoggerListener 日志监听器
type LoggerListener struct {
	logger *logger
}

// DoLogin 事件发布：xx 账号登录
//
//	@param loginType string 登录类型
//	@param loginId string 登录用户
//	@param loginId string 登录的tokenValue值
//	@param loginOptions LoginOptions 额外参数
func (l *LoggerListener) DoLogin(loginType string, loginId any, tokenValue string, loginOptions LoginOptions) {
	l.logger.Infof("账号 %v 登录成功 (loginType=%s), 会话凭证 token=%s", loginId, loginType, tokenValue)
}

// DoLogout 事件发布：xx 账号注销
//
//	@param logoutType string 退出类型
//	@param logoutId any 退出用户id
//	@param tokenValue string token值
func (l *LoggerListener) DoLogout(logoutType string, logoutId any, tokenValue string) {
	l.logger.Infof("账号 %v 注销成功 (loginType=%s), 会话凭证 token=%s", logoutId, logoutType, tokenValue)
}

// DoReplaced 事件发布：xx 账号被顶下线
//
//	@param logoutType string 登录类型
//	@param logoutId any 登录用户
//	@param tokenValue string token值
func (l *LoggerListener) DoReplaced(logoutType string, logoutId any, tokenValue string) {
	l.logger.Infof("账号 %v 被顶下线 (loginType=%s), 会话凭证 token=%s", logoutId, logoutType, tokenValue)
}

// DoDisable 事件发布：xx 账号被封禁
//
//	@param logoutType string 登录类型
//	@param loginId any 登录账号id
//	@param service string 封禁服务
//	@param level int 封禁等级
//	@param timeout int64 过期时间
func (l *LoggerListener) DoDisable(logoutType string, loginId any, service string, level int, timeout int64) {
	t := time.Second * time.Duration(timeout)
	l.logger.Infof("账号 %v 被封禁 (loginType=%s),  封禁服务：%s 封禁等级：%d 封禁时间：%s", loginId, logoutType, service, level, t.String())
}
