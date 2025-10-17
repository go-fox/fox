package token

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	set "github.com/duke-git/lancet/v2/datastructure/set"
)

var _ Token = (*token)(nil)

// Token token接口
type Token interface {
	// Login 登录方法
	//
	//  @param ctx context.Context
	//  @param loginId any 登录账号id，推荐使用int, int64,string
	//  @param config ...LoginOption 登录额外参数
	//  @return string 登录后的token
	//  @return error 是否有错
	Login(ctx context.Context, loginId any, opts ...LoginOption) (string, error)
	// Logout 退出方法
	//
	//  @param loginId any 登录账号id
	//  @return error 是否有错
	//  @return error
	Logout(ctx context.Context, loginId any) error
	// LogoutByDevice 退出指定设备
	//
	//  @param ctx context.Context
	//  @param loginId any 退出账号id
	//  @param device string 退出的设备
	//  @return error 是否有错
	LogoutByDevice(ctx context.Context, loginId any, device string) error
	// LogoutByTokenValue 指定token值退出
	//
	//  @param ctx context.Context
	//  @param tokenValue string 指定的token值
	//  @return error
	LogoutByTokenValue(ctx context.Context, tokenValue string) error
	// Replaced 顶人下线，根据账号id 和 设备类型
	//
	//  @param ctx context.Context
	//  @param loginId any 登录账号id
	//  @param device string
	//  @return error
	Replaced(ctx context.Context, loginId any, device string) error
	// IsLogin 查询当前token是否登录
	//
	//  @param tokenValue string token值
	//  @return bool 是否登录
	//  @return error 查询过程中是否有错
	IsLogin(ctx context.Context, tokenValue string) (bool, error)
	// IsLoginByLoginId 查询指定账号是否登录
	//
	//  @param ctx context.Context
	//  @param loginId any 登录账号id
	//  @return bool 是否登录
	//  @return error 查询过程中是否有错
	IsLoginByLoginId(ctx context.Context, loginId any) (bool, error)
	// GetLoginIdAsInt 获取登录账号（数字类型）
	//
	//  @param tokenValue string
	//  @return any
	//  @return error
	GetLoginIdAsInt(ctx context.Context, tokenValue string) (int64, error)
	// GetLoginIdAsString 获取登录账号（字符串类型）
	//
	//  @param ctx context.Context 携带上下文
	//  @param tokenValue string token值
	//  @return string 字符串账号
	//  @return error 是否出错
	GetLoginIdAsString(ctx context.Context, tokenValue string) (string, error)
	// GetSessionByLoginId 获取指定用户id的 Account-Session
	//
	//  @param ctx context.Context
	//  @param loginId any 登录id
	//  @param isCreate bool 如果没有是否创建
	//  @return Session session结构
	//  @return error 是否有错
	GetSessionByLoginId(ctx context.Context, loginId any, isCreate bool) (Session, error)
	// GetSessionByLoginIdDefault 获取指定用户的 Account-session，如果没有则默认创建一个
	//
	//  @param ctx context.Context
	//  @param LoginId any 登录账号
	//  @return Session session结构
	//  @return error 是否有错
	GetSessionByLoginIdDefault(ctx context.Context, LoginId any) (Session, error)
	// GetTokenSessionByTokenValue 获取指定 token 的 Token-Session，如果该 SaSession 尚未创建，isCreate代表是否新建并返回
	//
	//  @param ctx context.Context
	//  @param tokenValue string 指定token值
	//  @param isCreate bool 如果没有，是否新建
	//  @return Session session结构
	//  @return error 是否有错
	GetTokenSessionByTokenValue(ctx context.Context, tokenValue string, isCreate bool) (Session, error)
	// Disable 封禁账号
	//
	//  @param ctx context.Context
	//	@param loginId any 指定账号id
	//	@param timeout int64 封禁时间, 单位: 秒 （-1=永久封禁）
	//	@return error
	Disable(ctx context.Context, loginId any, service string, timeout int64) error
	// DisableLevel 封禁：指定账号的指定服务，并指定封禁等级
	//
	//  @param ctx context.Context
	//	@param loginId any 指定账号id
	//	@param service string 指定封禁服务
	//	@param level int 指定封禁等级
	//	@param timeout int64 封禁时间, 单位: 秒 （-1=永久封禁）
	//	@return error
	DisableLevel(ctx context.Context, loginId any, service string, level int, timeout int64) error
}

// token 实例
type token struct {
	config *Config
}

// New create token with option
//
//	@param config ...Option 创建参数
//	@return Token 实例
func New(opts ...Option) Token {
	conf := DefaultConfig()
	for _, opt := range opts {
		opt(conf)
	}
	return NewWithConfig(conf)
}

// NewWithConfig create token with config
func NewWithConfig(configs ...*Config) Token {
	conf := DefaultConfig()
	if len(configs) > 0 {
		conf = configs[0]
	}
	return &token{conf}
}

