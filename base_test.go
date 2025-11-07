package gadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"text/template"

	"github.com/go-playground/form/v4"
	"github.com/stretchr/testify/assert"
	"gopkg.in/leonelquinteros/gotext.v1"
)

func TestBase(t *testing.T) {
	is := assert.New(t)

	is.Equal("a", firstOr([]string{"a", "b"}, "c"))
	is.Equal("c", firstOr([]string{}, "c"))
	is.Equal(0, firstOr([]int{}, 0))

	//
	is.Equal("a=1", pairsToQuery("a", 1).Encode())
	is.Equal("a=1&a=2", pairsToQuery("a", "1", "a", "2").Encode())
	is.Equal("a=+", pairsToQuery("a", " ").Encode())

	// abnormal input
	is.Equal("", pairsToQuery("a").Encode())
	is.Equal("a=b", pairsToQuery("a", "b", "c").Encode())

	is.Equal([]string{"1", "foo", "3.4", "1", "0", "0"}, pyslice(1, "foo", 3.4, true, 0, false))

	//
	is.Len(merge(map[int]int{}, nil), 0)
}

func TestBaseMust(t *testing.T) {
	is := assert.New(t)

	fr := func() (int, error) { return 42, nil }
	is.Equal(42, must(fr()))

	fe := func() (int, error) { return 1, errors.New("sth. wrong") }
	is.Panics(func() { must(fe()) })

	ft := func() (int, bool) { return 42, true }
	is.Equal(42, must(ft()))

	ff := func() (int, bool) { return 42, false }
	is.Panics(func() { must(ff()) })
}

func TestStd(t *testing.T) {
	is := assert.New(t)

	is.Equal([]string{"a", "b"}, strings.SplitN("a.b", ".", 2))
	is.Equal([]string{"a", "b.c"}, strings.SplitN("a.b.c", ".", 2))
	is.Equal([]string{"", "b"}, strings.SplitN(".b", ".", 2))

	e := form.NewEncoder()
	is.Equal("%5Ba%5D=1", must(e.Encode(map[string]any{
		"a": 1,
	})).Encode())

	type list_form struct {
		list_form_pk any
		CamelCase    string
	}
	is.Equal("CamelCase=abc", must(e.Encode(list_form{
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

func TestOnce(t *testing.T) {
	is := assert.New(t)
	oc := 0
	once := sync.OnceValue(func() int {
		sum := 0
		for i := 0; i < 100; i++ {
			sum += i
		}
		fmt.Println("Computed once:", sum)
		oc += 1
		return sum
	})
	for a := 0; a < 10; a++ {
		got := once()
		_ = got
		for j := 0; j < 10; j++ {
			got := once()
			is.Equal(4950, got)
		}
	}
	is.Equal(1, oc)
}

func TestText(t *testing.T) {
	is := assert.New(t)

	gotext.Configure("translations", "zh_Hant_TW", "admin")
	is.Equal("首頁", gotext.Get("Home"))
	is.Equal(`檔案 "foo" 已經存在。`, gotext.Get(`File "%s" already exists.`, "foo"))
}

func TestBufferWriter(t *testing.T) {
	is := assert.New(t)

	called := false
	bf := func(w http.ResponseWriter) {
		w.Header().Add("Hello", "World")
		called = true
	}
	w0 := httptest.NewRecorder()

	w := NewBufferWriter(w0, bf)
	w.Write([]byte("body"))

	is.False(called)

	w.(http.Flusher).Flush()

	is.Equal("World", w.Header().Get("Hello"))
	is.Equal("body", w0.Body.String())
	is.True(called)
}

type base struct{}

type derived struct {
	*base
}

func TestBasePtr(t *testing.T) {
	is := assert.New(t)

	d := &derived{}
	bp, ok := any(d).(*base)
	is.False(ok)
	is.Nil(bp)
}

func ExampleTemplate() {
	base := template.Must(template.New("base").Parse(`{{.}}base{{block "body" .}}{{end}}{{println}}`))
	base.Execute(os.Stdout, 1)

	d1 := template.Must(template.Must(base.Clone()).Parse(`{{define "body"}}d1{{end}}`))
	d2 := template.Must(base.Parse(`{{define "body"}}d2{{end}}`))

	d1.Execute(os.Stdout, 2)
	base.Execute(os.Stdout, 3)
	d2.Execute(os.Stdout, 4)
	base.Execute(os.Stdout, 5)

	// Output:
	// 1base
	// 2based1
	// 3based2
	// 4based2
	// 5based2
}
