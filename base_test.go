package gadmin

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBase(t *testing.T) {
	is := assert.New(t)

	is.Equal("a", firstOrEmpty("a", "b"))
	is.Equal("", firstOrEmpty[string]())
	is.Equal(0, firstOrEmpty[int]())
}

func TestPairToQuery(t *testing.T) {
	is := assert.New(t)

	is.Equal("a=1", pairToQueryString("a", "1"))
	is.Equal("a=1&a=2", pairToQueryString("a", "1", "a", "2"))
	is.Equal("a=+", pairToQueryString("a", " "))

	// abnormal input
	is.Equal("", pairToQueryString("a"))
	is.Equal("a=b", pairToQueryString("a", "b", "c"))
}

func TestBaseMust(t *testing.T) {
	is := assert.New(t)

	fr := func() (int, error) {
		return 42, nil
	}
	is.Equal(42, must[int](fr()))

	fe := func() (int, error) {
		return 1, errors.New("sth. wrong")
	}
	is.Panics(func() { must[int](fe()) })

	ft := func() (int, bool) {
		return 42, true
	}
	is.Equal(42, must[int](ft()))

	ff := func() (int, bool) {
		return 42, false
	}
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
}
