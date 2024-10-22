package api

import (
	"fmt"
	"net/http"
	"strings"
)

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
	Children map[string]*Blueprint // endpoint => Blueprint
	Name     string                // Foo
	Handler  http.HandlerFunc
	// StaticFolder
	// StaticUrlPath
	// TemplateFolder
	// ErrorHandler
}

func (B *Blueprint) Register(child *Blueprint) {
	if B.Children == nil {
		B.Children = map[string]*Blueprint{}
	}
	B.Children[child.Endpoint] = child
}

func (B *Blueprint) GetUrl(endpoint string, args ...string) string {
	path := ""
	arr := strings.SplitN(endpoint, ".", 2)
	if arr[0] == "" || arr[0] == B.Endpoint {
		path = B.Path
	} else {
		child, ok := B.Children[arr[0]]
		if ok {
			return B.Path + "/" + child.GetUrl(endpoint, args...)
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
