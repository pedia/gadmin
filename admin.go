package gadmin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func NewAdmin(name string, db *gorm.DB) *Admin {
	A := Admin{
		DB: db,

		menus: []MenuItem{},
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
				Path:     "/debug.html",
				Handler:  A.debug_handle,
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

	return &A
}

type Admin struct {
	*Blueprint

	DB    *gorm.DB
	menus []MenuItem
	views []View

	debug bool

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
	// if bv, ok := view.(*BaseView); ok {}

	// TODO: flag for migrate
	if mv, ok := view.(*ModelView); ok {
		mv.admin = A

		if err := A.DB.AutoMigrate(mv.model.new()); err != nil {
			return err
		}
	}

	b := view.GetBlueprint()
	if b != nil {
		A.views = append(A.views, view)
		A.addViewToMenu(view)

		A.register(b)
	}
	return nil
}

func (A *Admin) addViewToMenu(view View) {
	menu := view.GetMenu()
	if menu != nil {
		A.menus = append(A.menus, *menu)
	}
}

// Return the menu hierarchy.
func (A *Admin) Menu() []MenuItem {
	return A.menus
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
		"menus": lo.Map(A.menus, func(m MenuItem, _ int) map[string]any {
			return m.dict()
		}),
	}

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func (A *Admin) render(fs ...string) *template.Template {
	fm := merge(sprig.FuncMap(), Funcs)
	merge(fm, template.FuncMap{
		"admin_static_url": A.staticURL, // used
		"get_url": func(endpoint string, args ...map[string]any) (string, error) {
			return A.UrlFor(endpoint, ""), nil
		},
		"marshal":    A.marshal, // test
		"config":     A.config,  // used
		"gettext":    A.gettext, //
		"csrf_token": func() string { return "xxxx-csrf-token" },
		// escape safe
		"safehtml": func(s string) template.HTML { return template.HTML(s) },
		"comment": func(format string, args ...any) template.HTML {
			return template.HTML(
				"<!-- " + fmt.Sprintf(format, args...) + " -->",
			)
		},
		"safejs": func(s string) template.JS { return template.JS(s) },
		"json": func(v any) (template.JS, error) {
			bs, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return template.JS(string(bs)), nil
		},
	})

	tx := template.Must(template.New("all").
		Option("missingkey=error").
		Funcs(fm).
		ParseFiles(fs...))
	// log.Println(tx.DefinedTemplates())
	return tx
}
func (A *Admin) index_handle(w http.ResponseWriter, r *http.Request) {
	// A.Render(w, "index.gotmpl", A.dict())
}

func (A *Admin) debug_handle(w http.ResponseWriter, r *http.Request) {
	replyJson(w, 200, A.dict(map[string]any{
		"blueprints": A.Blueprint.dict(),
	}))
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
func (A *Admin) static_handle(w http.ResponseWriter, r *http.Request) {

}
