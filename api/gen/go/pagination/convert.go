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
	if query == "" {
		query = "{}"
	}
	if err := codec.GetCodec("json").Unmarshal([]byte(query), &condition); err != nil {
		return nil, err
	}
	var orderByArray []string
	orderBy := x.GetOrderBy()
	if len(x.GetOrderBy()) > 0 {
		orderByArray = strings.Split(orderBy, ",")
	}
	selectFields := strings.Split(x.GetFields(), ",")
	return &pagination.PagingRequest{
		Page:       x.GetPage(),
		Size:       x.GetSize(),
		OrderBy:    orderByArray,
		Pagination: !x.GetNoPaging(),
		Fields:     selectFields,
		Condition:  &condition,
	}, nil
}
