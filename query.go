package gadmin

import (
	"errors"
	"html/template"
	"net/url"

	"github.com/go-playground/form/v4"
	"github.com/samber/lo"
)

type Query struct {
	// 0 based with omit
	Page     int `form:"page,omitempty"`
	PageSize int `form:"page_size,omitempty"`
	// column index: 0,1,... maybe `null.String` is better
	Sort string `form:"sort,omitempty"`
	// desc or asc, default is asc
	Desc   bool   `form:"desc,omitempty"`
	Search string `form:"search,omitempty"`
	// flt0_35=2024-10-28&flt2_27=Harry&flt3_0=1
	// filters []

	args []string

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
		q.args = intoStringSlice(args...)
	} else {
		q.args = append(q.args, intoStringSlice(args)...)
	}
	return q
}

func (q *Query) Get(arg string) string {
	for i := 0; i < len(q.args); i += 2 {
		if q.args[i] == arg {
			return q.args[i+1]
		}
	}
	return ""
}

func (q *Query) toValues() url.Values {
	encoder := must[*form.Encoder](form.NewEncoder())
	encoder.RegisterCustomTypeFunc(encodeBool, true)

	uv := must[url.Values](encoder.Encode(q))
	for i := 0; i < len(q.args); i += 2 {
		uv.Add(q.args[i], q.args[i+1])
	}
	return uv
}

// Encode true to "1", false to "0" in `form`
func encodeBool(x any) ([]string, error) {
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

// <Previous 1, 2 ... 134 Next>
// <ul class="pagination">
//
//	<li class="page-item">
//	    <a href="{{ .page }}">&lt;</a>
//	</li>
//
// </ul>
func (Q *Query) SimplePage() template.HTML {
	return template.HTML("")
}
