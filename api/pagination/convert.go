package pagination

import (
	"strings"

	"github.com/go-fox/go-utils/pagination"

	"github.com/go-fox/fox/codec"
)

// ToPagination 转换为分页请求
func (x *PagingRequest) ToPagination() (*pagination.PagingRequest, error) {
	query := x.GetQuery()
	var condition pagination.Condition
	if err := codec.GetCodec("json").Unmarshal([]byte(query), &x.Query); err != nil {
		return nil, err
	}
	orderBy := x.GetOrderBy()
	orderByArray := strings.Split(orderBy, ",")
	selectFields := strings.Split(x.GetFields(), ",")
	return &pagination.PagingRequest{
		Page:       x.GetPage(),
		Size:       x.GetSize(),
		OrderBy:    orderByArray,
		Pagination: !x.GetNoPaging(),
		Fields:     selectFields,
		Where: &pagination.Where{
			Conditions:      condition.Conditions,
			LogicalOperator: condition.LogicalOperator,
		},
	}, nil
}
