package gadmin

import (
	"html/template"
	"net/http"
	"reflect"
	"strconv"

	"github.com/go-playground/form/v4"
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
	model := newModel(m)

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
			"action_view":  {Endpoint: "action_view", Path: "/action", Handler: mv.ajax_update},
			"execute_view": {Endpoint: "execute_view", Path: "/execute", Handler: mv.index},
			"edit_view":    {Endpoint: "edit_view", Path: "/edit", Handler: mv.edit},
			"delete_view":  {Endpoint: "delete_view", Path: "/delete", Handler: mv.delete},
			// not .export_view
			"export": {Endpoint: "export", Path: "/export", Handler: mv.index},
			"debug":  {Endpoint: "debug", Path: "/debug", Handler: mv.debug},
			// for json
			"list": {Endpoint: "list", Path: "/list", Handler: mv.list},
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

func (mv *ModelView) dict(r *http.Request, others ...map[string]any) map[string]any {
	o := mv.BaseView.dict(r, map[string]any{
		"table_prefix_html": mv.table_prefix_html,
		"editable_columns":  true,
		"can_create":        mv.can_create,
		"can_edit":          mv.can_edit,
		"can_export":        mv.can_export,
		"can_view_details":  mv.can_view_details,
		"can_delete":        mv.can_delete,
		"export_types":      []string{"csv", "xls"},
		// TODO: modal for edit/create/details
		"edit_modal":    false,
		"create_modal":  false,
		"details_modal": false,
		"form":          mv.get_form().dict(),
		"form_opts": map[string]any{
			"widget_args": nil,
			"form_rules":  []any{},
		},
		"filters":              []string{},
		"filter_groups":        []string{},
		"actions_confirmation": mv.list_row_actions_confirmation(),
		"search_supported":     false,
		"return_url":           mv.GetUrl(".index_view", nil),
	})

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func (mv *ModelView) debug(w http.ResponseWriter, r *http.Request) {
	mv.Render(w, r, "debug.gotmpl", nil, map[string]any{
		"menu":      mv.menu.dict(),
		"blueprint": mv.Blueprint.dict(),
	})
}
func (mv *ModelView) index(w http.ResponseWriter, r *http.Request) {
	if mv.admin.debug {
		Flash(r, "hello")
		Flash(r, "Worked", "success")
		Flash(r, "Caution", "danger")
	}

	q := mv.queryFrom(r)

	total, data, err := mv.model.get_list(r.Context(), mv.admin.DB, q)
	_ = err // TODO: messages

	q.setTotal(total)

	mv.Render(w, r, "model_list.gotmpl", nil, map[string]any{
		"count":     len(data),
		"page":      q.Page,
		"num_pages": q.num_pages,
		"page_size": q.PageSize,
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
		// in template, `sort url` is: ?sort={index}
		// transform `index` to `column name`
		"sort_column": func() string {
			if q.Sort != "" {
				idx := must[int](strconv.Atoi(q.Sort))
				if idx != -1 {
					return mv.column_list[idx]
				}
			}
			return ""
		}(),
		"sort_desc": q.Desc,
		"sort_url": func(name string, invert ...bool) string {
			q := *q // simply copy
			q.Sort = strconv.Itoa(mv.get_column_index(name))
			q.Desc = firstOr(invert)
			return mv.GetUrl(".index_view", &q)
		},
		"is_editable": mv.is_editable,
		"column_descriptions": func(name string) string {
			if desc, ok := mv.column_descriptions[name]; ok {
				return desc
			}
			return mv.model.find(name)["description"].(string)
		},
		"list_form": mv.list_form,
	})
}
func (mv *ModelView) list(w http.ResponseWriter, r *http.Request) {
	q := mv.queryFrom(r)

	total, data, err := mv.model.get_list(r.Context(), mv.admin.DB, q)
	if err != nil {
		ReplyJson(w, 200, map[string]any{"error": err.Error()})
		return
	}
	ReplyJson(w, 200, map[string]any{"total": total, "data": data})
}

func (mv *ModelView) queryFrom(r *http.Request) *Query {
	q := Query{default_page_size: mv.page_size}
	uv := r.URL.Query()

	form.NewDecoder().Decode(&q, uv)
	for k, v := range uv {
		if lo.IndexOf([]string{"page", "page_size", "sort", "desc", "search"}, k) != -1 {
			continue
		}
		q.args = append(q.args, k, v[0])
	}
	return &q
}

func (mv *ModelView) list_columns() []column {
	return lo.Filter(mv.model.columns, func(col column, _ int) bool {
		// in `column_list`
		_, ok := lo.Find(mv.column_list, func(c string) bool {
			return c == col.name()
		})

		// not in `column_exclude_list`
		_, exclude := lo.Find(mv.column_exclude_list, func(c string) bool {
			return c == col.name()
		})
		return ok && !exclude
	})
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

func (mv *ModelView) list_row_actions_confirmation() map[string]string {
	res := map[string]string{}
	for _, a := range mv.list_row_actions() {
		if c, ok := a["confirmation"]; ok {
			res[a["name"].(string)] = c.(string)
		}
	}
	return res
}

// row -> Model().Create() RETURNING *
func (mv *ModelView) new(w http.ResponseWriter, r *http.Request) {
	mv.Render(w, r, "model_create.gotmpl", nil, map[string]any{
		"request": rd(r)})
}

func (mv *ModelView) edit(w http.ResponseWriter, r *http.Request) {
}

// Model().Where(pk field = pk value).Delete()
func (mv *ModelView) delete(w http.ResponseWriter, r *http.Request) {
}

// Model().Where(pk field = pk value).First()
func (mv *ModelView) details(w http.ResponseWriter, r *http.Request) {
	q := mv.queryFrom(r)
	redirect := func() {
		url := q.Get("url")
		if url == "" {
			url = mv.GetUrl(".index_view", nil)
		}
		http.Redirect(w, r, url, http.StatusFound)
	}

	if !mv.can_view_details {
		redirect()
		return
	}

	one, err := mv.model.get_one(r.Context(), mv.admin.DB, q.Get("id"))
	if err != nil {
		Flash(r, gettext("Record does not exist."), "danger")

		redirect()
		return
	}

	mv.Render(w, r, "model_details.gotmpl", nil, map[string]any{
		"model":           one,
		"details_columns": mv.list_columns(),
		"request":         rd(r),
	})
	_ = q
}

// list_form_pk=a1d13310-7c10-48d5-b63b-3485995ad6a4&currency=USD
// Record was successfully saved.
// Model().Where().Update(currency=USD)
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

// request to dict, like flask.request
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

func (mv *ModelView) Render(w http.ResponseWriter, r *http.Request, name string, funcs template.FuncMap, data map[string]any) {
	w.Header().Add("content-type", ContentTypeUtf8Html)
	fs := []string{
		"templates/actions.gotmpl",
		"templates/base.gotmpl",
		"templates/layout.gotmpl",
		"templates/lib.gotmpl",
		"templates/master.gotmpl",
		"templates/model_layout.gotmpl",
		"templates/model_row_actions.gotmpl",
	}

	fm := template.FuncMap{
		"return_url": func() (string, error) {
			return mv.admin.GetUrl(mv.Endpoint+".index_view", nil)
		},
		"get_flashed_messages": func() []map[string]any {
			return FlashedFrom(r).GetMessages()
		},
		"get_url": func(endpoint string, args ...any) (string, error) {
			return mv.GetUrl(endpoint, nil, args...), nil
		},
		"get_value": func(m map[string]any, col column) any {
			return m[col.name()]
		},
		"page_size_url": func(page_size int) string {
			return mv.GetUrl(".index_view", nil, "page_size", page_size)
		},
		"pager_url": func(page int) string {
			return mv.GetUrl(".index_view", nil, "page", page)
		},
	}
	if funcs != nil {
		merge(fm, funcs)
	}

	fs = append(fs, "templates/"+name)
	if err := createTemplate(fs, mv.admin.funcs(fm)).
		ExecuteTemplate(w, name, mv.dict(r, data)); err != nil {
		panic(err)
	}
}
