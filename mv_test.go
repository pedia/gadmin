package gadmin

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModelView(t *testing.T) {
	is := assert.New(t)

	v := NewModelView(Foo{})
	is.Equal([]string{"id", "name", "email", "age", "normal", "valid", "member_number", "birthday", "activated_at", "created_at", "updated_at", "decimal"}, v.column_list)
	is.Equal([]string{"id", "name", "email", "age", "normal", "valid", "member_number", "birthday", "activated_at", "created_at", "updated_at", "decimal"}, v.column_sortable_list)

	// query
	r1 := httptest.NewRequest("", "/admin/tag/?sort=0&desc=1&page_size=23&page=2", nil)
	q1 := v.queryFrom(r1)
	is.Equal("0", q1.Sort)
	is.Equal(true, q1.Desc)
	is.Equal(23, q1.PageSize)
	is.Equal(2, q1.Page)

	r2 := httptest.NewRequest("", "/admin/tag/?sort=1", nil)
	q2 := v.queryFrom(r2)
	is.Equal("1", q2.Sort)
	is.Equal(false, q2.Desc)
	is.Equal(0, q2.PageSize)
	is.Equal(0, q2.Page)

	r3 := httptest.NewRequest("", "/admin/tag/details?id=6&url=%2Fadmin%2Ftag%2F%3Fdesc%3D1%26sort%3D1", nil)
	q3 := v.queryFrom(r3)
	is.Equal("", q3.Sort)
	is.Equal(false, q3.Desc)
	is.Equal(0, q3.PageSize)
	is.Equal(0, q3.Page)
	is.Equal("6", q3.Get("id"))
	is.Equal("/admin/tag/?desc=1&sort=1", q3.Get("url"))
}
