package pagination

import (
	"github.com/go-fox/fox/codec"
	"github.com/go-fox/fox/codec/json"
	"strings"
)

// ToPagingParams 转换为PagingParams
func (x *PagingRequest) ToPagingParams() *PagingParams {
	var condition Condition
	if x.Query != nil {
		_ = codec.GetCodec(json.Name).Unmarshal([]byte(x.GetQuery()), &condition)
	}
	var orderByArray []string
	if len(x.GetOrderBy()) > 0 {
		orderByArray = strings.Split(x.GetOrderBy(), ",")
	}
	var fields []string
	if len(x.GetFields()) > 0 {
		fields = strings.Split(x.GetFields(), ",")
	}
	return &PagingParams{
		Page:     x.Page,
		Size:     x.Size,
		Query:    &condition,
		OrderBy:  orderByArray,
		NoPaging: x.NoPaging,
		Fields:   fields,
	}
}
