package gadmin

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-playground/form/v4"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ModelView struct {
	*BaseView
	*Model

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

	joins      []queryArg
	innerJoins []queryArg
	preloads   []queryArg
}

// for db.Joins(query string, args... any)
type queryArg struct {
	query string
	args  []any
}

// TODO: ensure m not ptr
func NewModelView(m any, category ...string) *ModelView {
	model := NewModel(m)

	cate := firstOr(category, model.label())

	mv := ModelView{
		BaseView:               NewView(MenuItem{Name: model.label(), Category: cate}),
		Model:                  model,
		can_create:             true,
		can_edit:               true,
		can_delete:             true,
		can_view_details:       true,
		can_export:             true, // false
		page_size:              20,
		can_set_page_size:      false,
		column_display_actions: true,
		//
		column_descriptions: map[string]string{},
	}

	mv.Blueprint = &Blueprint{
		Name:     model.label(),
		Endpoint: model.name(),
		Path:     "/" + model.name(),
		Children: map[string]*Blueprint{
			// In flask-admin use `view.index`. Should use `view.index_view` in `gadmin`
			"index":        {Endpoint: "index", Path: "/", Handler: mv.index},
			"index_view":   {Endpoint: "index_view", Path: "/", Handler: mv.index},
			"create_view":  {Endpoint: "create_view", Path: "/new", Handler: mv.newHandler},
			"details_view": {Endpoint: "details_view", Path: "/details", Handler: mv.detailHandler},
			"action_view":  {Endpoint: "action_view", Path: "/action", Handler: mv.ajaxUpdate},
			"execute_view": {Endpoint: "execute_view", Path: "/execute", Handler: mv.index},
			"edit_view":    {Endpoint: "edit_view", Path: "/edit", Handler: mv.editHandler},
			"delete_view":  {Endpoint: "delete_view", Path: "/delete", Handler: mv.deleteHandler},
			// not .export_view
			"export": {Endpoint: "export", Path: "/export", Handler: mv.index},
			"debug":  {Endpoint: "debug", Path: "/debug", Handler: mv.debug},
			// for json
			"list": {Endpoint: "list", Path: "/list", Handler: mv.listJson},
		},
	}

	mv.column_list = lo.Map(mv.columns, func(col Column, _ int) string {
		return col.Name
	})
	mv.column_sortable_list = mv.sortable_list()

	return &mv
}

func (V *ModelView) applyJoins(db *gorm.DB) *gorm.DB {
	for _, q := range V.joins {
		db = db.Joins(q.query, q.args...)
	}
	for _, q := range V.innerJoins {
		db = db.InnerJoins(q.query, q.args...)
	}
	for _, q := range V.preloads {
		db = db.Preload(q.query, q.args...)
	}
	return db
}

func (V *ModelView) Joins(query string, args ...any) *ModelView {
	V.joins = append(V.joins, queryArg{query, args})
	return V
}
func (V *ModelView) InnerJoins(query string, args ...any) *ModelView {
	V.innerJoins = append(V.innerJoins, queryArg{query, args})
	return V
}
func (V *ModelView) Preload(query string, args ...any) *ModelView {
	V.preloads = append(V.preloads, queryArg{query, args})
	return V
}

// Permissions
// Is model creation allowed
func (V *ModelView) SetCanCreate(v bool) *ModelView {
	V.can_create = v
	return V
}

// Is model editing allowed
func (V *ModelView) SetCanEdit(v bool) *ModelView {
	V.can_edit = v
	return V
}
func (V *ModelView) SetCanExport(v bool) *ModelView {
	V.can_export = v
	return V
}

// Collection of the model field names for the list view.
// If not set, will get them from the model.
func (V *ModelView) SetColumnList(vs ...string) *ModelView {
	V.column_list = vs
	return V
}

func (V *ModelView) SetColumnEditableList(vs ...string) *ModelView {
	V.column_editable_list = vs
	// build list_forms here
	V.list_forms = []base_form{}
	return V
}
func (V *ModelView) SetColumnDescriptions(m map[string]string) *ModelView {
	V.column_descriptions = m
	return V
}
func (V *ModelView) SetTablePrefixHtml(v string) *ModelView {
	V.table_prefix_html = v
	return V
}
func (V *ModelView) SetCanSetPageSize(v bool) *ModelView {
	V.can_set_page_size = v
	return V
}
func (V *ModelView) SetPageSize(v int) *ModelView {
	V.page_size = v
	return V
}

