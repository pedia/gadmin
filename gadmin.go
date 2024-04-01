package gadmin

// Q:
// - template result to arg
// - request.args

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/Masterminds/sprig/v3"
	"github.com/glebarez/sqlite"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type Menu struct {
	Name  string // category?
	Views []View
	class string
}

// merge two maps
func merge[K comparable, V any](a, b map[K]V) map[K]V {
	for k, v := range b {
		a[k] = v
	}
	return a
}

func (m *Menu) dict() map[string]any {
	return map[string]any{
		"name":  m.Name,
		"class": m.class,
		"icon":  "TODO",
	}
}

type Admin struct {
	name string
	*http.Server
	menus []*Menu // TODO: ordered_map
	Mux   *http.ServeMux
	DB    *gorm.DB
	debug bool
}

func NewAdmin(name string) *Admin {
	mux := http.NewServeMux()
	db, err := gorm.Open(
		sqlite.Open("examples/sqla/admin/sample_db.sqlite"),
		&gorm.Config{NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		panic(err)
	}
	a := Admin{
		name:   name,
		Server: &http.Server{Handler: mux},
		menus:  []*Menu{},
		Mux:    mux,
		DB:     db,
		debug:  true,
	}

	// Home -> /admin
	// a.AddView(NewView("", "Home"))
	a.AddView(NewAdminIndex())

	// a.Router.HandleFunc("/admin/{$}", a.index)
	a.Mux.HandleFunc("/admin/test", a.test)

	// serve in Admin.staticUrl
	// /admin/static/{} => static/{}
	a.Mux.HandleFunc("/admin/static/{path...}",
		func(w http.ResponseWriter, r *http.Request) {
			path := r.PathValue("path")
			fmt.Printf("static: %s", path)
			http.ServeFileFS(w, r, os.DirFS("static"), path)
		})
	return &a
}

func (a *Admin) add_view(v View) {
	cate, ok := lo.Find(a.menus, func(item *Menu) bool {
		return item.Name == v.Category()
	})
	if !ok {
		cate = &Menu{Name: v.Category(), Views: []View{}}
		a.menus = append(a.menus, cate)
	}
	cate.Views = append(cate.Views, v)
}

func (a *Admin) AddView(v View) {
	a.add_view(v)

	v.Install(a)

	// /admin/					=> admin_index_view.index
	// /admin/{modal}        	=> model_view.index
	// /admin/{modal}/create 	=> model_view.create

	// for k, f := range v.Router() {
	// 	a.Router.Handle(fmt.Sprintf("/admin%s%s", v.Name(), k), f)
	// }
}

func (a *Admin) ts(fs ...string) *template.Template {
	fm := merge(sprig.FuncMap(), FuncsText)
	merge(fm, template.FuncMap{
		"admin_static_url": a.staticUrl, // used
		"get_url":          a.getUrl,
		"marshal":          a.marshal, // test
		"config":           a.config,  // used
		"gettext":          a.gettext, //
		"csrf_token":       func() string { return "xxxx-csrf-token" },
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

	t, err := template.New("all").
		Option("missingkey=error").
		Funcs(fm).ParseFiles(fs...)
	if err != nil {
		panic(err)
	}

	// fmt.Println(t.DefinedTemplates())
	return t
}

func (a *Admin) test(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")

	if err := a.ts().ExecuteTemplate(w, "test.gotmpl", map[string]any{
		"foo":   "bar",
		"empty": "",
		"null":  nil,
		"Zoo":   "Bar",
		"list":  []string{"a", "b"},
		"ss":    []struct{ A string }{{A: "a"}, {A: "b"}},
		"ls":    []map[string]any{{"A": "a"}, {"B": "b"}},
		"conda": true,
		"condb": false,
		"int":   34,

		// map is better than struct
		"admin": a.dict(),

		// bad
		// {{ .admin_static.Url x y}}
		"admin_static": a,

		// {{ .admin_static_url x y}}
		"admin_static_url": a.staticUrl,
	}); err != nil {
		fmt.Println(err)
	}
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
func (*Admin) gettext(key string) string {
	return key
}

func (*Admin) getUrl(vs ...any) string {
	t := vs[0].(string)
	path, ok := map[string]string{
		".create_view": "new",
	}[t]
	if ok {
		return path
	}
	return fmt.Sprintf("TODO %v", vs)
}
func (*Admin) staticUrl(filename, ver string) string {
	// TODO: /admin/static
	s := "/admin/static/" + filename
	if ver == "" {
		return s
	}
	return s + "?ver=" + ver
}

func (a *Admin) dict() map[string]any {
	return map[string]any{
		"debug":               a.debug,
		"name":                a.name,
		"url":                 "/admin",
		"admin_base_template": "base.html",
		"swatch":              "cerulean", // "default",
		// {{ .admin_static_url x y }}
		// "admin_static_url": a.staticUrl,
		"menus": lo.Map(a.menus, func(m *Menu, _ int) map[string]any {
			return m.dict()
		}),
	}
}

func (a *Admin) Run() {
	a.Handler = a.Mux
	l, _ := net.Listen("tcp", "127.0.0.1:3333")
	a.Serve(l)
}

type View interface {
	Category() string
	Name() string
	Install(*Admin)
	dict(others ...map[string]any) map[string]any
}

type BaseView struct {
	*Admin
	category string
	name     string
}

func (bv *BaseView) Category() string { return bv.category }
func (bv *BaseView) Name() string     { return bv.name }
func (bv *BaseView) Install(a *Admin) {
	bv.Admin = a
}

func (bv *BaseView) render(w http.ResponseWriter, page string, fs []string, dict map[string]any) {
	w.Header().Add("content-type", "text/html; charset=utf-8")
	if err := bv.ts(fs...).Lookup(page).Execute(w, dict); err != nil {
		panic(err)
	}
}

func (bv *BaseView) HandleFunc(pattern string, f http.HandlerFunc) {
	p := fmt.Sprintf("/admin/%s%s", bv.name, pattern)
	if bv.name == "" {
		p = fmt.Sprintf("/admin%s", pattern)
	}
	bv.Mux.HandleFunc(p, f)
}

func (bv *BaseView) dict(others ...map[string]any) map[string]any {
	o := map[string]any{
		"category":           bv.category,
		"name":               bv.name,
		"extra_css":          []string{},
		"extra_js":           []string{}, // "a.js", "b.js"}
		"admin":              bv.Admin.dict(),
		"admin_fluid_layout": true,
	}

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func (bv *BaseView) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")

	if err := bv.ts().Lookup("index.gotmpl").Execute(w, bv.dict()); err != nil {
		panic(err)
	}
}

func NewView(category, name string) View {
	return &BaseView{
		category: category,
		name:     name,
	}
}

type admin_index_view struct {
	*BaseView
}

func (aiv *admin_index_view) Install(admin *Admin) {
	aiv.BaseView.Install(admin)
	aiv.HandleFunc("/", aiv.index)
}
func (aiv *admin_index_view) index(w http.ResponseWriter, r *http.Request) {
	aiv.render(w, "index.gotmpl", []string{}, aiv.dict())
}
func NewAdminIndex() *admin_index_view {
	aiv := admin_index_view{
		BaseView: &BaseView{
			category: "Home",
			name:     "",
		},
	}
	return &aiv
}

type ModelView struct {
	*BaseView
	model any
	db    *gorm.DB // session
}

func NewModalView(m any, DB *gorm.DB) *ModelView {
	// TODO: package.name => name
	cate := reflect.ValueOf(m).Type().Name()
	return &ModelView{
		BaseView: &BaseView{category: cate, name: strings.ToLower(cate)},
		model:    m, // TODO: ptr to elem
		db:       DB,
	}
}

func (mv *ModelView) dict(others ...map[string]any) map[string]any {
	o := mv.BaseView.dict(map[string]any{
		"editable_columns": true,
		"can_create":       true,
		"can_edit":         true,
		"can_export":       false,
		"edit_modal":       false,
		"create_modal":     false,
		"details_modal":    false,
		"return_url":       "../",
		"form":             mv.get_form().dict(),
		"form_opts": map[string]any{
			"widget_args": nil,
			"form_rules":  []any{},
		},
		"filters":                []string{},
		"filter_groups":          []string{},
		"can_set_page_size":      false, // TODO:
		"actions":                []string{},
		"actions_confirmation":   []string{},
		"search_supported":       false,
		"column_display_actions": true,
		"list_columns":           []string{},
		"page_size_url": func() string {
			return "?page_size=2"
		},
	})

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}
func (mv *ModelView) Install(admin *Admin) {
	mv.BaseView.Install(admin)
	mv.HandleFunc("/", mv.index)
	mv.HandleFunc("/new/", mv.new)
	mv.HandleFunc("/edit/", mv.edit)
	mv.HandleFunc("/details/", mv.details)
	mv.HandleFunc("/delete/", mv.index)
	mv.HandleFunc("/action/", mv.index)
	mv.HandleFunc("/export/{export_type}", mv.index)
	mv.HandleFunc("/ajax/lookup/", mv.index)
	mv.HandleFunc("/ajax/update/", mv.index)
	if mv.Admin.debug {
		mv.HandleFunc("/debug/", mv.debug)
	}
}
func (mv *ModelView) debug(w http.ResponseWriter, r *http.Request) {
	mv.render(w, "debug.gotmpl", []string{}, mv.dict())
}
func (mv *ModelView) index(w http.ResponseWriter, r *http.Request) {
	mv.render(w, "model_list.gotmpl", []string{
		"templates/layout.gotmpl",
		"templates/master.gotmpl",
		"templates/base.gotmpl",
		"templates/lib.gotmpl",
		"templates/model_layout.gotmpl",
		"templates/actions.gotmpl",
		"templates/model_list.gotmpl",
	}, mv.dict(
		map[string]any{
			"count":     2,
			"page":      1,
			"pages":     1,
			"num_pages": 1,
			"pager_url": "pager_url",
			"data":      []any{},
		},
	))
}
func (mv *ModelView) new(w http.ResponseWriter, r *http.Request) {
	mv.render(w, "model_create.gotmpl", []string{
		"templates/layout.gotmpl",
		"templates/master.gotmpl",
		"templates/base.gotmpl",
		"templates/lib.gotmpl",
		"templates/model_create.gotmpl",
		"templates/model_layout.gotmpl",
		"templates/actions.gotmpl",
	}, mv.dict(rd(r)))
}
func (mv *ModelView) edit(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")
}
func (mv *ModelView) delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")
}
func (mv *ModelView) details(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")
}

var schemaStore = sync.Map{}

// request's dict
func rd(r *http.Request) map[string]any {
	return map[string]any{
		"request": map[string]any{
			"args": r.URL.RawQuery,
		},
	}
}

func (mv *ModelView) get_form() Form {
	s, err := schema.Parse(mv.model, &schemaStore, schema.NamingStrategy{})
	if err != nil {
		panic(err)
	}

	return Form{
		Fields: lo.Map(s.Fields, func(field *schema.Field, _ int) map[string]any {
			// fmt.Printf("FieldType: %s %v\n", field.Name, field.FieldType)
			return map[string]any{
				"id":          field.DBName, //
				"name":        field.DBName, //
				"description": field.Comment,
				"required":    field.NotNull,
				"choices":     nil,
				// "type": "StringField",
				"label":  strings.Title(field.Name),
				"widget": field2widget(field),
				"errors": nil,
			}
		}),
	}
}

// TODO: more field type
func field2widget(field *schema.Field) map[string]any {
	table := map[reflect.Kind]string{
		reflect.String: "text",
		// reflect.:"password",
		// reflect.Kind:"hidden",
		reflect.Bool: "checkbox",
		// reflect.Kind:"radio",
		// reflect.Kind:"file",
		// reflect.Kind:"submit",
	}

	return map[string]any{
		"input_type": table[field.FieldType.Kind()],
	}
}

// @expose('/')
// @expose('/new/', methods=('GET', 'POST'))
// @expose('/edit/', methods=('GET', 'POST'))
// @expose('/details/')
// @expose('/delete/', methods=('POST',))
// @expose('/action/', methods=('POST',))
// @expose('/export/<export_type>/')
// @expose('/ajax/lookup/')
// @expose('/ajax/update/', methods=('POST',))

type Form struct {
	Fields      []map[string]any
	Prefix      string
	ExtraFields []string
}

func (f Form) dict() map[string]any {
	return map[string]any{
		"action":     "", // empty
		"hidden_tag": false,
		"fields":     f.Fields,
		"cancel_url": "TODO:cancel_url",
		"is_modal":   true,
		"csrf_token": true,
	}
}
