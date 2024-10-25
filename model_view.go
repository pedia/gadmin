package gadmin

import (
	"html/template"
	"net/http"
	"reflect"
	"strconv"

	"github.com/samber/lo"
)

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

// TODO: ensure m not ptr
func NewModelView(m any, category ...string) *ModelView {
	model := new_model(m)

	cate := reflect.ValueOf(m).Type().Name()
	if len(category) > 0 {
		cate = category[0]
	}

	mv := ModelView{
		BaseView:               NewView(MenuItem{Name: cate}),
		model:                  model,
		can_create:             true,
		can_edit:               true,
		can_delete:             true,
		can_view_details:       true,
		can_export:             true, // false
		page_size:              20,
		can_set_page_size:      false,
		column_display_actions: true,
	}

	mv.Blueprint = &Blueprint{
		Name:     model.label(),
		Endpoint: model.name(),
		Path:     "/" + model.name(),
		Children: map[string]*Blueprint{
			// In flask-admin use `view.index`. Should use `view.index_view` in `gadmin`
			"index":        {Endpoint: "index", Path: "/", Handler: mv.index},
			"index_view":   {Endpoint: "index_view", Path: "/", Handler: mv.index},
			"create_view":  {Endpoint: "create_view", Path: "/new", Handler: mv.new},
			"details_view": {Endpoint: "details_view", Path: "/details", Handler: mv.details},
			"action_view":  {Endpoint: "action_view", Path: "/action", Handler: mv.index},
			"execute_view": {Endpoint: "execute_view", Path: "/execute", Handler: mv.index},
			"edit_view":    {Endpoint: "edit_view", Path: "/edit", Handler: mv.edit},
			"delete_view":  {Endpoint: "delete_view", Path: "/delete", Handler: mv.delete},
			// not .export_view
			"export": {Endpoint: "export", Path: "/export", Handler: mv.index},
			"debug":  {Endpoint: "debug", Path: "/debug", Handler: mv.debug},
		},
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

func (mv *ModelView) debug(w http.ResponseWriter, r *http.Request) {
	mv.Render(w, "debug.gotmpl", mv.dict(map[string]any{
		"menu":      mv.menu.dict(),
		"blueprint": mv.Blueprint.dict(),
	}))
}
func (mv *ModelView) index(w http.ResponseWriter, r *http.Request) {
	q := mv.queryFrom(r)

	total, data, err := mv.model.get_list(mv.admin.DB, q)
	_ = err // TODO: notify error

	num_pages := 1 + (total-1)/q.page_size
	// num_pages := math.Ceil(float64(total) / float64(q.limit))

	mv.Render(w, "model_list.gotmpl", mv.dict(
		map[string]any{
			"count":      len(data),
			"page":       q.page,
			"pages":      1,
			"num_pages":  num_pages,
			"return_url": must[string](mv.GetUrl(".index_view", q)),
			"pager_url": func(page int) string {
				return mv.GetUrl(".index_view", q, "page", page)
			},
			"page_size": q.page_size,
			"page_size_url": func(page_size int) string {
				return mv.GetUrl(".index_view", q, "page_size", page_size)
			},
			"can_set_page_size":        mv.can_set_page_size,
			"data":                     data,
			"request":                  rd(r),
			"get_pk_value":             mv.model.get_pk_value,
			"column_display_pk":        mv.column_display_pk,
			"column_display_actions":   mv.column_display_actions,
			"column_extra_row_actions": nil,
			"list_row_actions":         mv.list_row_actions(),
			"actions":                  []string{"delete", "Delete"}, // [('delete', 'Delete')]
			"list_columns":             mv.list_columns(),
			"is_sortable": func(name string) bool {
				_, ok := lo.Find(mv.column_sortable_list, func(s string) bool {
					return s == name
				})
				return ok
			},
			// in template, the sort url is: ?sort={index}
			"sort_column": q.sort_column(),
			"sort_desc":   q.sort_desc(),
			"sort_url": func(name string, invert ...bool) string {
				q := *q // simply copy
				if len(invert) > 0 && invert[0] {
					q.sort = Desc(name)
				} else {
					q.sort = Asc(name)
				}
				return mv.GetUrl(".index_view", &q)
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

func (mv *ModelView) queryFrom(r *http.Request) *query {
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
	actions := []action{}
	if mv.can_view_details {
		actions = append(actions, view_row_action)
	}
	if mv.can_edit {
		actions = append(actions, edit_row_action)
	}
	if mv.can_delete {
		actions = append(actions, delete_row_action)
	}
	return actions
}

func (mv *ModelView) new(w http.ResponseWriter, r *http.Request) {
	mv.Render(w, "model_create.gotmpl", mv.dict(map[string]any{
		"request": rd(r)},
	))
}
func (mv *ModelView) edit(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", contentTypeUtf8Html)
}
func (mv *ModelView) delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", contentTypeUtf8Html)
}
func (mv *ModelView) details(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", contentTypeUtf8Html)
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
	if err := mv.model.update(mv.admin.DB, pk, row); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(mv.admin.gettext("Failed to update record. %s", err)))
		return
	}
	w.Write([]byte(mv.admin.gettext("Record was successfully saved.")))
}

// request to dict
func rd(r *http.Request) map[string]any {
	return map[string]any{
		"method": r.Method,
		"url":    r.URL.String(),
		"args":   r.URL.Query(),
	}
}

func (mv *ModelView) get_form() model_form {
	return model_form{
		Fields: mv.model.columns,
	}
}
