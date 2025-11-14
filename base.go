package gadm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"net/http"
	"net/url"
	"reflect"

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

// a != "" ? a or b
func emptyOr[T any](a, b T) T {
	if !reflect.ValueOf(a).IsZero() {
		return a
	}
	return b
}

// Ensure value avoid error/bool trouble
func must[T any](v T, lefts ...any) T {
	if len(lefts) == 0 {
		return v
	}

	// try: func() (x, error)
	if lefts[0] == nil {
		return v
	}

	if err, ok := lefts[0].(error); ok {
		panic(err)
	}

	// try: func() (x, bool)
	if err, ok := lefts[0].(bool); ok && err {
		return v
	}
	panic("type wrong")
}

// any to query in tradition url
// bool => "1", "0". behavior in python/flask
// others => string
func pyslice(as ...any) []string {
	return lo.Map(as, func(a any, _ int) string {
		if b, ok := a.(bool); ok {
			return lo.Ternary(b, "1", "0")
		}
		return cast.ToString(a)
	})
}

// Input paired args, like: a,b,c,d return "a=b&c=d"
func pairsToQuery(args ...any) url.Values {
	uv := url.Values{}
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			key := cast.ToString(args[i])

			// `Add` is better than `Set`
			uv.Add(key, cast.ToString(args[i+1]))
		}
	}
	return uv
}

func queryToPairs(uv url.Values) []any {
	arr := []any{}
	for k, vs := range uv {
		for _, v := range vs {
			arr = append(arr, k, v)
		}
	}
	return arr
}

func mapContains[K comparable, V any](m map[K]V, k K) bool {
	_, ok := m[k]
	return ok
}

// Merge b to a
func merge[K comparable, V any](a, b map[K]V) map[K]V {
	maps.Copy(a, b)
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

// Some go template actions might change headers(Cookie)
// [CacheWriter] cache body and output headers before any body
// Flush in admin.ServeHTTP
type CachedWriter struct {
	http.ResponseWriter
	cache      bytes.Buffer
	header     http.Header
	statusCode int
}

func NewCachedWriter(w http.ResponseWriter) *CachedWriter {
	return &CachedWriter{ResponseWriter: w,
		header:     http.Header{},
		statusCode: http.StatusOK,
	}
}

func (cw *CachedWriter) Header() http.Header         { return cw.header }
func (cw *CachedWriter) Write(b []byte) (int, error) { return cw.cache.Write(b) }
func (cw *CachedWriter) WriteHeader(statusCode int)  { cw.statusCode = statusCode }

// Hijack implements the http.Hijacker interface by attempting to unwrap the writer.
func (cw *CachedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	// Attempt to unwrap the underlying ResponseWriter if possible
	unwrapped := cw.ResponseWriter
	if u, ok := unwrapped.(interface{ Unwrap() http.ResponseWriter }); ok {
		unwrapped = u.Unwrap()
	}

	// Try to perform the hijack on the potentially unwrapped writer
	if hijacker, ok := unwrapped.(http.Hijacker); ok {
		fmt.Println("Hijacking connection...")
		return hijacker.Hijack()
	}

	// If the underlying writer doesn't support hijacking, use ResponseController (Go 1.20+)
	// to attempt it in a standard way, or return an error.
	rc := http.NewResponseController(cw.ResponseWriter)
	if conn, bufrw, err := rc.Hijack(); err == nil {
		fmt.Println("Hijacking connection via ResponseController...")
		return conn, bufrw, nil
	}

	// Return error if not supported
	return nil, nil, http.ErrNotSupported
}

func (cw *CachedWriter) Flush() {
	for k, vs := range cw.header {
		for _, v := range vs {
			cw.ResponseWriter.Header().Add(k, v)
		}
	}
	cw.ResponseWriter.WriteHeader(cw.statusCode)
	cw.ResponseWriter.Write(cw.cache.Bytes())
}
