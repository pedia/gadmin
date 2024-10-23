package gadmin

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type RegisterFunc func(*http.ServeMux, string, *Blueprint)

// like flask.Blueprint
//
// | Name  | Endpoint       | Path       |
// |-------|----------------|------------|
// | Foo   | foo            | foo        |
// |       | .index         | /          |
// |       | .action_view   | /action    |
// |       | foo.index      | foo/       |
// | Admin | admin          | /admin     |
// |       | .index         | /          |
type Blueprint struct {
	Endpoint string                // {foo}.index
	Path     string                // /foo
	Children map[string]*Blueprint // endpoint => *Blueprint
	Name     string                // Foo
	Handler  http.HandlerFunc
	Register RegisterFunc // Custom register to mux, serve static file
	// StaticFolder
	// StaticUrlPath
	// TemplateFolder
	// ErrorHandler
}

// like flask `Blueprint.Register`
func (B *Blueprint) Add(child *Blueprint) {
	if B.Children == nil {
		B.Children = map[string]*Blueprint{}
	}
	B.Children[child.Endpoint] = child
}

// Add `Blueprint` to `http.ServeMux`
func (B *Blueprint) RegisterTo(mux *http.ServeMux, path string) {
	log.Printf("handle %s %v", path+B.Path, B.Handler != nil || B.Register != nil)
	if B.Register != nil {
		B.Register(mux, path, B)
	}

	if B.Handler != nil {
		mux.HandleFunc(path+B.Path, B.Handler)
	}

	unique := map[string]bool{}
	for _, cb := range B.Children {
		cp := path + B.Path + cb.Path
		_, ok := unique[cp]
		if !ok {
			cb.RegisterTo(mux, path+B.Path)
			unique[cp] = true
		} else {
			log.Printf("duplicated handle %s", cp)
		}
	}
}

func (B *Blueprint) GetUrl(endpoint string, args ...string) string {
	path := ""
	arr := strings.SplitN(endpoint, ".", 2)
	if arr[0] == "" || arr[0] == B.Endpoint {
		path = B.Path
	} else {
		child, ok := B.Children[arr[0]]
		if ok {
			return B.Path + child.GetUrl(endpoint, args...)
		} else {
			panic(fmt.Errorf("endpoint '%s' miss in `%s`", arr[1], B.Endpoint))
		}
	}

	if len(arr) == 2 {
		child, ok := B.Children[arr[1]]
		if ok {
			return path + child.GetUrl(strings.Join(arr[1:], "."), args...)
		} else {
			panic(fmt.Errorf("endpoint '%s' miss in `%s`", arr[1], B.Endpoint))
		}
	}
	return B.withQuery(B.Path, args...)
}

// Safe append query to path
func (B *Blueprint) withQuery(path string, args ...string) string {
	if qs := pairToQueryString(args...); qs != "" {
		return path + "?" + qs
	}
	return path
}

func (B *Blueprint) dict() map[string]any {
	o := map[string]any{
		"endpoint": B.Endpoint,
		"path":     B.Path,
		"handler":  B.Handler == nil,
	}

	if B.Name != "" {
		o["name"] = B.Name
	}

	if B.Children != nil {
		oc := map[string]any{}
		for k, v := range B.Children {
			oc[k] = v.dict()
		}
		o["children"] = oc
	}
	return o
}

// menu

// admin scope:
// admin.logout_view
// admin.index
// admin.static

// security(login) scope:
// security.login
// security.logout
// security.register
// security.forgot_password
// security.send_confirmation
