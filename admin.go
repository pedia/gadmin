package gadmin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/leonelquinteros/gotext.v1"
	"gorm.io/gorm"
)

func NewAdmin(name string, db *gorm.DB) *Admin {
	A := Admin{
		DB: db,

		menu:  []*MenuItem{},
		views: []View{},

		debug: true,
		// auto_migrate: true,
		staticUrl: "/admin/static",
		secret:    NewSecret("hello"), // TODO: read from config
		mux:       http.NewServeMux(),
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
	A.RegisterTo(&A, A.mux, "")

	// TODO: read lang from config
	gotext.Configure("translations", "zh_Hant_TW", "admin")
	A.menu.Add(&MenuItem{Path: "/admin/", Name: A.gettext("Home")})
	// A.csrf = NewCSRF(A.secret)
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

	secret *Secret
	csrf   *CSRF
	mux    *http.ServeMux
}

func (A *Admin) register(b *Blueprint) {
	A.Add(b)

	b.RegisterTo(A, A.mux, A.Path)
}

func (A *Admin) AddView(view View) View {
	if mv, ok := view.(*ModelView); ok {
		mv.admin = A
		if A.auto_migrate {
			if err := A.DB.AutoMigrate(mv.model.new()); err != nil {
				return nil
			}
		}

		b := mv.GetBlueprint()
		if b != nil {
			A.views = append(A.views, mv)
			A.register(b)

			A.addViewToMenu(mv)
		}
	} else {
		bv, ok := view.(*BaseView)
		if ok {
			bv.admin = A
		}
		A.views = append(A.views, view)
		// TODO:
		view.GetBlueprint().RegisterTo(A, A.mux, "")
	}

	return view
}

func (A *Admin) addViewToMenu(view View) {
	menu := view.GetMenu()
	if menu != nil {
		// patch MenuItem.Path
		if menu.Path == "" {
			menu.Path, _ = A.GetUrl(view.GetBlueprint().Endpoint + ".index")
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

// Flask.url_for, `endpoint` like:
// admin.index
// model.create_view
// .create_view
func (A *Admin) UrlFor(model, endpoint string, args ...any) (string, error) {
	prefix := ""
	b := A.Blueprint
	if model != "" {
		cb, ok := A.Blueprint.Children[model]
		if !ok {
			return "", fmt.Errorf("model '%s' miss", model)
		}
		prefix = A.Path
		b = cb
	}

	res, err := b.GetUrl(endpoint, pairToQuery(args...))
	if err != nil {
		return "", err
	}
	return prefix + res, nil
}

func (A *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r2 := PatchSession(r, A)
	w2 := NewBufferWriter(w, func(w http.ResponseWriter) {
		sess := CurrentSession(r2)
		// log.Printf("session %s save %d", r.URL, len(sess.Values))
		err := sess.Save(w)
		if err != nil {
			// log.Printf("session save failed %s", err)
		}
	})
	A.mux.ServeHTTP(w2, r2)
	w2.(http.Flusher).Flush()
}

func (A *Admin) Run() {
	serv := http.Server{Handler: A}

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

// convince for outside of `Admin`
func gettext(format string, a ...any) string {
	return gotext.Get(format, a...)
}

var themeIndex = 5

func theme() string {
	var themes = []string{
		"cyborg", "darkly", "slate", // night
		"solar", "superhero", // dark
		"cerulean", "cosmo", "default", "flatly", "journal", "litera",
		"lumen", "lux", "materia", "minty", "united", "pulse",
		"sandstone", "simplex", "sketchy", "spacelab", "yeti",
	}
	// themeIndex = 0
	return themes[themeIndex%len(themes)]
}

func (A *Admin) dict(others ...map[string]any) map[string]any {
	o := map[string]any{
		"debug": A.debug,
		"name":  A.Name,
		"url":   "/admin",
		// "admin_base_template": "base.html",
		"swatch": theme(), // "cerulean", "default"
		"menus":  A.menu.dict(),
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
	s := CurrentSession(r)
	c := s.Get("C")
	if c == nil {
		c = 0
	}
	s.Set("C", c.(int)+1)
	ReplyJson(w, 200, A.dict(map[string]any{
		"blueprints": A.Blueprint.dict(),
		"session":    s.Values,
	}))
}
func (A *Admin) test_handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", ContentTypeUtf8Html)
	tx, err := template.New("test").
		Option("missingkey=error").
		Funcs(A.funcs(nil)).
		ParseFiles("templates/test.gotmpl")
	if err == nil {
		type foo struct {
			lower string
			Upper string
		}

		err = tx.Lookup("test.gotmpl").Execute(w, map[string]any{
			"lower":            "lower",
			"Upper":            "Upper",
			"int":              34,
			"emptyString":      "",
			"emptyInt":         0,
			"emptyIntArray":    []int{},
			"emptyStringArray": []string{},
			//
			"rfoo": foo{Upper: "Upper", lower: "lower"},
			"ffoo": func() foo { return foo{Upper: "Upper", lower: "lower"} },

			//
			"msa": map[string]any{"Upper": "Upper", "lower": "lower"},

			"null":  nil,
			"list":  []string{"a", "b"},
			"ss":    []struct{ A string }{{A: "a"}, {A: "b"}},
			"ls":    []map[string]any{{"A": "a"}, {"B": "b"}},
			"boolf": func() bool { return false },
			"boolt": func() bool { return true },
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

func (A *Admin) funcs(funcs template.FuncMap) template.FuncMap {
	fm := merge(sprig.FuncMap(), Funcs)
	merge(fm, template.FuncMap{
		"admin_static_url": A.staticURL, // used
		"marshal":          A.marshal,   // test
		"config":           A.config,    // used
		"gettext":          A.gettext,   //
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

	if funcs != nil {
		merge(fm, funcs)
	}
	return fm
}
