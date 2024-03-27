package gadmin

import (
	"fmt"
	"html/template"
	"maps"
	"net"
	"net/http"
)

type Admin struct {
	name string
	*http.Server
	Views  []View
	Router *http.ServeMux
}

func NewAdmin(name string) *Admin {
	mux := http.NewServeMux()
	a := Admin{
		name: name,
		Server: &http.Server{
			Handler: mux,
		},
		Views:  []View{},
		Router: mux,
	}

	//
	a.Router.HandleFunc("/admin/", a.Index)
	a.Router.HandleFunc("/admin/test", a.test)

	// Admin.Url
	a.Router.Handle("/static/", http.FileServer(http.Dir(".")))
	return &a
}

func (a *Admin) AddView(v View) {
	a.Views = append(a.Views, v)
	if bv, ok := v.(*BaseView); ok {
		bv.Admin = a
	}

	// {admin} / {view name} /
	// {admin} / {view name} / create

	for k, f := range v.Routers() {
		a.Router.Handle(fmt.Sprintf("/admin/%s%s", v.Name(), k), f)
	}
}

func (a *Admin) ts() *template.Template {
	t, err := template.New("all").Funcs(template.FuncMap{
		"admin_static_url": a.Url, // used
	}).ParseFiles(
		"templates/test.html",
		"templates/test_base.html",

		// "templates/layout.html",
		"templates/master.html",
		"templates/base.html",
		"templates/index.html",
		// "templates/layout.html",
		// "templates/static.html",
		// "templates/lib.html",
		// "templates/actions.html",
	)
	if err != nil {
		panic(err)
	}

	fmt.Print(t.DefinedTemplates())

	return t
}

func (a *Admin) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")

	if err := a.ts().Lookup("index.html").Execute(w, map[string]any{
		"category":            "ac",
		"name":                "admin",
		"admin_base_template": "base.html",
		"swatch":              "default",
		// {{ .admin_static_url x y}}
		"admin_static_url": a.Url,
	}); err != nil {
		panic(err)
	}
}

func (a *Admin) test(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")

	if err := a.ts().ExecuteTemplate(w, "test.html", map[string]any{
		"foo":   "bar",
		"empty": "",
		"null":  nil,
		"Zoo":   "Bar",
		"List":  []string{"a", "b"},
		"Conda": true,
		"Condb": false,
		// {{ .admin_static.Url x y}}
		"admin_static": a,

		// {{ .admin_static_url x y}}
		"admin_static_url": a.Url,
	}); err != nil {
		fmt.Print(err)
	}
}

// template function
func (*Admin) Url(filename, ver string) string {
	// TODO: hash
	s := "/static/" + filename
	if ver == "" {
		return s
	}
	return s + "?ver=" + ver
}

func (a *Admin) dict() map[string]any {
	return map[string]any{
		"name":                a.name,
		"url":                 "/admin/",
		"admin_base_template": "base.html",
		"swatch":              "default",
		// {{ .admin_static_url x y }}
		"admin_static_url": a.Url,
	}
}

func (a *Admin) Run() {
	a.Handler = a.Router
	l, _ := net.Listen("tcp", "127.0.0.1:3333")
	a.Serve(l)
}

type View interface {
	Name() string
	Category() string
	Routers() map[string]http.HandlerFunc
}

type BaseView struct {
	*Admin
	name     string
	category string
}

func (bv *BaseView) Name() string     { return bv.name }
func (bv *BaseView) Category() string { return bv.category }

func (bv *BaseView) Routers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"/": bv.Index,
	}
}

func (bv *BaseView) dict() map[string]any {
	ad := bv.Admin.dict()
	d := maps.Clone(ad)

	update(d, map[string]any{
		"category":  bv.category,
		"name":      bv.name,
		"admin":     ad,
		"extra_css": []string{},
		"extra_js":  []string{}, // "a.js", "b.js"}
	})
	return d
}

func update(a, b map[string]any) {
	for k, v := range b {
		a[k] = v
	}
}

func (bv *BaseView) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")

	if err := bv.ts().Lookup("index.html").Execute(w, bv.dict()); err != nil {
		panic(err)
	}
}

// TODO: NewModelView NewBaseView
func NewView(name, category string) View {
	return &BaseView{
		name:     name,
		category: category,
	}
}
