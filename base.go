package gadmin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func firstOrEmpty[T any](as ...T) T {
	if len(as) > 0 {
		return as[0]
	}
	var t T
	return t
}

// Ensure value avoid error/bool trouble
func must[T any](xs ...any) T {
	// try: func() (x, error)
	err, ok := xs[len(xs)-1].(error)
	if ok && err != nil {
		panic(err)
	}

	if !ok {
		// try: func() (x, bool)
		if b, ok := xs[len(xs)-1].(bool); ok && !b {
			panic("not ok")
		}
	}

	return xs[0].(T)
}

func anyMapToQuery(m map[string]any) url.Values {
	uv := url.Values{}
	for key, val := range m {
		uv.Set(key, fmt.Sprint(val))
	}
	return uv
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
			uv.Add(key, fmt.Sprint(args[i+1]))
		}
	}
	return uv
}

// Merge b map to a
func merge[K comparable, V any](a, b map[K]V) map[K]V {
	for k, v := range b {
		a[k] = v
	}
	return a
}

var (
	contentTypeJson     = "application/json; charset=utf-8"
	contentTypeUtf8Html = "text/html; charset=utf-8"
)

func replyJson(w http.ResponseWriter, status int, obj any) {
	w.Header().Add("content-type", contentTypeJson)
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(obj); err != nil {
		panic(err)
	}
}