// Login 登录方法
//
//	@receiver t
//	@param ctx context.Context 上下文
//	@param loginId any 推荐使用（int64，int，string）类型
//	@param config ...LoginOption 登录参数
//	@return string
//	@return error
func (t *token) Login(ctx context.Context, loginId any, opts ...LoginOption) (string, error) {
	// 1.额外参数拼装
	o := LoginOptions{}
	for _, opt := range opts {
		opt(&o)
	}
	// 2.检测参数
	err := t.checkLoginArgs(loginId, o)
	if err != nil {
		return "", err
	}
	// 3.补充参数
	o.Apply(t.config)
	// 4.分配一个可用的token
	tokenValue, err := t.distUsableToken(ctx, loginId, o)
	if err != nil {
		return "", err
	}
	// 5、获取此账号的 Account-Session , 续期
	se, err := t.getSessionByLoginId(ctx, loginId, true)
	if err != nil {
		return "", err
	}
	err = se.updateMinTimeout(ctx, o.GetTimeout())
	if err != nil {
		return "", err
	}
	// 6、在 Account-Session 上记录本次登录的 token 签名
	err = se.addTokenSign(ctx, NewSign(tokenValue, o.GetDevice(), o.GetExtraData()))
	if err != nil {
		return "", err
	}
	// 7、保存 token -> id 的映射关系，方便日后根据 token 找账号 id
	err = t.saveTokenToIdMapping(ctx, tokenValue, loginId, o.GetTimeout())
	if err != nil {
		return "", err
	}
	// 8、如果开启了活跃度校验，写入这个 token 的最后活跃时间 token-last-active
	if t.isOpenCheckActiveTimeout() {
		err = t.setLastActiveToNow(ctx, tokenValue, o.GetTimeout(), o.GetActiveTimeout())
		if err != nil {
			return "", err
		}
	}
	// 9、使用协程发布全局事件：账号 xxx 登录成功
	t.config.listener.DoLogin(t.config.LoginType, loginId, tokenValue, o)

	// 10、检查此账号会话数量是否超出最大值，如果超过，则按照登录时间顺序，把最开始登录的给注销掉
	if t.config.MaxLoginCount != -1 {
		err = t.logoutByMaxLoginCount(ctx, loginId, se, "", t.config.MaxLoginCount)
		if err != nil {
			return "", err
		}
	}
	return tokenValue, nil
}

// Logout 根据登录id退出
//
//	@receiver t
//	@param loginId any 登录id
//	@return error 错误信息
func (t *token) Logout(ctx context.Context, loginId any) error {
	return t.LogoutByDevice(ctx, loginId, "")
}

