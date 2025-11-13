package gadm

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	_ "github.com/gorilla/csrf"
	_ "github.com/gorilla/handlers"
	_ "github.com/gorilla/securecookie"
	_ "github.com/gorilla/sessions"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/svg"
)

// Middleware wraps given http.Handler adding some extra functionality
type Middleware func(next http.Handler) http.Handler

// Too weak, use gorilla/handlers.LoggingHandler
// Logger middleware logs every incoming request
func Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Printf("%s %s -- %v", r.Method, r.URL.Path, time.Since(start))
		})
	}
}

// Recovery middleware handles panics inside handlers
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					debug.PrintStack()
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Hook middleware runs a function on each request, can be used to reload templates, build frontend etc
func Hook(f func(*http.Request) bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !f(r) {
				next.ServeHTTP(w, r)
			}
		})
	}
}

// Use applies multiple middleware to the given handler
func Use(h http.Handler, mw ...Middleware) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// minified.Middleware
var minified *minify.M

func init() {
	minified = minify.New()
	minified.AddFunc("text/css", css.Minify)
	minified.AddFunc("text/html", html.Minify)
	minified.AddFunc("image/svg+xml", svg.Minify)
	minified.AddFunc("text/javascript", js.Minify)
	minified.AddFunc("application/javascript", js.Minify)
}
