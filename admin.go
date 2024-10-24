package gadmin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"gorm.io/gorm"
)

func NewAdmin(name string, db *gorm.DB) *Admin {
	A := Admin{
		DB: db,

		menu:  []*MenuItem{},
		views: []View{},

		debug:     true,
		staticUrl: "/admin/static",

		mux: http.NewServeMux(),
	}

	A.Blueprint = &Blueprint{
		Name:     name,
		Endpoint: "admin",
		Path:     "/admin",
		Handler:  A.index_handle,
		Children: map[string]*Blueprint{
			// TODO: A.debug
			"debug": {
				Endpoint: "debug",
				Path:     "/debug.json",
				Handler:  A.debug_handle,
			},
			"test": {
				Endpoint: "test",
				Path:     "/test.html",
				Handler:  A.test_handle,
			},
			"static": {
				Endpoint: "static",
				Path:     "/static/",
				Register: func(mux *http.ServeMux, path string, bp *Blueprint) {
					fs := http.FileServer(http.Dir("static")) // TODO: Blueprint.StaticFolder
					mux.Handle(path+bp.Path, http.StripPrefix(path+bp.Path, fs))
				},
			},
		}}
	A.RegisterTo(A.mux, "")
	// TODO: gettext("Home")
	A.menu.Add(&MenuItem{Path: "/admin/", Name: "Home"})

	return &A
}

type Admin struct {
	*Blueprint

	DB    *gorm.DB
	menu  Menu
	views []View

	debug        bool
	auto_migrate bool

	// Url prefix for all static resource, default as `/static`
	// with Admin.name, default as `/admin/static`
	staticUrl string

	mux *http.ServeMux
}

func (A *Admin) register(b *Blueprint) {
	A.Add(b)

	b.RegisterTo(A.mux, A.Path)
}

func (A *Admin) AddView(view View) error {
	// not work:
	// if bv, ok := view.(*BaseView); ok {}

	if mv, ok := view.(*ModelView); ok {
		mv.admin = A

		if A.auto_migrate {
			if err := A.DB.AutoMigrate(mv.model.new()); err != nil {
				return err
			}
		}
	}

	b := view.GetBlueprint()
	if b != nil {
		A.views = append(A.views, view)
		A.register(b)

		A.addViewToMenu(view)
	}
	return nil
}

func (A *Admin) addViewToMenu(view View) {
	menu := view.GetMenu()
	if menu != nil {
		// patch MenuItem.Path
		if menu.Path == "" {
			menu.Path = A.GetUrl(view.GetBlueprint().Endpoint + ".index")
		}
		A.menu.Add(menu)
	}
}

func (A *Admin) AddLink(cate, name, path string) {
	A.menu.Add(&MenuItem{Category: cate, Name: name, Path: path})
}
func (A *Admin) AddCategory(cate string) {
	A.menu.Add(&MenuItem{Category: cate, Name: cate})
}
func (A *Admin) AddMenuItem(mi MenuItem) {
	A.menu.Add(&mi)
}

// AddCategory/AddLink/AddMenuItem

func (A *Admin) staticURL(filename, ver string) string {
	s := A.staticUrl + "/" + filename
	if ver != "" {
		s += "?ver=" + ver
	}
	return s
}

// `endpoint` like:
// admin.index
// model.create_view
func (A *Admin) UrlFor(endpoint string, args ...string) string {
	return A.GetUrl(endpoint, args...)
}

func (A *Admin) Run() {
	serv := http.Server{Handler: A.mux}

	l, _ := net.Listen("tcp", ":3333")
	serv.Serve(l)
}

// template function
func (*Admin) marshal(v any) string {
	bs, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err.Error()
	}
	return string(bs)
}
func (*Admin) config(key string) bool {
	return false
}
func (*Admin) gettext(format string, a ...any) string {
	return gettext(format, a...)
}

func gettext(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

func (A *Admin) dict(others ...map[string]any) map[string]any {
	o := map[string]any{
		"debug":               A.debug,
		"name":                A.Name,
		"url":                 "/admin",
		"admin_base_template": "base.html",
		"swatch":              "cerulean", // "default",
		// {{ .admin_static_url x y }}
		// "admin_static_url": a.staticUrl,
		"menus": A.menu.dict(),
	}

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func (A *Admin) index_handle(w http.ResponseWriter, r *http.Request) {
	// A.Render(w, "index.gotmpl", A.dict())
}

func (A *Admin) debug_handle(w http.ResponseWriter, r *http.Request) {
	replyJson(w, 200, A.dict(map[string]any{
		"blueprints": A.Blueprint.dict(),
	}))
}
func (A *Admin) test_handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", contentTypeUtf8Html)
	tx, err := template.New("test").
		Option("missingkey=error").
		Funcs(templateFuncs(A)).
		ParseFiles("templates/test.gotmpl")
	if err == nil {
		type foo struct {
			lower string
			Upper string
		}
		type msa map[string]any

		err = tx.Lookup("test.gotmpl").Execute(w, map[string]any{
			"foo":              "bar",
			"emptyString":      "",
			"emptyInt":         0,
			"emptyIntArray":    []int{},
			"emptyStringArray": []string{},
			"null":             nil,
			"Zoo":              "Bar",
			"list":             []string{"a", "b"},
			"ss":               []struct{ A string }{{A: "a"}, {A: "b"}},
			"ls":               []map[string]any{{"A": "a"}, {"B": "b"}},
			"conda":            true,
			"condb":            false,
			"int":              34,
			"boolf":            func() bool { return false },
			"boolt":            func() bool { return true },
			"rs":               func() foo { return foo{lower: "lower", Upper: "Upper"} },
			"map":              func() map[string]any { return map[string]any{"lower": "Lower", "Upper": "Upper"} },
			"msa":              func() msa { return msa{"lower": "Lower", "Upper": "Upper"} },
			"msas":             func() []msa { return []msa{{"lower": "Lower", "Upper": "Upper"}} },
			"msas2":            func() ([]msa, error) { return []msa{{"lower": "Lower", "Upper": "Upper"}}, nil },

			// map is better than struct
			"admin":   A.dict(),
			"request": rd(r),
			// bad
			// {{ .admin_static.Url x y}}
			"admin_static": A,
			// {{ .admin_static_url x y}}
			"admin_static_url": A.staticUrl,
		})
	}

	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

// serve admin.static
// url /admin/static/{} => local static/{}
// first way:
// fs := http.FileServer(http.Dir("static"))
// a.Mux.Handle("/admin/static/", http.StripPrefix("/admin/static/", fs))
//
// second way:
// a.Mux.HandleFunc("/admin/static/{path...}",
//
//	func(w http.ResponseWriter, r *http.Request) {
//	        path := r.PathValue("path")
//	        http.ServeFileFS(w, r, os.DirFS("static"), path)
//	})
// func (A *Admin) static_handle(w http.ResponseWriter, r *http.Request) {}
