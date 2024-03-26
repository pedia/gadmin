package gadmin

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"reflect"
)

type Admin struct {
	*http.Server
	Views  []View
	Router *http.ServeMux
}

func NewAdmin(name string) *Admin {
	mux := http.NewServeMux()
	a := Admin{
		Server: &http.Server{
			Handler: mux,
		},
		Views:  []View{},
		Router: mux,
	}

	//
	a.Router.HandleFunc("/admin/", a.Index)
	return &a
}

func (a *Admin) AddView(v View) {
	// a.Views = append(a.Views, v)
	a.Router.Handle(fmt.Sprintf("/%s/", v.Name()), v)
}

func (a *Admin) Index(w http.ResponseWriter, r *http.Request) {
	t, err := template.New("all").Funcs(template.FuncMap{
		"admin_static_url": func(arg0 reflect.Value, arg1 reflect.Value) reflect.Value {
			return reflect.ValueOf("/admin/static/" + arg0.String() + "?" + arg1.String())
		},
	}).ParseFiles(
		// "templates/base.html",
		// "templates/index.html",
		// "templates/layout.html",
		// "templates/static.html",
		// "templates/lib.html",
		// "templates/actions.html",
		"templates/test.html",
	)
	if err != nil {
		fmt.Print(err)
		return
	}

	if err := t.Lookup("test.html").Execute(w, map[string]any{
		"foo":  "bar",
		"Zoo":  "Bar",
		"List": []string{"a", "b"},
		// {{ .admin_static.Url x y}}
		"admin_static": a,

		// {{call .admin_static_url x y}}
		"admin_static_url": a.Url,
	}); err != nil {
		fmt.Print(err)
	}
}

// template function
func (*Admin) Url(filename, ver string) string {
	return filename + "?" + ver
}

func (a *Admin) Run() {
	a.Handler = a.Router
	l, _ := net.Listen("tcp", "127.0.0.1:3333")
	a.Serve(l)
}

type View interface {
	Name() string
	Category() string
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type BaseView struct {
	*http.ServeMux
	name     string
	category string
}

func (bv *BaseView) Name() string     { return bv.name }
func (bv *BaseView) Category() string { return bv.category }

// TODO: NewModelView NewBaseView
func NewView(name, category string) View {
	return &BaseView{
		ServeMux: http.NewServeMux(),
		name:     name,
		category: category,
	}
}
