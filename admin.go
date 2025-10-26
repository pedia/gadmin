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
	A := &Admin{
		DB: db,
		BaseView: &BaseView{Menu: Menu{
			Path: "/admin/",
			Name: gettext("Home"),
		}},
		views:       []View{},
		debug:       true,
		autoMigrate: true,
		secret:      NewSecret("hello"), // TODO: read from config
		mux:         http.NewServeMux(),

		indexTemplateFile: "index.gotmpl",
	}
	A.BaseView.admin = A

	A.Blueprint = &Blueprint{
		Name:     name,
		Endpoint: "admin",
		Path:     "/admin",
		Handler:  A.indexHandler,
		Children: map[string]*Blueprint{
			"index":      {Endpoint: "index", Path: "/", Handler: A.indexHandler},
			"debug":      {Endpoint: "debug", Path: "/debug.json", Handler: A.debugHandler},
			"debug.html": {Endpoint: "debug.html", Path: "/debug.html", Handler: A.debugHtmlHandler},
			"test":       {Endpoint: "test", Path: "/test", Handler: A.testHandler},
			"static":     {Endpoint: "static", Path: "/static/", StaticFolder: "static"},
		}}

	A.Blueprint.registerTo(A.mux, "")

	// TODO: read lang from config
	gotext.Configure("translations", "zh_Hant_TW", "admin")

	// AddSecurity(&A)
	return A
}

type Admin struct {
	*BaseView
	DB *gorm.DB

	views             []View
	debug             bool
	autoMigrate       bool
	secret            *Secret
	mux               *http.ServeMux
	indexTemplateFile string
}

func (A *Admin) Register(b *Blueprint) {
	A.Blueprint.AddChild(b)

	b.registerTo(A.mux, A.Blueprint.Path)
}

func (A *Admin) AddView(view View) View {
	view.setAdmin(A)

	if b := view.GetBlueprint(); b != nil {
		A.views = append(A.views, view)
		A.Register(b)

		A.addViewToMenu(view)
	}
	return view
}

func (A *Admin) addViewToMenu(view View) {
	if menu := view.GetMenu(); menu != nil {
		// patch MenuItem.Path
		if menu.Path == "" {
			menu.Path, _ = A.Blueprint.GetUrl(view.GetBlueprint().Endpoint + ".index")
		}
		A.BaseView.Menu.AddMenu(menu)
	}
}

func (A *Admin) staticURL(filename, ver string) string {
	path, err := A.Blueprint.GetUrl(".static")
	if err == nil {
		if ver != "" {
			return path + filename + "?ver=" + ver
		}
		return path + filename
	}
	panic(err)
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
		prefix = A.Blueprint.Path
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
	serv := http.Server{Handler: Use(A, Logger())}

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

func theme() string {
	var themes = []string{
		"cyborg", "darkly", "slate", // night
		"solar", "superhero", // dark
		"cerulean", "cosmo", "default", "flatly", "journal", "litera",
		"lumen", "lux", "materia", "minty", "united", "pulse",
		"sandstone", "simplex", "sketchy", "spacelab", "yeti",
	}
	themeIndex := 2
	return themes[themeIndex]
}

func (A *Admin) dict(others ...map[string]any) map[string]any {
	o := map[string]any{
		"debug": A.debug,
		"name":  A.Blueprint.Name,
		"url":   A.Blueprint.Path, // "/admin"
		// "admin_base_template": "base.html",
		// 'swatch' from flask-admin
		"swatch": theme(), // "cerulean", "default"
		"menu":   A.BaseView.Menu.dict(),
		"config": config,
	}

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func (A *Admin) indexHandler(w http.ResponseWriter, r *http.Request) {
	A.Render(w, r, A.indexTemplateFile, nil, A.dict())
}
func (A *Admin) testHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test"))
}
func (A *Admin) debugHandler(w http.ResponseWriter, r *http.Request) {
	s := CurrentSession(r)
	c := s.Get("C")
	if c == nil {
		c = 0
	}
	s.Set("C", c.(int)+1)
	ReplyJson(w, 200, A.dict(map[string]any{
		"blueprint": A.Blueprint.dict(),
		"session":   s.Values,
	}))
}
func (A *Admin) debugHtmlHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", ContentTypeUtf8Html)
	tx, err := template.New("debug").
		Option("missingkey=error").
		Funcs(A.funcs(nil)).
		ParseFiles("templates/debug.gotmpl")
	if err != nil {
		panic(err)
	}

	err = tx.Lookup("debug.gotmpl").Execute(w, A.dict())
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

func (A *Admin) funcs(more template.FuncMap) template.FuncMap {
	res := merge(sprig.FuncMap(), Funcs)
	merge(res, template.FuncMap{
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

	if more != nil {
		merge(res, more)
	}
	return res
}

func (A *Admin) SetIndexTemplateFile(nfn string) {
	A.indexTemplateFile = nfn
}