// LogoutByDevice 退出登录根据登录设备
//
//	@receiver t
//	@param loginId any 登陆账号
//	@param device string 登陆设备
//	@return error
func (t *token) LogoutByDevice(ctx context.Context, loginId any, device string) error {
	ss, err := t.getSessionByLoginId(ctx, loginId, false)
	if err != nil {
		return err
	}
	if ss != nil {
		// 2、遍历此账号所有从这个 device 设备上登录的客户端，清除相关数据
		for _, sign := range ss.getTokenSignListByDevice(device) {
			tokenValue := sign.Value
			// 2.1、从 Account-Session 上清除 token 签名
			if err := ss.removeTokenSign(ctx, tokenValue); err != nil {
				return err
			}

			// 2.2、清除这个 token 的最后活跃时间记录
			if t.isOpenCheckActiveTimeout() {
				if err := t.clearLastActive(tokenValue); err != nil {
					return err
				}
			}

			// 2.3、清除 token -> id 的映射关系
			if err := t.deleteTokenToIdMapping(tokenValue); err != nil {
				return err
			}

			// 2.4、清除这个 token 的 Token-Session 对象
			if err := t.deleteTokenSession(ctx, tokenValue); err != nil {
				return err
			}

			// 2.5、$$ 发布事件：xx 账号的 xx 客户端注销了
			t.config.listener.DoLogout(t.config.LoginType, loginId, tokenValue)
		}
		// 3、如果代码走到这里的时候，此账号已经没有客户端在登录了，则直接注销掉这个 Account-Session
		err = ss.logoutByTokenSignCountIsZero(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// LogoutByTokenValue 退出登录
//
//	@receiver t
//	@param tokenValue string
//	@return error
func (t *token) LogoutByTokenValue(ctx context.Context, tokenValue string) error {
	// 如果没有token值则直接跳过
	if len(tokenValue) == 0 {
		return nil
	}
	// 1、清除这个 token 的最后活跃时间记录
	if t.isOpenCheckActiveTimeout() {
		err := t.clearLastActive(tokenValue)
		if err != nil {
			return err
		}
	}
	// 2、清除这个 token 的 Token-Session 对象
	err := t.deleteTokenSession(ctx, tokenValue)
	if err != nil {
		return err
	}

	// 3、清除 token -> id 的映射关系
	loginId, err := t.getLoginIdNotHandle(ctx, tokenValue)
	if err != nil {
		return err
	}
	if len(loginId) > 0 {
		err = t.deleteTokenToIdMapping(tokenValue)
		if err != nil {
			return err
		}
	}

	// 4、判断一下：如果此 token 映射的是一个无效 loginId，则此处立即返回，不需要再往下处理了
	if !t.isValidLoginId(loginId) {
		return nil
	}

	// 5、$$ 发布事件：某某账号的某某 token 注销下线了
	t.config.listener.DoLogout(t.config.LoginType, loginId, tokenValue)

	// 6、清理这个账号的 Account-Session 上的 token 签名，并且尝试注销掉 Account-Session
	ss, err := t.getSessionByLoginId(ctx, loginId, false)
	if err != nil {
		return err
	}
	if ss != nil {
		err = ss.removeTokenSign(ctx, tokenValue)
		if err != nil {
			return err
		}
		err = ss.logoutByTokenSignCountIsZero(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Replaced 顶人下线，根据账号id 和 设备类型 ，当用户顶下线后，再次访问则会返回
func (t *token) Replaced(ctx context.Context, loginId any, device string) error {
	ss, err := t.getSessionByLoginId(ctx, loginId, false)
	if err != nil {
		return err
	}
	if ss != nil {
		for _, sign := range ss.getTokenSignListByDevice(device) {
			tokenValue := sign.Value
			// 2.1、从 Account-Session 上清除 token 签名
			if err := ss.removeTokenSign(ctx, tokenValue); err != nil {
				return err
			}

			// 2.2、清除这个 token 的最后活跃时间记录
			if t.isOpenCheckActiveTimeout() {
				if err := t.clearLastActive(tokenValue); err != nil {
					return err
				}
			}

			// 2.3 将此 token 标记为：已被顶下线
			if err := t.updateTokenToIdMapping(ctx, tokenValue, BeReplaced); err != nil {
				return err
			}

			// 2.4、发布事件：xx 账号的 xx 客户端注销了
			t.config.listener.DoReplaced(t.config.LoginType, loginId, tokenValue)
		}
	}
	return nil
}

// IsLogin 判断给定token是否登录
//
//	@receiver t
//	@param ctx context.Context
//	@param tokenValue string token值
//	@return bool 是否登录
func (t *token) IsLogin(ctx context.Context, tokenValue string) (bool, error) {
	defaultNil, err := t.getLoginIdDefaultNil(ctx, tokenValue)
	if err != nil {
		return false, err
	}
	if defaultNil == nil {
		return false, nil
	}
	return true, nil
}

// IsLoginByLoginId 根据登录id判断该用户是否已经登录
//
//	@receiver t
//	@param ctx context.Context
//	@param loginId any 登录id，推荐使用int,int64,string
//	@return bool 是否登录
//	@return error 是否有错
func (t *token) IsLoginByLoginId(ctx context.Context, loginId any) (bool, error) {
	tokenValues, err := t.getTokenValueByLoginId(ctx, loginId, "")
	if err != nil {
		return false, err
	}
	return len(tokenValues) > 0, nil
}

// GetLoginIdAsString 获取登录账号字符串
//
//	@receiver t
//	@param ctx context.Context
//	@param tokenValue string
//	@return string
//	@return error
//	@player
func (t *token) GetLoginIdAsString(ctx context.Context, tokenValue string) (string, error) {
	return t.getLoginIdNotHandle(ctx, tokenValue)
}

// GetLoginId 获取登录账号
//
//	@receiver t
//	@param ctx context.Context
//	@param tokenValue string 登录的token值
//	@return any 登录id
//	@return error 是否有错
func (t *token) GetLoginId(ctx context.Context, tokenValue string) (any, error) {
	// 1、查询登录id
	loginId, err := t.getLoginIdDefaultNil(ctx, tokenValue)
	if err != nil {
		return nil, err
	}
	// 2、如果loginId不存在，则表示获取了无效的token
	if loginId == nil {
		return nil, NewNotLoginError(InvalidToken, t.config.LoginType, InvalidTokenMessage, tokenValue)
	}
	// 3、如果是token已过期
	if convertor.ToString(loginId) == TimeoutToken {
		return nil, NewNotLoginError(TimeoutToken, t.config.LoginType, TimeoutTokenMessage, tokenValue)
	}
	// 4、如果是token被踢下线
	if convertor.ToString(loginId) == BeReplaced {
		return nil, NewNotLoginError(BeReplaced, t.config.LoginType, BeReplacedMessage, tokenValue)
	}
	// 5、如果是被踢下线
	if convertor.ToString(loginId) == KickOut {
		return nil, NewNotLoginError(KickOut, t.config.LoginType, KickOutMessage, tokenValue)
	}
	// 6、检查此 token 的最后活跃时间是否已经超过了 active-timeout 的限制，如果是则代表其已被冻结，需要抛出：token 已被冻结
	if t.isOpenCheckActiveTimeout() {
		if err := t.checkActiveTimeout(tokenValue); err != nil {
			return nil, err
		}
		// 如果配置了续签
		if t.config.AutoRenew {
			if err := t.updateLastActiveToNow(ctx, tokenValue); err != nil {
				return nil, err
			}
		}
	}
	return loginId, nil
}

// GetLoginIdAsInt 获取登录账号数字
//
//	@receiver t
//	@param ctx context.Context
//	@param tokenValue string 登录token值
//	@return int64 转换后的数字id，如果没有则为0
//	@return error 是否有错
func (t *token) GetLoginIdAsInt(ctx context.Context, tokenValue string) (int64, error) {
	loginId, err := t.getLoginIdDefaultNil(ctx, tokenValue)
	if err != nil {
		return 0, err
	}
	if loginId == nil {
		return 0, nil
	}
	return convertor.ToInt(loginId)
}

// Disable 封禁账号
//
//	@receiver t
//	@param ctx context.Context
//	@param loginId any
//	@param service string
//	@param timeout int64
//	@return error
//	@player
func (t *token) Disable(ctx context.Context, loginId any, service string, timeout int64) error {
	return t.DisableLevel(ctx, loginId, service, DefaultDisableLevel, timeout)
}

// DisableLevel 封禁：指定账号的指定服务，并指定封禁等级
//
//	 @receiver t
//	 @param ctx context.Context
//		@param loginId any 指定账号id
//		@param service string 指定封禁服务
//		@param level int 指定封禁等级
//		@param timeout int64 封禁时间, 单位: 秒 （-1=永久封禁）
//		@return error
func (t *token) DisableLevel(ctx context.Context, loginId any, service string, level int, timeout int64) error {
	if len(convertor.ToString(loginId)) == 0 {
		return errors.New("loginId is empty")
	}
	if len(service) == 0 {
		return errors.New("service is empty")
	}
	if level < MinDisableLevel {
		return errors.New("level is less than MinDisableLevel")
	}
	// 打上封印标记
	if err := t.config.repository.Set(ctx, t.splicingKeyDisable(loginId, service), convertor.ToString(level), time.Duration(timeout)*time.Second); err != nil {
		return err
	}
	// 发布事件
	t.config.listener.DoDisable(t.config.LoginType, loginId, service, level, timeout)
	return nil
}

// checkLoginArgs 检测登录参数
//
//	@receiver t
//	@param loginId any 登录的id
//	@param config ...LoginOption 此次登录的额外参数
//	@return error
func (t *token) checkLoginArgs(loginId any, opts LoginOptions) error {
	loginIdStr := convertor.ToString(loginId)
	// 1. 检测loginId是否为空
	if len(loginIdStr) == 0 {
		return errors.New("empty loginId")
	}
	// 2.登录id不能是异常标记值
	if set.FromSlice(AbnormalList).Contain(loginIdStr) {
		return fmt.Errorf("invalid loginId: %v", loginId)
	}
	// 3.检测loginId是否是数字或者字符串
	if !t.isBasicType(loginId) {
		t.config.logger.Warn("loginId 应该为简单类型，例如：string | int | int64")
	}
	// 4.如果全局配置未启动动态 activeTimeout 功能，但是此次登录却传入了 activeTimeout 参数，那么就打印警告信息
	if !t.config.DynamicActiveTimeout && opts.activeTimeout > 0 {
		t.config.logger.Warn("当前全局配置未开启动态 activeTimeout 功能，传入的 activeTimeout 参数将被忽略")
	}
	return nil
}

// distUsableToken 分配token
//
//	@receiver t
//	@param ctx context.Context
//	@param loginId any
//	@param opts LoginOptions
//	@return string
//	@return error
//	@player
func (t *token) distUsableToken(ctx context.Context, loginId any, opts LoginOptions) (string, error) {
	// 1、获取全局配置的 isConcurrent 参数
	//    如果配置为：不允许一个账号多地同时登录，则需要先将这个账号的历史登录会话标记为：被顶下线
	if !t.config.IsConcurrent {
		err := t.Replaced(ctx, loginId, opts.GetDevice())
		if err != nil {
			return "", err
		}
	}
	// 2、如果调用者预定了要生成的 token，则直接返回这个预定的值，框架无需再操心了
	if len(opts.GetToken()) > 0 {
		return opts.GetToken(), nil
	}
	// 3、只有在配置了 [ 允许一个账号多地同时登录 ] 时，才尝试复用旧 token，这样可以避免不必要的查询，节省开销
	if t.config.IsConcurrent {
		if t.config.IsShare {
			tokenValue, err := t.getTokenValueByLoginId(ctx, loginId, opts.GetDevice())
			if err != nil {
				return "", err
			}
			if len(tokenValue) > 0 {
				return tokenValue, nil
			}
		}
	}
	// 4、如果代码走到此处，说明未能成功复用旧 token，需要根据算法新建 token
	return t.config.generateUniqueToken(
		"token",
		t.config.MaxTryCount,
		func() string {
			return t.createTokenValue(loginId, opts.GetDevice(), opts.GetTimeout(), opts.GetExtraData())
		},
		func(value string) (bool, error) {
			id, err := t.getLoginIdNotHandle(ctx, value)
			if err != nil {
				return false, err
			}
			return len(id) > 0, nil
		},
	)
}

// getLoginIdDefaultNil 获取登录账号
//
//	 @receiver t
//	 @param ctx context.Context
//		@param tokenValue string toke值
//		@return any 登录账号，如果没有，返回nil
//		@return error 错误信息
func (t *token) getLoginIdDefaultNil(ctx context.Context, tokenValue string) (any, error) {
	// 1、如果token为空则直接返回false
	if len(tokenValue) == 0 {
		return nil, nil
	}
	// 2、根据 token 找到对应的 loginId，如果 loginId 为 null 或者属于异常标记里面，均视为未登录, 统一返回 null
	loginId, err := t.getLoginIdNotHandle(ctx, tokenValue)
	if err != nil {
		return nil, err
	}
	if !t.isValidLoginId(loginId) {
		return nil, nil
	}
	// 3、如果 token 已被冻结，也返回 null
	timeout, err := t.getTokenActiveTimeoutByToken(tokenValue)
	if err != nil {
		return nil, err
	}
	if timeout == NotValueExpire {
		return nil, nil
	}
	// 4、操作过程中没有错误，并且是合格的账号
	return loginId, nil
}

// logoutByMaxLoginCount 如果指定账号 id、设备类型的登录客户端已经超过了指定数量，则按照登录时间顺序，把最开始登录的给注销掉
//
//	@receiver t
//	@param ctx context.Context
//	@param id any 账号id
//	@param s Session 此账号的 Account-Session 对象，可填写 null，框架将自动获取
//	@param null any 设备类型（填 null 代表注销此账号所有设备类型的登录）
//	@param count int 最大登录数量，超过此数量的将被注销
//	@return error
func (t *token) logoutByMaxLoginCount(ctx context.Context, loginId any, ss *session, device string, count int) (err error) {
	// 1、如果调用者提供的  Account-Session 对象为空，则我们先手动获取一下
	if ss == nil {
		ss, err = t.getSessionByLoginId(ctx, loginId, false)
		if err != nil {
			return
		}
		if ss == nil {
			return
		}
	}
	// 2、获取指定账号指定设备类型下的所有登录客户端
	signList := ss.getTokenSignListByDevice(device)

	// 3、按照登录时间倒叙，超过 maxLoginCount 数量的，全部注销掉
	for i := 0; i < len(signList)-count; i++ {
		tokenValue := signList[i].Value
		// 3.1 从account-session上移除token签名
		err = ss.removeTokenSign(ctx, tokenValue)
		if err != nil {
			return err
		}
		// 3.2 清除这个 token 的最后活跃时间记录
		if t.isOpenCheckActiveTimeout() {
			err = t.clearLastActive(tokenValue)
			if err != nil {
				return err
			}
		}
		// 3.3 删除 token - id 映射
		err = t.deleteTokenToIdMapping(tokenValue)
		if err != nil {
			return err
		}
		// 3.4、清除这个 token 的 Token-Session 对象
		err = t.deleteTokenSession(ctx, tokenValue)
		if err != nil {
			return err
		}
		// 3.5、发布事件：xx 账号的 xx 客户端注销了
		t.config.listener.DoLogout(t.config.LoginType, loginId, tokenValue)
	}

	// 4、如果客户端的登录账号数量为0，则直接清除account-session
	return ss.logoutByTokenSignCountIsZero(ctx)
}

// isBasicType 判断是否是string,int...
//
//	@receiver t
//	@param val any
//	@return bool
func (t *token) isBasicType(val any) bool {
	switch val.(type) {
	case int, int64, int32, int16, int8, string:
		return true
	}
	return false
}

// isValidLoginId 检查登录账号是否是有效账号
//
//	@receiver t
//	@param loginId any
//	@return bool
func (t *token) isValidLoginId(loginId any) bool {
	return len(convertor.ToString(loginId)) > 0 && !set.New(AbnormalList...).Contain(convertor.ToString(loginId))
}

// isOpenCheckActiveTimeout 判断是否开启了活跃度检测
//
//	@receiver t
//	@return bool
func (t *token) isOpenCheckActiveTimeout() bool {
	if t.config.DynamicActiveTimeout || t.config.ActiveTimeout != NeverExpire {
		return true
	}
	return false
}

// clearLastActive 清除最后活跃时间
//
//	@receiver t
//	@param tokenValue string token值
//	@return error
func (t *token) clearLastActive(tokenValue string) error {
	return t.config.repository.Delete(context.Background(), t.splicingKeyLastActiveTime(tokenValue))
}

// deleteTokenToIdMapping 删除 token - id 的映射
//
//	@receiver t
//	@param tokenValue string token值
//	@return error 错误信息
func (t *token) deleteTokenToIdMapping(tokenValue string) error {
	return t.config.repository.Delete(context.Background(), t.splicingKeyTokenValue(tokenValue))
}

// getTokenActiveTimeoutByToken 获取指定 token 剩余活跃有效期：这个 token 距离被冻结还剩多少时间（单位: 秒，返回 -1 代表永不冻结，-2 代表没有这个值或 token 已被冻结了）
//
//	@receiver t
//	@param tokenValue string token值
//	@return int 剩余时长（单位：秒）
func (t *token) getTokenActiveTimeoutByToken(tokenValue string) (int64, error) {
	// 1、如果全局配置了永不冻结, 则返回 -1
	if !t.isOpenCheckActiveTimeout() {
		return NeverExpire, nil
	}
	// 2、如果提供的 token 为 null，则返回 -2
	if len(tokenValue) == 0 {
		return NotValueExpire, nil
	}
	// --------开始查询

	// 1、先获取这个 token 的最后活跃时间，13位时间戳
	key := t.splicingKeyLastActiveTime(tokenValue)
	var lastActiveTimeStr = ""
	err := t.config.repository.Get(context.Background(), key, &lastActiveTimeStr)
	if err != nil || len(lastActiveTimeStr) == 0 {
		return NotValueExpire, err
	}
	// 2、计算最后活跃时间 距离 此时此刻 的时间差
	//    计算公式为: (当前时间 - 最后活跃时间) / 1000
	timeValue := newActiveTimeValue(lastActiveTimeStr)
	lastActiveTime := timeValue.getCurrentTime().UnixMilli()
	// 时间差
	timeDiff := (time.Now().UnixMilli() - lastActiveTime) / 1000
	// 该 token 允许的时间差
	allowTimeDiff := t.getActiveTimeAllowTimeDiffOrGlobalConfig(timeValue)
	if *allowTimeDiff == NeverExpire {
		// 如果允许的时间差为 -1 ，则代表永不冻结，此处需要立即返回 -1 ，无需后续计算
		return NeverExpire, nil
	}
	// 3、校验这个时间差是否超过了允许的值
	//    计算公式为: 允许的最大时间差 - 实际时间差，判断是否 < 0， 如果是则代表已经被冻结 ，返回-2
	activeTimeout := *allowTimeDiff - timeDiff
	if activeTimeout < 0 {
		return NotValueExpire, nil
	}
	// 否则代表没冻结，返回剩余有效时间
	return activeTimeout, nil
}

// checkActiveTimeout 检查指定 token 是否已被冻结，如果是则返回错误
//
//	@receiver t
//	@param tokenValue string
//	@return error
func (t *token) checkActiveTimeout(tokenValue string) error {
	// 1、获取这个 token 的剩余活跃有效期
	activeTimeout, err := t.getTokenActiveTimeoutByToken(tokenValue)
	if err != nil {
		return err
	}
	// 2、值为 -1 代表此 token 已经被设置永不冻结，无须继续验证
	if activeTimeout == NeverExpire {
		return nil
	}
	// 3、值为 -2 代表已被冻结，此时需要抛出异常
	if activeTimeout == NotValueExpire {
		return NewNotLoginError(FreezeToken, t.config.LoginType, FreezeTokenMessage, tokenValue)
	}
	return nil
}

// getActiveTimeAllowTimeDiffOrGlobalConfig
//
//	@receiver t
//	@param value *activeTimeValue
//	@return *int64
func (t *token) getActiveTimeAllowTimeDiffOrGlobalConfig(value *activeTimeValue) *int64 {
	activeTime := value.getActiveTimeout()
	if activeTime == nil {
		return &t.config.ActiveTimeout
	}
	return activeTime
}

// getActiveTimeAllowTimeDiff 获取允许的时间差值
//
//	@receiver t
//	@param value activeTimeValue
//	@return *int64
func (t *token) getActiveTimeAllowTimeDiff(value *activeTimeValue) *int64 {
	if !t.config.DynamicActiveTimeout {
		return nil
	}
	return value.getActiveTimeout()
}

// deleteTokenSession 删除指定token的token-session
//
//	@receiver t
//	@param ctx context.Context
//	@param value string token值
//	@return error 错误信息
func (t *token) deleteTokenSession(ctx context.Context, value string) error {
	return t.config.repository.Delete(ctx, t.splicingKeyTokenSession(value))
}

// getTokenValueByLoginId 获取token值根据登录编号和设备
//
//	 @receiver t
//	 @param ctx context.Context
//		@param loginId any 登录编号
//		@param device string 设备信息
//		@return string token值
//		@return error 是否有错
func (t *token) getTokenValueByLoginId(ctx context.Context, loginId any, device string) (string, error) {
	list, err := t.getTokenSignListByLoginId(ctx, loginId, device)
	if err != nil {
		return "", err
	}
	if len(list) == 0 {
		return "", nil
	}
	return list[len(list)-1], nil
}

// getTokenSignListByLoginId 根据登录id获取用户的所有登录签名列表
//
//	@receiver t
//	@param loginId any 登录编号
//	@param device string 设备信息，为空时查询全部
//	@return signList 所有的签名列表
//	@return err 是否有错
func (t *token) getTokenSignListByLoginId(ctx context.Context, loginId any, device string) (signList []string, err error) {
	ss, err := t.getSessionByLoginId(ctx, loginId, false)
	if err != nil {
		return []string{}, err
	}
	// 如果没有获取到session
	if ss == nil {
		return []string{}, nil
	}
	return ss.getTokenValueListByDevice(device), nil
}

// GetSessionByLoginId  获取指定账号 id 的 Account-Session
//
//	@receiver t
//	@param loginId any 账号id
//	@param isCrete bool 如果不存在是否创建一个
//	@return Session session结构体
//	@return error 是否有错
func (t *token) GetSessionByLoginId(ctx context.Context, loginId any, isCrete bool) (Session, error) {
	return t.getSessionByLoginId(ctx, loginId, isCrete)
}

// GetSessionByLoginIdDefault 获取指定用户id的 Account-session，如果没有则创建一个
//
//	@receiver t
//	@param ctx context.Context
//	@param loginId any 登录id
//	@return Session session结构
//	@return error 是否有错
func (t *token) GetSessionByLoginIdDefault(ctx context.Context, loginId any) (Session, error) {
	return t.getSessionByLoginId(ctx, loginId, true)
}

// getSessionByLoginId 获取指定账号 id 的 Account-Session
//
//	@receiver t
//	@param ctx context.Context 上下文
//	@param loginId any 账号id
//	@param isCreate bool 如果不存在是否创建一个
//	@return Session
//	@return error
func (t *token) getSessionByLoginId(ctx context.Context, loginId any, isCreate bool) (*session, error) {
	if len(convertor.ToString(loginId)) == 0 {
		return nil, errors.New("Account-Session 获取失败：loginId 不能为空")
	}
	return t.getSessionBySessionId(
		ctx,
		t.splicingKeySession(loginId),
		isCreate,
		func(s *session) error {
			s.SessionType = AccountSessionType
			s.LoginType = t.config.LoginType
			s.LoginId = loginId
			return nil
		},
	)
}

// getSessionBySessionId 获取指定key的session
//
//	@receiver t
//	@param sessionId string
//	@param isCreate bool
//	@param appendOperation ...func(s session) error
//	@return Session session实例
//	@return error 是否有错
func (t *token) getSessionBySessionId(ctx context.Context, sessionId string, isCreate bool, appendOperation ...func(s *session) error) (*session, error) {
	// 如果提供的 sessionId 为 null，则直接返回 null
	if len(sessionId) == 0 {
		return nil, errors.New("sessionId 不能为空")
	}
	var operation func(s *session) error
	if len(appendOperation) > 0 {
		operation = appendOperation[0]
	}
	var ss = &session{}
	var err error
	if err := t.config.repository.Get(ctx, sessionId, ss); err != nil {
		return nil, err
	}
	// 如果没有
	if ss.ID == "" && isCreate {
		ss, err = t.config.createSessionFunction(sessionId, t.config.repository)
		if err != nil {
			return nil, err
		}
		if operation != nil {
			if err := operation(ss); err != nil {
				return nil, err
			}
		}
		if err := t.config.repository.Set(ctx, sessionId, ss, time.Duration(t.config.Timeout)*time.Second); err != nil {
			return nil, err
		}
	} else if ss.ID != "" && ss.repo == nil {
		ss.repo = t.config.repository
	}
	// 如果还是没有则返回nil
	if ss.ID == "" {
		return nil, nil
	}
	return ss, nil
}

// GetTokenSessionByTokenValue 获取指定 token 的 Token-Session，如果该 SaSession 尚未创建，isCreate代表是否新建并返回
//
//	@receiver t
//	@param tokenValue string 指定token值
//	@param isCreate bool 如果没有，是否创建
//	@return Session session结构
//	@return error 是否有错
func (t *token) GetTokenSessionByTokenValue(ctx context.Context, tokenValue string, isCreate bool) (Session, error) {
	if len(tokenValue) == 0 {
		return nil, errors.New("Token-Session 获取失败：token 不能为空")
	}
	return t.getSessionBySessionId(ctx, t.splicingKeyTokenSession(tokenValue), isCreate, func(s *session) error {
		s.SessionType = TokenSessionType
		s.LoginType = t.config.LoginType
		s.Token = tokenValue
		return nil
	})
}

// createTokenValue 创建token值的方法实现
//
//	@receiver t
//	@param loginId any 登录用户
//	@param device string 登录设备
//	@param timeout int64 有效期（单位：秒）
//	@param extraData map[string]interface{} 额外信息
//	@return string token值
func (t *token) createTokenValue(loginId any, device string, timeout int64, extraData map[string]interface{}) string {
	return t.config.createTokenFunction(loginId, t.config.LoginType, t.config.Style)
}

// getLoginIdNotHandle 获取指定token对应的id
//
//	@receiver t
//	@param ct context.Context
//	@param tokenValue string
//	@return string
//	@return error
//	@player
func (t *token) getLoginIdNotHandle(ctx context.Context, tokenValue string) (string, error) {
	var loginId string
	if err := t.config.repository.Get(ctx, t.splicingKeyTokenValue(tokenValue), &loginId); err != nil {
		return "", err
	}
	return loginId, nil
}

// saveTokenToIdMapping 保存token与id映射关系
//
//	@receiver t
//	@param ctx context.Context
//	@param tokenValue string token值
//	@param loginId any 登录Id
//	@param timeout int64 过期时间
//	@return string
func (t *token) saveTokenToIdMapping(ctx context.Context, tokenValue string, loginId any, timeout int64) error {
	return t.config.repository.Set(ctx, t.splicingKeyTokenValue(tokenValue), convertor.ToString(loginId), time.Duration(timeout)*time.Second)
}

// updateTokenToIdMapping 更改 token - id 映射关系
//
//	@receiver t
//	@param ctx context.Context
//	@param value string token值
//	@param loginId any 登陆账号
//	@return error
func (t *token) updateTokenToIdMapping(ctx context.Context, tokenValue string, loginId any) error {
	if len(convertor.ToString(loginId)) == 0 {
		return errors.New("loginId 不能为空")
	}
	return t.config.repository.Update(ctx, tokenValue, convertor.ToString(loginId))
}

// setLastActiveToNow 设置token的最后活跃时间为当前时间
//
//	@receiver t
//	@param ctx context.Context
//	@param tokenValue string token值
//	@param activeTimeout int64 token的最低活跃频率，如果为0，则使用全局配置的 activeTimeout 值
//	@param timeout int64 保存数据时使用的ttl数值，如果为0，则使用全局配置的 timeout 值
//	@return error 是否有错
func (t *token) setLastActiveToNow(ctx context.Context, tokenValue string, activeTimeout int64, timeout int64) error {
	if timeout == 0 {
		timeout = t.config.Timeout
	}
	// 将此 token 的 [ 最后活跃时间 ] 标记为当前时间戳
	key := t.splicingKeyLastActiveTime(tokenValue)
	val := convertor.ToString(time.Now().UnixMilli()) // 当前时间duration
	if t.config.DynamicActiveTimeout && activeTimeout != 0 {
		val += "," + convertor.ToString(activeTimeout)
	}
	return t.config.repository.Set(ctx, key, val, time.Duration(timeout)*time.Second)
}

// getTokenUseActiveTimeout 获取指定 token 在缓存中的 activeTimeout 值，如果不存在则返回 nil
//
//	@receiver t
//	@param tokenValue string 指定 token
//	@return int64
func (t *token) getTokenUseActiveTimeout(ctx context.Context, tokenValue string) (*int64, error) {
	if !t.config.DynamicActiveTimeout {
		return nil, nil
	}
	key := t.splicingKeyLastActiveTime(tokenValue)
	var value string
	if err := t.config.repository.Get(ctx, key, &value); err != nil {
		return nil, err
	}
	storeValue := newActiveTimeValue(value)
	return storeValue.getActiveTimeout(), nil
}

// updateLastActiveToNow 修改指定token的活跃时间为当前时间
//
//	@receiver t
//	@param tokenValue string 指定token值
//	@return error 是否有错
func (t *token) updateLastActiveToNow(ctx context.Context, tokenValue string) error {
	key := t.splicingKeyLastActiveTime(tokenValue)
	timeout, err := t.getTokenUseActiveTimeout(ctx, tokenValue)
	if err != nil {
		return err
	}
	now := time.Now()
	value := newActiveTimeValueWithValue(&now, timeout).Fmt()
	return t.config.repository.Update(ctx, key, value)
}

// splicingKeyTokenValue  拼接： 在保存 token - id 映射关系时，使用的key
//
//	@receiver t
//	@param tokenValue string token值
//	@return string
func (t *token) splicingKeyTokenValue(tokenValue string) string {
	return t.config.TokenName + ":" + t.config.LoginType + ":token:" + tokenValue
}

// splicingKeySession 保存session时使用的key
//
//	@receiver t
//	@param loginId any
func (t *token) splicingKeySession(loginId any) string {
	return t.config.TokenName + ":" + t.config.LoginType + ":session:" + convertor.ToString(loginId)
}

// splicingKeyTokenSession 拼装：在保存 token-session时使用的key
//
//	@receiver t
//	@param tokenValue string
//	@return string
func (t *token) splicingKeyTokenSession(tokenValue string) string {
	return t.config.TokenName + ":" + t.config.LoginType + ":token-session:" + tokenValue
}

// splicingKeyLastActiveTime 拼接：在保存 token - lastActiveTime 映射关系时，使用的key
//
//	@receiver t
//	@param tokenValue string token值
//	@return string
func (t *token) splicingKeyLastActiveTime(tokenValue string) string {
	return t.config.TokenName + ":" + t.config.LoginType + ":last-active:" + tokenValue
}

// splicingKeyDisable 拼接key ，存储封禁信息的key
//
//	@receiver t
//	@param loginId any
//	@param service string
//	@return string
func (t *token) splicingKeyDisable(loginId any, service string) string {
	return t.config.TokenName + ":" + t.config.LoginType + ":disable:" + service + ":" + convertor.ToString(loginId)
}
