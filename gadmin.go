package gadmin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
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

	// serve Admin.staticUrl
	// url /admin/static/{} => local static/{}
	// first way:
	fs := http.FileServer(http.Dir("static"))
	a.Mux.Handle("/admin/static/", http.StripPrefix("/admin/static/", fs))

	// second way:
	// a.Mux.HandleFunc("/admin/static/{path...}",
	// 	func(w http.ResponseWriter, r *http.Request) {
	// 		path := r.PathValue("path")
	// 		http.ServeFileFS(w, r, os.DirFS("static"), path)
	// 	})
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
		"get_url": func(endpoint string, args ...map[string]any) (string, error) {
			if len(args) == 0 {
				args = []map[string]any{{}}
			}
			return a.urlFor("", endpoint, args[0])
		},
		"marshal":    a.marshal, // test
		"config":     a.config,  // used
		"gettext":    a.gettext, //
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

		"request": rd(r),

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
func (*Admin) gettext(format string, a ...any) string {
	return gettext(format, a...)
}

func gettext(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

// Like Flask.url_for
func (*Admin) urlFor(model, endpoint string, args map[string]any) (string, error) {
	// endpoint to path
	if pos := strings.Index(endpoint, "."); pos != 0 {
		model = endpoint[:pos]
		endpoint = endpoint[pos:]
	}
	path, ok := map[string]string{
		".index_view":   "",
		".create_view":  "new",
		".details_view": "details",
		".action_view":  "action",
		".execute_view": "execute",
		".edit_view":    "edit",
		".delete_view":  "delete",
		".export":       "export",
	}[endpoint]
	if !ok {
		return "", fmt.Errorf(`endpoint "%s" not found`, endpoint)
	}

	// apply custom args
	uv := map_into_values(args)
	if model != "" {
		path = "/admin/" + model + "/"
	} else {
		if path != "" {
			path = path + "/" // TODO: use URL.JoinPath
		}
	}
	if len(uv) > 0 {
		return path + "?" + uv.Encode(), nil
	}
	return path, nil
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
		"templates/model_row_actions.gotmpl",
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
	model *model

	// Permissions
	can_create       bool
	can_edit         bool
	can_delete       bool
	can_view_details bool
	can_export       bool

	// Customizations
	column_list          []string
	column_exclude_list  []string
	column_editable_list []string
	column_sortable_list []string
	column_descriptions  map[string]string

	table_prefix_html string

	// Pagination settings
	page_size         int
	can_set_page_size bool
	column_filters    []string

	column_display_pk      bool
	column_display_actions bool
	form_columns           []string
	form_excluded_columns  []string

	//
	list_forms []base_form
}

func NewModalView(m any) *ModelView {
	// TODO: package.name => name
	cate := reflect.ValueOf(m).Type().Name()
	mv := ModelView{
		BaseView:               &BaseView{category: cate, name: strings.ToLower(cate)},
		model:                  new_model(m), // TODO: ptr to elem
		can_create:             true,
		can_edit:               true,
		can_delete:             true,
		can_view_details:       true,
		can_export:             true,
		page_size:              20,
		can_set_page_size:      false,
		column_display_actions: true,
	}

	mv.column_list = lo.Map(mv.model.columns, func(col column, _ int) string {
		return col.name()
	})
	mv.column_sortable_list = mv.model.sortable_list()
	return &mv
}

// Permissions
// Is model creation allowed
func (mv *ModelView) SetCanCreate(v bool) *ModelView {
	mv.can_create = v
	return mv
}

// Is model editing allowed
func (mv *ModelView) SetCanEdit(v bool) *ModelView {
	mv.can_edit = v
	return mv
}
func (mv *ModelView) SetCanExport(v bool) *ModelView {
	mv.can_export = v
	return mv
}

// Collection of the model field names for the list view.
// If not set, will get them from the model.
func (mv *ModelView) SetColumnList(vs ...string) *ModelView {
	mv.column_list = vs
	return mv
}

func (mv *ModelView) SetColumnEditableList(vs ...string) *ModelView {
	mv.column_editable_list = vs
	// build list_forms here
	mv.list_forms = []base_form{}
	return mv
}
func (mv *ModelView) SetColumnDescriptions(m map[string]string) *ModelView {
	mv.column_descriptions = m
	return mv
}
func (mv *ModelView) SetTablePrefixHtml(v string) *ModelView {
	mv.table_prefix_html = v
	return mv
}
func (mv *ModelView) SetCanSetPageSize(v bool) *ModelView {
	mv.can_set_page_size = v
	return mv
}
func (mv *ModelView) SetPageSize(v int) *ModelView {
	mv.page_size = v
	return mv
}

func (mv *ModelView) urlFor(endpoint string, args ...map[string]any) (string, error) {
	arg := map[string]any{}
	if len(args) > 0 {
		arg = args[0]
	}
	return mv.Admin.urlFor(mv.model.name(), endpoint, arg)
}

func (mv *ModelView) dict(others ...map[string]any) map[string]any {
	o := mv.BaseView.dict(map[string]any{
		"table_prefix_html": mv.table_prefix_html,
		"editable_columns":  true,
		"can_create":        mv.can_create,
		"can_edit":          mv.can_edit,
		"can_export":        mv.can_export,
		"can_view_details":  mv.can_view_details,
		"can_delete":        mv.can_delete,
		"export_types":      []string{"csv", "xls"},
		"edit_modal":        false,
		"create_modal":      false,
		"details_modal":     false,
		"form":              mv.get_form().dict(),
		"form_opts": map[string]any{
			"widget_args": nil,
			"form_rules":  []any{},
		},
		"filters":              []string{},
		"filter_groups":        []string{},
		"actions_confirmation": []string{},
		"search_supported":     false,
		// will replace
		"return_url": "../",
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
	mv.HandleFunc("/ajax/update/", mv.ajax_update)
	if mv.Admin.debug {
		mv.HandleFunc("/debug/", mv.debug)
	}
}
func (mv *ModelView) debug(w http.ResponseWriter, r *http.Request) {
	mv.render(w, "debug.gotmpl", mv.dict())
}
func (mv *ModelView) index(w http.ResponseWriter, r *http.Request) {
	q := mv.query_from(r)

	total, data, err := mv.model.get_list(mv.DB, q)
	_ = err // TODO: notify error

	num_pages := 1 + (total-1)/q.page_size
	// num_pages := math.Ceil(float64(total) / float64(q.limit))

	mv.render(w, "model_list.gotmpl", mv.dict(
		map[string]any{
			"count":      len(data),
			"page":       q.page,
			"pages":      1,
			"num_pages":  num_pages,
			"return_url": must[string](mv.urlFor(".index_view", mv.query_into_args(q))),
			"pager_url": func(page int) (string, error) {
				args := mv.query_into_args(q)
				args["page"] = page
				return mv.urlFor(".index_view", args)
			},
			"page_size": q.page_size,
			"page_size_url": func(page_size int) (string, error) {
				args := mv.query_into_args(q)
				args["page_size"] = page_size
				return mv.urlFor(".index_view", args)
			},
			"can_set_page_size":        mv.can_set_page_size,
			"actions":                  []string{},
			"data":                     data,
			"request":                  rd(r),
			"get_pk_value":             mv.model.get_pk_value,
			"column_display_pk":        mv.column_display_pk,
			"column_display_actions":   mv.column_display_actions,
			"column_extra_row_actions": nil,
			"list_row_actions":         mv.list_row_actions,
			"list_columns":             mv.list_columns,
			"is_sortable": func(name string) bool {
				_, ok := lo.Find(mv.column_sortable_list, func(s string) bool {
					return s == name
				})
				return ok
			},
			// in template, the sort url is: ?sort={index}
			"sort_column": q.sort_column(),
			"sort_desc":   q.sort_desc(),
			"sort_url": func(name string, invert ...bool) (string, error) {
				q := *q // simply copy
				if len(invert) > 0 && invert[0] {
					q.sort = Desc(name)
				} else {
					q.sort = Asc(name)
				}
				args := mv.query_into_args(&q)
				return mv.urlFor(".index_view", args)
			},
			"is_editable": mv.is_editable,
			"column_descriptions": func(name string) string {
				if desc, ok := mv.column_descriptions[name]; ok {
					return desc
				}
				return mv.get_column(name)["description"].(string)
			},
			"get_value": func(m map[string]any, col column) any {
				return m[col.name()]
			},
			"list_form": mv.list_form,
		},
	))
}

func (mv *ModelView) query(q *query) url.Values {
	args := url.Values{}
	if q.page > 0 {
		args.Set("page", fmt.Sprintf("%d", q.page))
	}
	args.Set("page_size", fmt.Sprintf("%d", q.page_size))
	if q.sort.Name != "" {
		args.Set("sort", fmt.Sprintf("%d", mv.get_column_index(q.sort.Name)))
		if q.sort.Desc == 1 {
			args.Set("desc", "1")
		}
	}
	return args
}

func (mv *ModelView) query_into_args(q *query) map[string]any {
	args := map[string]any{}
	if q.page > 0 {
		args["page"] = strconv.Itoa(q.page)
	}
	args["page_size"] = strconv.Itoa(q.page_size)
	if q.sort.Name != "" {
		args["sort"] = strconv.Itoa(mv.get_column_index(q.sort.Name))
		if q.sort.Desc == 1 {
			args["desc"] = "1"
		}
	}
	return args
}

func (mv *ModelView) query_from(r *http.Request) *query {
	q := r.URL.Query()

	// ?sort=0&desc=1
	var sort_column string
	if q.Has("sort") {
		i := must[int](strconv.Atoi(q.Get("sort")))
		sort_column = mv.column_list[i]
	}

	var sort_desc int
	if q.Has("desc") {
		sort_desc = 1
	}

	o := Asc("")
	if sort_column != "" {
		if sort_desc == 0 {
			o = Asc(sort_column)
		} else {
			o = Desc(sort_column)
		}
	}

	limit := mv.page_size
	if mv.can_set_page_size && q.Has("page_size") {
		i := must[int](strconv.Atoi(q.Get("page_size")))
		limit = i
	}

	var page int
	if q.Has("page") {
		i := must[int](strconv.Atoi(q.Get("page")))
		page = i
	}
	return &query{
		page:      page,
		page_size: limit,
		sort:      o,
	}
}

func (mv *ModelView) list_columns() []column {
	return lo.Filter(mv.model.columns, func(col column, _ int) bool {
		// in column_list
		ok := true
		if len(mv.column_list) > 0 {
			_, ok = lo.Find(mv.column_list, func(c string) bool {
				return c == col.name()
			})
		}
		// not in column_exclude_list
		_, exclude := lo.Find(mv.column_exclude_list, func(c string) bool {
			return c == col.name()
		})
		return ok && !exclude
	})
}

func (mv *ModelView) get_column(name string) column {
	if col, ok := lo.Find(mv.model.columns, func(col column) bool {
		return col.name() == name
	}); ok {
		return col
	}
	return nil
}
func (mv *ModelView) get_column_index(name string) int {
	if _, i, ok := lo.FindIndexOf(mv.column_list, func(c string) bool {
		return c == name
	}); ok {
		return i
	}
	return -1
}
func (mv *ModelView) list_form(col column, r row) template.HTML {
	x := XEditableWidget{model: mv.model, column: col}
	return x.html(r)
}
func (mv *ModelView) is_editable(name string) bool {
	if !mv.can_edit {
		return false
	}
	_, ok := lo.Find(mv.column_editable_list, func(i string) bool {
		return i == name
	})
	return ok
}

func (mv *ModelView) list_row_actions() []action {
	as := []action{}
	if mv.can_view_details {
		as = append(as, view_row_action())
	}
	if mv.can_edit {
		as = append(as, edit_row_action())
	}
	if mv.can_delete {
		as = append(as, delete_row_action())
	}
	return as
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

// list_form_pk=a1d13310-7c10-48d5-b63b-3485995ad6a4&currency=USD
// Record was successfully saved.
func (mv *ModelView) ajax_update(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	if len(mv.column_editable_list) == 0 {
		w.WriteHeader(404)
		return
	}

	// form
	r.ParseForm()
	pk := r.Form.Get("list_form_pk")

	// TODO: pk to int

	// TODO: type list_form struct, parse
	row := row{}
	for k, v := range r.Form {
		if k == "list_form_pk" {
			continue
		}
		row[k] = v[0]
	}

	// validate
	// get_one
	// record, err := mv.model.get(mv.DB, pk)
	// if err == gorm.ErrRecordNotFound {
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(mv.gettext("Record does not exist.")))
	// 	return
	// }
	// _ = record

	// update_model
	if err := mv.model.update(mv.DB, pk, row); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(mv.gettext("Failed to update record. %s", err)))
		return
	}
	w.Write([]byte(mv.gettext("Record was successfully saved.")))
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

func (mv *ModelView) get_form() model_form {
	return model_form{
		Fields: mv.model.columns,
	}
}
