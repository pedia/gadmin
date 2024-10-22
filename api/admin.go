package api

import (
	"encoding/json"
	"fmt"
	"gadmin"
	"html/template"
	"net"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

func NewAdmin(name string, db *gorm.DB) *Admin {
	A := Admin{
		Blueprint: &Blueprint{Name: name, Endpoint: "admin", Path: "/admin"},
		DB:        db,

		menus: []MenuItem{},
		views: []View{},

		debug:     true,
		staticUrl: "/admin/static",

		mux: http.NewServeMux(),
	}

	// TODO:
	A.mux.HandleFunc("/admin/debug", A.debug_handle)

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

func (A *Admin) AddView(view View) {
	if mv, ok := view.(*ModelView); ok {
		if err := A.DB.AutoMigrate(mv.model.new()); err != nil {
			panic(err)
		}
	}

	A.views = append(A.views, view)
	A.Register(view.GetBlueprint())

	A.addViewToMenu(view)
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

func (A *Admin) dict() map[string]any {
	return map[string]any{
		"debug":               A.debug,
		"name":                A.Name,
		"url":                 A.Path,
		"admin_base_template": "base.html",
		"swatch":              "cerulean", // "default",
		// {{ .admin_static_url x y }}
		// "admin_static_url": a.staticUrl,
		"menus": lo.Map(A.menus, func(m MenuItem, _ int) map[string]any {
			return m.dict()
		}),
	}
}

// AddCategory/AddLink/AddMenuItem

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

func (A *Admin) render(fs ...string) *template.Template {
	fm := merge(sprig.FuncMap(), gadmin.Funcs)
	merge(fm, template.FuncMap{
		"admin_static_url": A.staticUrl, // used
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

func (A *Admin) debug_handle(w http.ResponseWriter, r *http.Request) {
	replyJson(w, 200, A.dict())
}