func (V *ModelView) dict(r *http.Request, others ...map[string]any) map[string]any {
	o := V.BaseView.dict(r, map[string]any{
		"table_prefix_html": V.table_prefix_html,
		"editable_columns":  true,
		"can_create":        V.can_create,
		"can_edit":          V.can_edit,
		"can_export":        V.can_export,
		"can_view_details":  V.can_view_details,
		"can_delete":        V.can_delete,
		"export_types":      []string{"csv", "xls"},
		// TODO: modal for edit/create/details
		"edit_modal":    false,
		"create_modal":  false,
		"details_modal": false,
		"is_modal":      false,
		"form_opts": map[string]any{
			"widget_args": nil,
			"form_rules":  []any{},
		},
		"filters":              []string{},
		"filter_groups":        []string{},
		"actions_confirmation": V.list_row_actions_confirmation(),
		"search_supported":     false,
		"return_url":           V.GetUrl(".index_view", nil),
	})

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func (V *ModelView) debug(w http.ResponseWriter, r *http.Request) {
	V.Render(w, r, "debug.gotmpl", nil, map[string]any{
		"menu":      V.menu.dict(),
		"blueprint": V.Blueprint.dict(),
	})
}
func (V *ModelView) index(w http.ResponseWriter, r *http.Request) {
	q := V.queryFrom(r)

	result := V.list(q)

	V.Render(w, r, "model_list.gotmpl", nil, map[string]any{
		"count":     len(result.Rows),
		"page":      q.Page,
		"num_pages": q.num_pages,
		"page_size": q.PageSize,
		"page_size_url": func(page_size int) string {
			return V.GetUrl(".index_view", q, "page_size", page_size)
		},
		"can_set_page_size":        V.can_set_page_size,
		"data":                     result.Rows,
		"request":                  rd(r),
		"get_pk_value":             V.get_pk_value,
		"column_display_pk":        V.column_display_pk,
		"column_display_actions":   V.column_display_actions,
		"column_extra_row_actions": nil,
		"list_row_actions":         V.list_row_actions(),
		"actions":                  []string{"delete", "Delete"}, // [('delete', 'Delete')]
		"list_columns":             V.list_columns(),
		"is_sortable": func(name string) bool {
			_, ok := lo.Find(V.column_sortable_list, func(s string) bool {
				return s == name
			})
			return ok
		},
		// in template, `sort url` is: ?sort={index}
		// transform `index` to `column name`
		"sort_column": func() string {
			if q.Sort != "" {
				idx := must(strconv.Atoi(q.Sort))
				if idx != -1 {
					return V.column_list[idx]
				}
			}
			return ""
		}(),
		"sort_desc": q.Desc,
		"sort_url": func(name string, invert ...bool) string {
			q := *q // simply copy
			q.Sort = strconv.Itoa(V.get_column_index(name))
			q.Desc = firstOr(invert)
			return V.GetUrl(".index_view", &q)
		},
		"column_descriptions": func(name string) string {
			if desc, ok := V.column_descriptions[name]; ok {
				return desc
			}
			return V.find(name).Description
		},
	})
}
func (V *ModelView) listJson(w http.ResponseWriter, r *http.Request) {
	q := V.queryFrom(r)

	res := V.list(q)
	if res.Error != nil {
		ReplyJson(w, 200, map[string]any{"error": res.Error})
		return
	}
	ReplyJson(w, 200, map[string]any{"total": res.Total, "data": res.Rows})
}

func (V *ModelView) queryFrom(r *http.Request) *Query {
	q := Query{default_page_size: V.page_size}
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

// TODO: store result
func (V *ModelView) list_columns() []Column {
	return lo.Filter(V.columns, func(col Column, _ int) bool {
		// in `column_list`
		_, ok := lo.Find(V.column_list, func(c string) bool {
			return c == col.DBName
		})

		// not in `column_exclude_list`
		_, exclude := lo.Find(V.column_exclude_list, func(c string) bool {
			return c == col.DBName
		})
		return ok && !exclude
	})
}

func (V *ModelView) get_column_index(name string) int {
	if _, i, ok := lo.FindIndexOf(V.column_list, func(c string) bool {
		return c == name
	}); ok {
		return i
	}
	return -1
}

// Generate inline edit form in list view
func (V *ModelView) list_form(col Column, r Row) template.HTML {
	x := XEditableWidget{model: V.Model, column: col}
	return x.html(r)
}
func (V *ModelView) is_editable(name string) bool {
	if !V.can_edit {
		return false
	}
	_, ok := lo.Find(V.column_editable_list, func(i string) bool {
		return i == name
	})
	return ok
}

func (V *ModelView) list_row_actions() []action {
	actions := []action{}
	if V.can_view_details {
		actions = append(actions, view_row_action)
	}
	if V.can_edit {
		actions = append(actions, edit_row_action)
	}
	if V.can_delete {
		actions = append(actions, delete_row_action)
	}
	return actions
}

func (V *ModelView) list_row_actions_confirmation() map[string]string {
	res := map[string]string{}
	for _, a := range V.list_row_actions() {
		if c, ok := a["confirmation"]; ok {
			res[a["name"].(string)] = c.(string)
		}
	}
	return res
}

func (V *ModelView) newHandler(w http.ResponseWriter, r *http.Request) {
	q := V.queryFrom(r)
	if !V.can_create {
		V.redirect(w, r, q.Get("url"))
		return
	}

	if r.Method == "POST" {
		// trigger ParseMultipartForm
		continue_editing := r.PostFormValue("_continue_editing")

		one := V.parseForm(r.PostForm)
		if err := V.create(one); err == nil {
			Flash(r, gettext("Record was successfully created."), "success")

			// "_add_another"
			// "_continue_editing"
			if continue_editing != "" {
				V.redirect(w, r, V.GetUrl(".edit_view", nil, "id", one["id"]))
				return
			}

			if r.PostFormValue("_add_another") != "" {
				V.redirect(w, r, V.GetUrl(".create_view", nil))
				return
			}

			V.redirect(w, r, q.Get("url"))
			return
		}

	}

	V.Render(w, r, "model_create.gotmpl", nil, map[string]any{
		"request":    rd(r),
		"form":       V.form(nil).dict(),
		"cancel_url": "TODO:cancel_url",
		"form_opts": map[string]any{
			"widget_args": nil, "form_rules": nil,
		},
		"action": nil,
	})
}

func (V *ModelView) redirect(w http.ResponseWriter, r *http.Request, url string) {
	if url == "" {
		url = V.GetUrl(".index_view", nil)
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (V *ModelView) editHandler(w http.ResponseWriter, r *http.Request) {
	q := V.queryFrom(r)

	if !V.can_edit {
		V.redirect(w, r, q.Get("url"))
		return
	}

	one, err := V.getOne(q.Get("id")) // TODO: "id" = model.pk.DBName
	if err != nil {
		// TODO: work?
		Flash(r, V.admin.gettext("Record does not exist."), "danger")

		V.redirect(w, r, q.Get("url"))
		return
	}

	V.Render(w, r, "model_edit.gotmpl", nil, map[string]any{
		"model":           one,
		"form":            V.form(one).dict(),
		"details_columns": V.list_columns(),
		"request":         rd(r),
	})
}

// Model().Where(pk field = pk value).Delete()
func (V *ModelView) deleteHandler(w http.ResponseWriter, r *http.Request) {
}

// Model().Where(pk field = pk value).First()
func (V *ModelView) detailHandler(w http.ResponseWriter, r *http.Request) {
	q := V.queryFrom(r)
	redirect := func() {
		url := q.Get("url")
		if url == "" {
			url = V.GetUrl(".index_view", nil)
		}
		http.Redirect(w, r, url, http.StatusFound)
	}

	if !V.can_view_details {
		redirect()
		return
	}

	one, err := V.getOne(q.Get("id"))
	if err != nil {
		Flash(r, V.admin.gettext("Record does not exist."), "danger")

		redirect()
		return
	}

	V.Render(w, r, "model_details.gotmpl", nil, map[string]any{
		"model":           one, // TODO: rename 'model' to 'row'
		"details_columns": V.list_columns(),
		"request":         rd(r),
	})
}

// list_form_pk=a1d13310-7c10-48d5-b63b-3485995ad6a4&currency=USD
// Record was successfully saved.
// Model().Where().Update(currency=USD)
func (V *ModelView) ajaxUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(405)
		return
	}

	if len(V.column_editable_list) == 0 {
		w.WriteHeader(404)
		return
	}

	// form
	r.ParseForm()
	pk := r.Form.Get("list_form_pk")

	// TODO: type list_form struct, parse
	row := Row{}
	for k, v := range r.Form {
		if k == "list_form_pk" {
			continue
		}
		row[k] = v[0]
	}

	// validate
	// getOne
	// record, err := mv.get(mv.DB, pk)
	// if err == gorm.ErrRecordNotFound {
	// 	w.WriteHeader(500)
	// 	w.Write([]byte(mv.gettext("Record does not exist.")))
	// 	return
	// }
	// _ = record

	// update_model
	if err := V.update(pk, row); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(V.admin.gettext("Failed to update record. %s", err)))
		return
	}
	w.Write([]byte(V.admin.gettext("Record was successfully saved.")))
}

