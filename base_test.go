package gadmin

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/go-playground/form/v4"
	"github.com/stretchr/testify/assert"
)

func TestFirstOrEmpty(t *testing.T) {
	is := assert.New(t)

	is.Equal("a", firstOr([]string{"a", "b"}, "c"))
	is.Equal("c", firstOr([]string{}, "c"))
	is.Equal(0, firstOr([]int{}, 0))
}

func TestPairToQuery(t *testing.T) {
	is := assert.New(t)

	is.Equal("a=1", pairToQuery("a", 1).Encode())
	is.Equal("a=1&a=2", pairToQuery("a", "1", "a", "2").Encode())
	is.Equal("a=+", pairToQuery("a", " ").Encode())

	// abnormal input
	is.Equal("", pairToQuery("a").Encode())
	is.Equal("a=b", pairToQuery("a", "b", "c").Encode())
}

func TestBaseMust(t *testing.T) {
	is := assert.New(t)

	fr := func() (int, error) { return 42, nil }
	is.Equal(42, must[int](fr()))

	fe := func() (int, error) { return 1, errors.New("sth. wrong") }
	is.Panics(func() { must[int](fe()) })

	ft := func() (int, bool) { return 42, true }
	is.Equal(42, must[int](ft()))

	ff := func() (int, bool) { return 42, false }
	is.Panics(func() { must[int](ff()) })
}

func TestBaseConvert(t *testing.T) {
	is := assert.New(t)

	m := map[string]any{
		"name":    "John Doe",
		"age":     30,
		"active":  true,
		"numbers": []int{1, 2, 3}, // This will be skipped as it's not a string
	}

	uv := anyMapToQuery(m)
	is.Equal("active=true&age=30&name=John+Doe&numbers=%5B1+2+3%5D", uv.Encode())
}

func TestStd(t *testing.T) {
	is := assert.New(t)

	is.Equal([]string{"a", "b"}, strings.SplitN("a.b", ".", 2))
	is.Equal([]string{"a", "b.c"}, strings.SplitN("a.b.c", ".", 2))
	is.Equal([]string{"", "b"}, strings.SplitN(".b", ".", 2))

	e := form.NewEncoder()
	is.Equal("%5Ba%5D=1", must[url.Values](e.Encode(map[string]any{
		"a": 1,
	})).Encode())

	type list_form struct {
		list_form_pk any
		CamelCase    string
	}
	is.Equal("CamelCase=abc", must[url.Values](e.Encode(list_form{
		list_form_pk: "33",
		CamelCase:    "abc",
	})).Encode())
}

func TestQuery(t *testing.T) {
	is := assert.New(t)

	q1 := Query{Page: 2}
	uv1, err1 := form.NewEncoder().Encode(q1)
	is.Nil(err1)
	is.Equal("page=2", uv1.Encode())

	var q2 Query
	form.NewDecoder().Decode(&q2, url.Values{
		"page": []string{"2"},
		"desc": []string{"1"},
	})
	is.Equal(2, q2.Page)
	is.Equal(true, q2.Desc)
	is.Equal("", q2.Sort)
	is.Equal(0, q2.PageSize)

	is.Equal("desc=1&page=2", q2.toValues().Encode())
}
