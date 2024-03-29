package gadmin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"maps"
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
	Name  string // category
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
}

func NewAdmin(name string) *Admin {
	mux := http.NewServeMux()
	db, err := gorm.Open(
		sqlite.Open("../flask-admin/examples/sqla/admin/sample_db.sqlite"),
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

func (a *Admin) ts() *template.Template {
	fm := merge(sprig.FuncMap(), FuncsText)
	merge(fm, template.FuncMap{
		"admin_static_url": a.staticUrl, // used
		"marshal":          a.marshal,   // test
		"config":           a.config,    // used
		"gettext":          a.gettext,   //
		"csrf_token":       func() string { return "xxxx-csrf-token" },
	})

	t, err := template.New("all").Funcs(fm).ParseFiles(
		"templates/test.html",
		"templates/test_base.html",

		// "templates/layout.html",
		"templates/test_layout.html",
		"templates/master.html",
		"templates/base.html",
		"templates/index.html",
		"templates/test_lib.html",
		"templates/model_create.html",
		// "templates/layout.html",
		// "templates/static.html",
		// "templates/lib.html",
		// "templates/actions.html",
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(t.DefinedTemplates())
	return t
}

func (a *Admin) test(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")

	if err := a.ts().ExecuteTemplate(w, "test.html", map[string]any{
		"foo":   "bar",
		"empty": "",
		"null":  nil,
		"Zoo":   "Bar",
		"list":  []string{"a", "b"},
		"ss":    []struct{ A string }{{A: "a"}, {A: "b"}},
		"Conda": true,
		"Condb": false,

		// exposed map is better than struct
		"admin": a.dict(nil),

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
func (*Admin) staticUrl(filename, ver string) string {
	// TODO: /admin/static
	s := "/admin/static/" + filename
	if ver == "" {
		return s
	}
	return s + "?ver=" + ver
}

func (a *Admin) dict(other map[string]any) map[string]any {
	o := map[string]any{
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
	cloned := maps.Clone(o)
	cloned["admin"] = o
	return merge(cloned, other)
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
}

type BaseView struct {
	*Admin
	category string
	name     string
}

func (bv *BaseView) Category() string { return bv.category }
func (bv *BaseView) Name() string     { return bv.name }
func (bv *BaseView) Install(a *Admin) { bv.Admin = a }

func (bv *BaseView) HandleFunc(pattern string, f http.HandlerFunc) {
	p := fmt.Sprintf("/admin/%s%s", bv.name, pattern)
	if bv.name == "" {
		p = fmt.Sprintf("/admin%s", pattern)
	}
	bv.Mux.HandleFunc(p, f)
}

func (bv *BaseView) dict(others ...map[string]any) map[string]any {
	o := bv.Admin.dict(map[string]any{
		"category":  bv.category,
		"name":      bv.name,
		"extra_css": []string{},
		"extra_js":  []string{}, // "a.js", "b.js"}
	})

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func (bv *BaseView) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")

	if err := bv.ts().Lookup("index.html").Execute(w, bv.dict()); err != nil {
		panic(err)
	}
}

func NewView(category, name string) View {
	bv := BaseView{
		category: category,
		name:     name,
	}
	return &bv
}

type admin_index_view struct {
	*BaseView
}

func (aiv *admin_index_view) Install(admin *Admin) {
	aiv.BaseView.Install(admin)
	aiv.HandleFunc("/", aiv.index)
}
func (aiv *admin_index_view) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")

	if err := aiv.ts().
		Lookup("index.html").
		Execute(w, aiv.dict()); err != nil {
		panic(err)
	}
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
	DB    *gorm.DB // session
}

func NewModalView(m any, DB *gorm.DB) *ModelView {
	// TODO: package.name => name
	cate := reflect.ValueOf(m).Type().Name()
	return &ModelView{
		BaseView: &BaseView{category: cate, name: strings.ToLower(cate)},
		model:    m, // TODO: ptr to elem
		DB:       DB,
	}
}

func (mv *ModelView) dict(others ...map[string]any) map[string]any {
	o := mv.BaseView.dict(map[string]any{
		"return_url": "..", // TODO: /admin/user/
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
}
func (mv *ModelView) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")
}
func (mv *ModelView) new(w http.ResponseWriter, r *http.Request) {
	// can_create
	// create_form
	// form_args
	// form_widget_args
	w.Header().Add("content-type", "text/html; charset=utf-8")
	if err := mv.ts().
		Lookup("model_create.html").
		Execute(w, mv.dict(map[string]any{
			"form": mv.get_form().dict(),
			"form_opts": map[string]any{
				"widget_args": nil,
				// "form_rules":  nil, // form_create_rules
			},
			"action": nil,
		})); err != nil {
		panic(err)
	}
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

func (mv *ModelView) get_form() Form {
	s, err := schema.Parse(mv.model, &schemaStore, schema.NamingStrategy{})
	if err != nil {
		panic(err)
	}

	return Form{
		Fields: lo.Map(s.Fields, func(field *schema.Field, _ int) map[string]any {
			fmt.Printf("FieldType: %s %v\n", field.Name, field.FieldType)
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
				// validators
				"data": nil,
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
		"can_create": true,
		"can_edit":   true,
		"action":     "", // empty
		"hidden_tag": false,
		"fields":     f.Fields,
		"cancel_url": "TODO:cancel_url",
		"is_modal":   true,
	}
}
