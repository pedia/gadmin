package gadmin

import (
	"encoding/json"
	"fmt"
	"gadmin/isdebug"
	"html/template"
	"log"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"github.com/gorilla/websocket"
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
		debug:       isdebug.Enabled,
		autoMigrate: true,
		secret:      NewSecret("hello"), // TODO: read from config
		mux:         http.NewServeMux(),

		indexTemplateFile: "templates/index.gotmpl",
	}
	A.BaseView.admin = A

	if A.DB != nil {
		A.trace = NewTrace(A.DB)
	}

	A.Blueprint = &Blueprint{
		Name:     name,
		Endpoint: "admin",
		Path:     "/admin",
		Handler:  A.indexHandler,
		Children: map[string]*Blueprint{
			"index":      {Endpoint: "index", Path: "/", Handler: A.indexHandler},
			"debug":      {Endpoint: "debug", Path: "/debug.json", Handler: A.debugHandler},
			"debug.html": {Endpoint: "debug.html", Path: "/debug.html", Handler: A.debugHtmlHandler},
			"generate":   {Endpoint: "generate", Path: "/generate", Handler: A.generateHandler},
			"console":    {Endpoint: "console", Path: "/console", Handler: A.consoleHandler},
			"trace":      {Endpoint: "trace", Path: "/trace", Handler: A.traceHandler},
			"test":       {Endpoint: "test", Path: "/test", Handler: A.testHandler},
			"static":     {Endpoint: "static", Path: "/static/", StaticFolder: "static"},
		}}

	A.Blueprint.registerTo(A.mux, "")

	// TODO: read lang from config
	gotext.Configure("translations", "en", "admin")

	// AddSecurity(&A)
	return A
}

type Admin struct {
	*BaseView
	DB    *gorm.DB
	trace *Trace

	views             []View
	debug             bool
	autoMigrate       bool
	secret            *Secret
	mux               *http.ServeMux
	indexTemplateFile string
}

func (A *Admin) Register(b *Blueprint) {
	if err := A.Blueprint.AddChild(b); err != nil {
		fmt.Println(err)
		return
	}

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
		// CAUTION: patch MenuItem.Path
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

	res, err := b.GetUrl(endpoint, args...)
	if err != nil {
		return "", err
	}
	return prefix + res, nil
}

func (A *Admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: NewBufferWriter not implement hijack
	// r2 := PatchSession(r, A)
	// w2 := NewBufferWriter(w, func(w http.ResponseWriter) {
	// 	sess := CurrentSession(r2)
	// 	// log.Printf("session %s save %d", r.URL, len(sess.Values))
	// 	err := sess.Save(w)
	// 	if err != nil {
	// 		// log.Printf("session save failed %s", err)
	// 	}
	// })
	// A.mux.ServeHTTP(w2, r2)
	// w2.(http.Flusher).Flush()

	A.mux.ServeHTTP(w, r)

	defer A.trace.CheckTrace(r)
}

func (A *Admin) Run() {
	serv := http.Server{
		Addr: ":3333",
		// Handler: handlers.LoggingHandler(os.Stdout, A)}
		Handler: A}

	serv.ListenAndServe()
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
	themeIndex := 5
	return themes[themeIndex]
}

func (A *Admin) dict(others ...map[string]any) map[string]any {
	o := map[string]any{
		"debug": A.debug,
		"db":    A.DB != nil,
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
		"get_url":          A.Blueprint.GetUrl,
		// escape safe
		"safehtml": func(s string) template.HTML { return template.HTML(s) },
		"safejs":   func(s string) template.JS { return template.JS(s) },
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

type wsWriter struct {
	*websocket.Conn
}

func (w *wsWriter) Write(p []byte) (n int, err error) {
	w.WriteMessage(websocket.TextMessage, p)
	return n, nil
}

var g *generator

func (A *Admin) generateHandler(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		upgrader := websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("upgrade failed", err)
			return
		}

		// ignore all input
		go conn.ReadMessage()
		go func() {
			g.Run(A, &wsWriter{conn})
			g = nil // cleanup
		}()
		return
	}

	// normal GET/POST
	if r.Method == http.MethodGet {
		A.Render(w, r, "templates/generate.gotmpl", nil, map[string]any{"gen": nil})
	} else {
		r.ParseForm()
		g = NewGenerator(r.FormValue("url"))
		g.Package = r.FormValue("package")
		A.Render(w, r, "templates/generate.gotmpl", nil, map[string]any{"gen": g})
	}
}
func (A *Admin) consoleHandler(w http.ResponseWriter, r *http.Request) {
	result := &Result{Query: DefaultQuery(), Rows: []Row{}}
	var sql string
	if r.Method == http.MethodPost && A.DB != nil {
		r.ParseForm()
		sql = r.FormValue("sql")
		// CAUTION: non-checked sql, even drop table
		var rs []map[string]any
		tx := A.DB.Raw(sql).Scan(&rs)
		// Raw/Scan not support offset/limit

		result.Error = tx.Error
		result.Total = tx.RowsAffected
		// TODO: How to cast better?
		result.Rows = make([]Row, len(rs))
		for i := 0; i < len(rs); i++ {
			result.Rows[i] = Row(rs[i])
		}
	}
	A.Render(w, r, "templates/console.gotmpl", nil, map[string]any{
		"sql":    sql,
		"result": result,
	})
}
func (A *Admin) traceHandler(w http.ResponseWriter, r *http.Request) {
	m := map[string]any{"entries": nil}
	if A.trace != nil {
		m["entries"] = A.trace.Entries()
	}

	A.Render(w, r, "templates/trace.gotmpl", nil, m)
}
