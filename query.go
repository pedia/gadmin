package gadmin

import (
	"errors"
	"net/url"

	"github.com/go-playground/form/v4"
	"github.com/samber/lo"
)

type Query struct {
	// 1 based
	Page     int `form:"page,omitempty"`
	PageSize int `form:"page_size,omitempty"`
	// column index: 0,1,... maybe `null.String` is better
	Sort string `form:"sort,omitempty"`
	// desc or asc, default is asc
	Desc   bool   `form:"desc,omitempty"`
	Search string `form:"search,omitempty"`
	// flt0_35=2024-10-28&flt2_27=Harry&flt3_0=1
	// filters []

	args []any

	//
	default_page_size int
	num_pages         int
	total             int
}

func (q *Query) setTotal(total int64) {
	q.total = int(total)
	// num_pages := math.Ceil(float64(total) / float64(q.limit))
	page_size := lo.Ternary(q.PageSize != 0, q.PageSize, q.default_page_size)
	q.num_pages = int(1 + (total-1)/int64(page_size))
}

func (q *Query) withArgs(args ...any) *Query {
	if q.args == nil {
		q.args = args
	} else {
		q.args = append(q.args, args...)
	}
	return q
}

func (q *Query) toValues() url.Values {
	encoder := must[*form.Encoder](form.NewEncoder())
	encoder.RegisterCustomTypeFunc(encodeBool, true)
	return must[url.Values](encoder.Encode(q))
}

// Encode true to "1", false to "0"
func encodeBool(x interface{}) ([]string, error) {
	v, ok := x.(bool)
	if !ok {
		return nil, errors.New("bad type conversion")
	}

	if v {
		return []string{"1"}, nil
	} else {
		return []string{"0"}, nil
	}
}
