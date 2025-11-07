package gadmin

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPager(t *testing.T) {
	is := assert.New(t)

	r := Result{Query: &Query{Page: 1, PageSize: 10}, Total: 100}
	is.Equal([]int{0, 1, 2, 3, 4, 5, 6}, slices.Collect(r.PageRange()))

	r = Result{Query: &Query{Page: 5, PageSize: 10}, Total: 100}
	is.Equal([]int{2, 3, 4, 5, 6, 7, 8}, slices.Collect(r.PageRange()))

	r = Result{Query: &Query{Page: 9, PageSize: 10}, Total: 100}
	is.Equal([]int{3, 4, 5, 6, 7, 8, 9}, slices.Collect(r.PageRange()))

	r = Result{Query: &Query{Page: 0, PageSize: 10}, Total: 4}
	is.Equal([]int{0}, slices.Collect(r.PageRange()))

	r = Result{Query: &Query{Page: 0, PageSize: 10}, Total: 11}
	is.Equal([]int{0, 1}, slices.Collect(r.PageRange()))
	r = Result{Query: &Query{Page: 1, PageSize: 10}, Total: 11}
	is.Equal([]int{0, 1}, slices.Collect(r.PageRange()))

	r = Result{Query: &Query{Page: 0, PageSize: 1}, Total: 1}
	is.Equal([]int{0}, slices.Collect(r.PageRange()))
}
