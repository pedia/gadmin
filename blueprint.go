package gadmin

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
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
//
// A blueprint is A model and dependent pages
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
func (B *Blueprint) RegisterTo(admin *Admin, mux *http.ServeMux, path string) {
	log.Printf("handle %s %v", path+B.Path, B.Handler != nil || B.Register != nil)
	if B.Register != nil {
		B.Register(mux, path, B)
	}

	if B.Handler != nil {
		mux.HandleFunc(path+B.Path, func(w http.ResponseWriter, r *http.Request) {
			// Inject with `Session` for all pages
			B.Handler(w, PatchSession(r, admin))
		})
	}

	unique := map[string]bool{}
	for _, cb := range B.Children {
		cp := path + B.Path + cb.Path
		_, ok := unique[cp]
		if !ok {
			cb.RegisterTo(admin, mux, path+B.Path)
			unique[cp] = true
		} else {
			log.Printf("duplicated handle %s", cp)
		}
	}
}

func (B *Blueprint) GetUrl(endpoint string, qs ...url.Values) (string, error) {
	path := ""
	arr := strings.SplitN(endpoint, ".", 2)
	if arr[0] == "" || arr[0] == B.Endpoint {
		path = B.Path
	} else {
		child, ok := B.Children[arr[0]]
		if ok {
			res, err := child.GetUrl(endpoint, qs...)
			return B.Path + res, err
		} else {
			return "", fmt.Errorf("endpoint '%s' miss in `%s`", arr[1], B.Endpoint)
		}
	}

	if len(arr) == 2 {
		child, ok := B.Children[arr[1]]
		if ok {
			res, err := child.GetUrl(strings.Join(arr[1:], "."), qs...)
			return path + res, err
		} else {
			return "", fmt.Errorf("endpoint '%s' miss in `%s`", arr[1], B.Endpoint)
		}
	}
	if len(qs) > 0 && len(qs[0]) > 0 {
		return B.Path + "?" + qs[0].Encode(), nil
	}
	return B.Path, nil
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
