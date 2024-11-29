package gadmin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func firstOr[T any](as []T, or ...T) T {
	if len(as) > 0 {
		return as[0]
	}
	if len(or) > 0 {
		return or[0]
	}
	var t T
	return t
}

// Ensure value avoid error/bool trouble
func must[T any](frs ...any) T {
	// try: func() (x, error)
	err, ok := frs[len(frs)-1].(error)
	if ok && err != nil {
		panic(err)
	}

	if !ok {
		// try: func() (x, bool)
		if b, ok := frs[len(frs)-1].(bool); ok && !b {
			panic("not ok")
		}
	}

	return frs[0].(T)
}

func anyMapToQuery(m map[string]any) url.Values {
	uv := url.Values{}
	for key, val := range m {
		uv.Add(key, cast.ToString(val))
	}
	return uv
}

// any to query in tradition url
// bool => "1", "0". behavior in python/flask
// others => string
func intoStringSlice(as ...any) []string {
	return lo.Map(as, func(a any, _ int) string {
		if b, ok := a.(bool); ok {
			return lo.Ternary(b, "1", "0")
		}
		return cast.ToString(a)
	})
}

// Input paired args, like: a,b,c,d return "a=b&c=d"
func pairToQuery(args ...any) url.Values {
	uv := url.Values{}
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key, ok := args[i].(string)
			if !ok {
				panic(fmt.Errorf("paired-args key not string %v", args[i]))
			}

			// `Add` is better than `Set`
			uv.Add(key, cast.ToString(args[i+1]))
		}
	}
	return uv
}

// Merge b to a
func merge[K comparable, V any](a, b map[K]V) map[K]V {
	for k, v := range b {
		a[k] = v
	}
	return a
}

var (
	ContentTypeJson     = "application/json; charset=utf-8"
	ContentTypeUtf8Html = "text/html; charset=utf-8"
)

func ReplyJson(w http.ResponseWriter, status int, o any) {
	w.Header().Add("content-type", ContentTypeJson)
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(o); err != nil {
		panic(err)
	}
}

// Cache http.ResponseWriter, Output Cookie after template rendering
func NewBufferWriter(w http.ResponseWriter, f func(http.ResponseWriter)) http.ResponseWriter {
	return &bufferWriter{
		buf:         bytes.NewBuffer([]byte{}),
		w:           w,
		beforeFlush: f}
}

type bufferWriter struct {
	buf *bytes.Buffer

	// origin `ResponseWriter`
	w http.ResponseWriter

	// excute before flush, eg. Set-Cookie
	beforeFlush func(http.ResponseWriter)
}

func (B *bufferWriter) Write(p []byte) (n int, err error) {
	return B.buf.Write(p)
}

func (B *bufferWriter) Header() http.Header {
	return B.w.Header()
}

func (B *bufferWriter) WriteHeader(statusCode int) {
	B.w.WriteHeader(statusCode)
}

func (B *bufferWriter) Flush() {
	if B.beforeFlush != nil {
		B.beforeFlush(B.w)
	}
	B.w.Write(B.buf.Bytes())
	B.w.(http.Flusher).Flush()
}
