package gadmin

import (
	"errors"
	"html/template"
	"iter"
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
	// asc[default] or desc
	Desc   bool   `form:"desc,omitempty"`
	Search string `form:"search,omitempty"`

	// flt0_35=2024-10-28&flt2_27=Harry&flt3_0=1
	// filters []
	args []string

	//
	default_page_size int
	num_pages         int
}

func DefaultQuery() *Query {
	return &Query{
		Page:              0,
		PageSize:          10,
		default_page_size: 10,
	}
}

func (q *Query) withArgs(args ...any) *Query {
	if q.args == nil {
		q.args = pyslice(args...)
	} else {
		q.args = append(q.args, pyslice(args)...)
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
	encoder := must(form.NewEncoder())
	encoder.RegisterCustomTypeFunc(encodeBool, true)

	uv := must(encoder.Encode(q))
	for i := 0; i < len(q.args); i += 2 {
		uv.Add(q.args[i], q.args[i+1])
	}
	return uv
}

// flask-admin Encode true to "1", false to "0" in `form`
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

// generate pager or json
type Result struct {
	*Query

	Total int64
	Rows  []Row
	Error error
}

func (r *Result) NumPages() int {
	page_size := lo.Ternary(r.PageSize != 0, r.PageSize, r.default_page_size)
	return int(1 + (r.Total-1)/int64(page_size))
}

func (r *Result) PageRange() iter.Seq[int] {
	n := r.NumPages()

	low, up := r.Page-3, r.Page+4
	if low < 0 {
		up = up - low
	}
	if up > n {
		low = low - up + n
	}

	low = max(low, 0)
	up = min(up, n)

	return func(yield func(int) bool) {
		for i := low; i < up; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func (r *Result) Html() template.HTML {
	return ""
}
