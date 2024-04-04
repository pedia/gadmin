package gadmin

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
	"gorm.io/gorm/logger"
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
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			Logger:         logger.Default.LogMode(logger.Info),
		})
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

func (a *Admin) Run() {
	a.Handler = a.Mux
	l, _ := net.Listen("tcp", "127.0.0.1:3333")
	a.Serve(l)
}

func (a *Admin) add_view(v View) {
	cate, _ := v.Name()
	menu, ok := lo.Find(a.menus, func(item *Menu) bool {
		return item.Name == cate
	})
	if !ok {
		menu = &Menu{Name: cate, Views: []View{}}
		a.menus = append(a.menus, menu)
	}
	menu.Views = append(menu.Views, v)
}

func (a *Admin) AddView(v View) {
	a.add_view(v)

	v.install(a)

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
		"comment": func(s string) template.HTML {
			return template.HTML(
				"<!-- " + s + " -->",
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

	type foo struct {
		name  string
		Label string
	}
	type msa map[string]any

	if err := a.ts([]string{
		"templates/test.gotmpl",
	}...).ExecuteTemplate(w, "test.gotmpl", map[string]any{
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
		"bool1": func() bool { return false },
		"bool2": func() bool { return true },
		"rs":    func() foo { return foo{name: "Jerry", Label: "Label"} },
		"map":   func() map[string]any { return map[string]any{"name": "Jerry", "Label": "Label"} },
		"msa":   func() msa { return msa{"name": "Jerry", "Label": "Label"} },
		"msas":  func() []msa { return []msa{{"name": "Jerry", "Label": "Label"}} },
		"msas2": func() ([]msa, error) { return []msa{{"name": "Jerry", "Label": "Label"}}, nil },

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

// Like Flask.url_for
func (*Admin) getUrl(endpoint string, args ...any) string {
	// endpoint to path
	path, ok := map[string]string{
		".create_view": "new",
		".export":      "export/csv/", // TODO: pass to view.getUrl, get type from args
	}[endpoint]
	if ok {
		return path
	}
	return fmt.Sprintf("TODO %s?%v", endpoint, args)
}
func (*Admin) staticUrl(filename, ver string) string {
	s := "/admin/static/" + filename
	if ver != "" {
		s += "?ver=" + ver
	}
	return s
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

type View interface {
	// return category, name
	Name() (string, string)
	install(*Admin)
	dict(others ...map[string]any) map[string]any
}

type BaseView struct {
	*Admin
	category string
	name     string
}

func (bv *BaseView) Name() (string, string) { return bv.category, bv.name }
func (bv *BaseView) install(a *Admin) {
	bv.Admin = a
}

func (bv *BaseView) render(w http.ResponseWriter, page string, dict map[string]any) {
	w.Header().Add("content-type", "text/html; charset=utf-8")
	bases := []string{
		"templates/layout.gotmpl",
		"templates/master.gotmpl",
		"templates/base.gotmpl",
		"templates/lib.gotmpl",
		"templates/model_layout.gotmpl",
		"templates/actions.gotmpl",
	}
	bases = append(bases, "templates/"+page)
	if err := bv.ts(bases...).Lookup(page).Execute(w, dict); err != nil {
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

func (aiv *admin_index_view) install(admin *Admin) {
	aiv.BaseView.install(admin)
	aiv.HandleFunc("/", aiv.index)
}
func (aiv *admin_index_view) index(w http.ResponseWriter, r *http.Request) {
	aiv.render(w, "index.gotmpl", aiv.dict())
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
	model   model
	db      *gorm.DB // session
	Columns []string
}

func NewModalView(m any, DB *gorm.DB) *ModelView {
	// TODO: package.name => name
	cate := reflect.ValueOf(m).Type().Name()
	return &ModelView{
		BaseView: &BaseView{category: cate, name: strings.ToLower(cate)},
		model:    model{typo: reflect.TypeOf(m)}, // TODO: ptr to elem
		db:       DB,
	}
}

func (mv *ModelView) dict(others ...map[string]any) map[string]any {
	o := mv.BaseView.dict(map[string]any{
		"editable_columns": true,
		"can_create":       true,
		"can_edit":         true,
		"can_export":       true,
		"export_types":     []string{"csv", "xls"},
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
		"page_size_url": func() string {
			return "?page_size=2"
		},
	})

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}
func (mv *ModelView) install(admin *Admin) {
	mv.BaseView.install(admin)
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
	mv.render(w, "debug.gotmpl", mv.dict())
}
func (mv *ModelView) index(w http.ResponseWriter, r *http.Request) {
	data, err := mv.model.list(mv.db)
	_ = err
	mv.render(w, "model_list.gotmpl", mv.dict(
		map[string]any{
			"count":                    len(data),
			"page":                     1,
			"pages":                    1,
			"num_pages":                1,
			"pager_url":                "pager_url",
			"actions":                  nil,
			"data":                     data,
			"request":                  rd(r),
			"get_pk_value":             mv.model.get_pk_value,
			"column_display_pk":        false,
			"column_display_actions":   true,
			"column_extra_row_actions": nil,
			// flask_admin.model.template.ViewRowAction object at 0x112600f40&gt;, &lt;
			// flask_admin.model.template.EditRowAction object at 0x1126004c0&gt;, &lt;
			// flask_admin.model.template.DeleteRowAction object at 0x112600310&gt;] -->
			"list_row_actions":    nil,
			"list_columns":        mv.list_columns,
			"is_sortable":         func(v ...any) []string { return nil },
			"sort_column":         func(v any) any { return nil },
			"sort_url":            func(v ...any) string { return "?sort" },
			"is_editable":         func(v any) any { return nil },
			"column_descriptions": func(vs ...any) any { return nil },
			"get_value": func(m map[string]any, col column) any {
				return m[col["label"]]
			},
		},
	))
}

func (mv *ModelView) SetColumns(list []string) *ModelView {
	mv.Columns = list
	return mv
}

type column map[string]string

func (mv *ModelView) list_columns(_ ...any) ([]column, error) {
	s, err := schema.Parse(mv.model.new(), &schemaStore, schema.NamingStrategy{})
	if err != nil {
		return nil, err
	}

	fs := lo.Filter(s.Fields, func(f *schema.Field, _ int) bool {
		_, ok := lo.Find(mv.Columns, func(c string) bool {
			return c == f.DBName
		})
		return ok
	})

	cs := lo.Map(fs, func(f *schema.Field, _ int) column {
		return column{
			"label": strings.Title(f.Name),
			"name":  f.DBName,
		}
	})
	return cs, nil
}

func (mv *ModelView) new(w http.ResponseWriter, r *http.Request) {
	mv.render(w, "model_create.gotmpl", mv.dict(map[string]any{
		"request": rd(r)},
	))
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

// request to dict
func rd(r *http.Request) map[string]any {
	return map[string]any{
		"method": r.Method,
		"url":    r.URL.String(),
		"args":   r.URL.Query(),
	}
}

var schemaStore = sync.Map{}

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

type model struct {
	typo reflect.Type
}

// new t
func (m *model) new() any {
	return reflect.New(m.typo).Interface()
}

// new []t
func (m *model) new_slice() reflect.Value {
	return reflect.New(reflect.SliceOf(m.typo))
}

func (m *model) get_pk_value(row any) any {
	return 1
}
func (m *model) list(db *gorm.DB) ([]map[string]any, error) {
	ptr := m.new_slice()
	if err := db.Find(ptr.Interface()).Error; err != nil {
		return nil, err
	}

	// better way?
	len := ptr.Elem().Len()
	res := make([]any, len)
	for i := 0; i < len; i++ {
		item := ptr.Elem().Index(i).Interface()
		res[i] = item
	}

	// _, ok := ptr.Interface().([]any)
	// fmt.Printf("conv %v", ok)

	bs, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	var ms []map[string]any
	if err := json.Unmarshal(bs, &ms); err != nil {
		return nil, err
	}
	return ms, nil
}
