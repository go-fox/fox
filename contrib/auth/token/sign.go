package token

// Sign 签名
type Sign struct {
	Value  string         `json:"value"`  // token值
	Device string         `json:"device"` // 所属设备
	Extra  map[string]any `json:"extra"`  // 额外数据
}

// SignList 签名列表
type SignList []*Sign

// NewSign 构造函数
//
//	@param tokenValue string token值
//	@param device string 登录设备
//	@param extra any 挂载数据
//	@return *Sign
func NewSign(tokenValue, device string, extra map[string]any) *Sign {
	return &Sign{
		Value:  tokenValue,
		Device: device,
		Extra:  extra,
	}
}
