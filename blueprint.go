package gadmin

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// like flask.Blueprint
//
// | Name  | Endpoint       | Path       |
// |-------|----------------|------------|
// | Foo   | foo            | /foo/      |
// |       | .index         | /          |
// |       | .action_view   | /action    |
// |       | foo.index      | /foo/      |
// | Admin | admin          | /admin/    |
// |       | .index         | /          |
//
// A blueprint is A model and dependent pages
type Blueprint struct {
	Endpoint string                // {foo}.index
	Path     string                // /foo
	Children map[string]*Blueprint // endpoint => *Blueprint
	Name     string                // Foo
	Handler  http.HandlerFunc

	// Custom register into http.ServerMux
	RegisterFunc func(*http.ServeMux, string, *Blueprint)

	StaticFolder   string
	TemplateFolder string

	// StaticUrlPath
	// ErrorHandler
}

// like flask `Blueprint.Register`
func (B *Blueprint) AddChild(child *Blueprint) {
	if B.Children == nil {
		B.Children = map[string]*Blueprint{}
	}
	B.Children[child.Endpoint] = child
}

// Register all Blueprint to `http.ServeMux`
func (B *Blueprint) registerTo(mux *http.ServeMux, parent string) {
	if B.RegisterFunc != nil {
		B.RegisterFunc(mux, parent, B)
	} else if B.Handler != nil {
		if !strings.HasPrefix(B.Path, "/") {
			log.Printf("warning: Blueprint(%s path: %s) not start with /", B.Name, B.Path)
		}

		log.Printf("%s handle %s", B.Name, parent+B.Path)
		if strings.HasSuffix(B.Path, "/") {
			mux.HandleFunc(parent+B.Path+"{$}", B.Handler)
		} else {
			mux.HandleFunc(parent+B.Path, B.Handler)
		}
	} else if B.Endpoint == "static" && B.StaticFolder != "" {
		log.Printf("%s handle %s fs: %s", B.Name, parent+B.Path, B.StaticFolder)

		if !strings.HasSuffix(B.Path, "/") {
			panic("Blueprint(Name='static').Path should end with /")
		}

		fs := http.FileServer(http.Dir(B.StaticFolder))
		mux.Handle(parent+B.Path, //minified.Middleware(
			http.StripPrefix(parent+B.Path, fs))

		// TODO: add an endpoint
	}

	// avoid duplicated Path
	up := map[string]bool{}
	for _, child := range B.Children {
		if unique := up[child.Path]; !unique {
			child.registerTo(mux, parent+B.Path)

			up[child.Path] = true
		}
	}
}

func (B *Blueprint) GetUrl(endpoint string, qs ...any) (string, error) {
	var path string
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
	if len(qs) > 0 {
		return B.Path + "?" + pairToQuery(qs...).Encode(), nil
	}
	return B.Path, nil
}

func (B *Blueprint) dict() map[string]any {
	o := map[string]any{
		"endpoint": B.Endpoint,
		"path":     B.Path,
		"handler":  B.Handler == nil,
		"name":     B.Name,
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
// security.login /login
// security.logout /logout
// security.register /register
// security.forgot_password
// security.send_confirmation