// request to dict, like flask.request
func rd(r *http.Request) map[string]any {
	return map[string]any{
		"method": r.Method,
		"url":    r.URL.String(),
		"args":   r.URL.Query(),
	}
}

func (V *ModelView) form(one Row) ModelForm {
	form := ModelForm{
		Fields: lo.Filter(V.columns, func(col Column, _ int) bool {
			return !col.PrimaryKey
		}),
	}
	form.setValue(one)
	return form
}

func (V *ModelView) Render(w http.ResponseWriter, r *http.Request, name string, funcs template.FuncMap, data map[string]any) {
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

	fm := merge(template.FuncMap{
		"return_url": func() (string, error) {
			return V.admin.GetUrl(V.Endpoint+".index_view", nil)
		},
		"get_flashed_messages": func() []map[string]any {
			return GetFlashedMessages(r)
		},
		"get_url": func(endpoint string, args ...any) string {
			return V.GetUrl(endpoint, nil, args...)
		},
		"page_size_url": func(page_size int) string {
			return V.GetUrl(".index_view", nil, "page_size", page_size)
		},
		"pager_url": func(page int) string {
			return V.GetUrl(".index_view", nil, "page", page)
		},
		"csrf_token":  NewCSRF(CurrentSession(r)).GenerateToken,
		"list_form":   V.list_form,
		"is_editable": V.is_editable,
	}, funcs)

	fs = append(fs, "templates/"+name)
	if err := createTemplate(fs, V.admin.funcs(fm)).
		ExecuteTemplate(w, name, V.dict(r, data)); err != nil {
		panic(err)
	}
}

func (V *ModelView) applyQuery(db *gorm.DB, q *Query, count_only bool) *gorm.DB {
	ndb := db
	limit := lo.Ternary(q.PageSize != 0, q.PageSize, q.default_page_size)
	if !count_only {
		ndb = ndb.Limit(limit)

		if q.Page > 0 {
			ndb = ndb.Offset(limit * q.Page)
		}
	}

	if q.Sort != "" {
		column_index := must(strconv.Atoi(q.Sort))
		column_name := V.columns[column_index].Name

		ndb = ndb.Order(clause.OrderByColumn{
			Column: clause.Column{Name: column_name},
			Desc:   q.Desc,
		})
	}

	// filter or search
	return ndb
}

func (V *ModelView) list(q *Query) *Result {
	r := Result{Query: q}

	var total int64
	if err := V.applyQuery(V.admin.DB, q, true).
		Model(V.Model.new()).
		Count(&total).Error; err != nil {
		r.Error = err
		return &r
	}
	r.Total = total

	ptr := V.newSlice()
	db := V.applyQuery(V.admin.DB, q, false)
	if err := V.applyJoins(db).
		Find(ptr.Interface()).Error; err != nil {
		r.Error = err
		return &r
	}

	// ctx := context.Background() // TODO:

	// better way?
	len := ptr.Elem().Len()

	r.Rows = make([]Row, len)
	for i := 0; i < len; i++ {
		o := ptr.Elem().Index(i).Interface()
		r.Rows[i] = V.intoRow(o)
	}
	return &r
}

func (V *ModelView) getOne(pk any) (Row, error) {
	ptr := V.Model.new()
	// TODO: set ptr's id=pk

	db := V.applyJoins(V.admin.DB)
	if err := db.First(ptr, fmt.Sprintf("%s=?", V.pk.DBName), pk).Error; err != nil {
		return nil, err
	}
	return V.intoRow(ptr), nil
}

func (V *ModelView) update(pk any, row map[string]any) error {
	ptr := V.Model.new()

	if rc := V.admin.DB.Model(ptr).
		Where(map[string]any{V.pk.DBName: pk}).
		Updates(row); rc.Error != nil || rc.RowsAffected != 1 {
		return rc.Error
	}
	return nil
}

// row -> Model().Create() RETURNING *
func (V *ModelView) create(row map[string]any) error {
	ptr := V.Model.new()

	if rc := V.admin.DB.Model(ptr).
		Clauses(clause.Returning{}). // RETURNING *
		Create(row); rc.Error != nil || rc.RowsAffected != 1 {
		return rc.Error
	}
	return nil
}
