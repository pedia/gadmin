package gadmin

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/fatih/camelcase"
	"github.com/go-playground/form/v4"
	"github.com/gorilla/csrf"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
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
	// TODO: id, name, dept.id
	column_sortable_list []string
	column_descriptions  map[string]string

	table_prefix_html string

	// Pagination settings
	page_size         int
	can_set_page_size bool
	column_filters    []string

	column_searchable_list []string

	column_display_pk      bool
	column_display_actions bool

	// form
	form_choices          map[string][]Choice
	form_columns          []string
	form_excluded_columns []string

	// <textarea row=5>
	textareaRow map[string]int
	// TODO: date-format="YYYY-MM-DD"

	// runtime: cache generate from
	createFormFields []*Field
	editFormFields   []*Field

	list_fields []*Field

	// db relation settings
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
		BaseView:               NewView(Menu{Name: model.label(), Category: cate}),
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
		Endpoint: model.path(),
		Path:     "/" + model.path(),
		Children: map[string]*Blueprint{
			// In flask-admin use `view.index`. Should use `view.index_view` in `gadmin`
			"index":        {Endpoint: "index", Path: "/", Handler: mv.indexHandler},
			"index_view":   {Endpoint: "index_view", Path: "/", Handler: mv.indexHandler},
			"create_view":  {Endpoint: "create_view", Path: "/new", Handler: mv.newHandler},
			"details_view": {Endpoint: "details_view", Path: "/details", Handler: mv.detailHandler},
			"ajax_update":  {Endpoint: "ajax_update", Path: "/ajax/update", Handler: mv.ajaxUpdate},
			"ajax_lookup":  {Endpoint: "ajax_lookup", Path: "/ajax/lookup", Handler: mv.ajaxLookup},
			"action_view":  {Endpoint: "action_view", Path: "/action", Handler: mv.actionHandler},
			"edit_view":    {Endpoint: "edit_view", Path: "/edit", Handler: mv.editHandler},
			"delete_view":  {Endpoint: "delete_view", Path: "/delete", Handler: mv.deleteHandler},
			// not .export_view
			"export": {Endpoint: "export", Path: "/export", Handler: mv.exportHandler},
			"debug":  {Endpoint: "debug", Path: "/debug", Handler: mv.debugHandler},
			// for json
			"list": {Endpoint: "list", Path: "/list", Handler: mv.listJson},
		},
	}

	// mv.column_list = mv.schema.DBNames
	mv.column_sortable_list = mv.sortableColumns()

	return &mv
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
func (V *ModelView) SetCanSetPageSize() *ModelView {
	V.can_set_page_size = true
	return V
}
func (V *ModelView) SetPageSize(v int) *ModelView {
	V.page_size = v
	return V
}
func (V *ModelView) SetColumnSearchableList(s ...string) *ModelView {
	V.column_searchable_list = s
	return V
}
func (V *ModelView) SetTextareaRow(rows map[string]int) *ModelView {
	V.textareaRow = rows
	return V
}

// form
func (V *ModelView) SetFormChoices(choices map[string][]Choice) *ModelView {
	V.form_choices = choices
	return V
}
func (V *ModelView) SetFormColumns(columns ...string) *ModelView {
	V.form_columns = columns
	return V
}
func (V *ModelView) SetFormExcludedColumns(columns ...string) *ModelView {
	V.form_excluded_columns = columns
	return V
}

func (V *ModelView) genFormFields(create bool) []*Field {
	list := lo.Filter(V.schema.Fields, func(field *schema.Field, _ int) bool {
		// exclude, return false
		if slices.Contains(V.form_excluded_columns, field.DBName) {
			return false
		}

		// include, return true
		if slices.Contains(V.form_columns, field.DBName) {
			return true
		}

		// default not primaryKey
		if create {
			// create form, do not need primaryKey
			return !field.PrimaryKey
		} else {
			// edit form, need primaryKey
			return true
		}
	})

	return lo.Map(list, func(field *schema.Field, _ int) *Field {
		return &Field{
			Field:       field,
			Label:       strings.Join(camelcase.Split(field.Name), " "),
			Choices:     V.form_choices[field.DBName],
			Description: emptyOr(V.column_descriptions[field.DBName], field.Comment),
			TextAreaRow: V.textareaRow[field.DBName],
			Hidden:      false,
		}
	})
}

func (V *ModelView) genListFields() []*Field {
	if V.list_fields != nil {
		return V.list_fields
	}

	list := lo.Filter(V.schema.Fields, func(field *schema.Field, _ int) bool {
		// exclude return false
		if slices.Contains(V.column_exclude_list, field.DBName) {
			return false
		}

		// include, return true
		if slices.Contains(V.column_list, field.DBName) {
			return true
		}

		if len(V.column_list) > 0 {
			return false
		} else {
			if field.PrimaryKey {
				return V.column_display_pk
			}
			return true
		}
	})

	V.list_fields = lo.Map(list, func(field *schema.Field, _ int) *Field {
		return &Field{
			Field:       field,
			Label:       strings.Join(camelcase.Split(field.Name), " "),
			Choices:     V.form_choices[field.DBName],
			Description: emptyOr(V.column_descriptions[field.DBName], field.Comment),
			TextAreaRow: V.textareaRow[field.DBName],
		}
	})
	return V.list_fields
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
		"return_url":           must(V.Blueprint.GetUrl(".index_view", nil)),
	})

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func (V *ModelView) debugHandler(w http.ResponseWriter, r *http.Request) {
	V.Render(w, r, "debug.gotmpl", nil, map[string]any{
		"menu":      V.Menu.dict(),
		"blueprint": V.Blueprint.dict(),
	})
}
func (V *ModelView) indexHandler(w http.ResponseWriter, r *http.Request) {
	q := V.queryFrom(r)

	result := V.list(q)

	V.Render(w, r, "model_list.gotmpl", template.FuncMap{
		"is_sortable": func(name string) bool {
			return slices.Contains(V.column_sortable_list, name)
		},
		"sort_url": func(name string, invert ...bool) string {
			q := *q // simply copy
			q.Sort = cast.ToString(V.get_column_index(name))
			q.Desc = firstOr(invert)
			return must(V.Blueprint.GetUrl(".index_view", queryToPairs(q.toValues())...))
		},
		"column_descriptions": func(name string) string {
			if desc, ok := V.column_descriptions[name]; ok {
				return desc
			}
			return V.find(name).Description
		},
	}, map[string]any{
		"count":             len(result.Rows),
		"page":              q.Page,
		"num_pages":         result.NumPages(),
		"page_size":         q.PageSize,
		"default_page_size": q.default_page_size,
		"page_size_url": func(page_size int) string {
			uv := q.toValues()
			uv.Set("page_size", cast.ToString(page_size))
			return must(V.Blueprint.GetUrl(".index_view", queryToPairs(uv)...))
		},
		"can_set_page_size":        V.can_set_page_size,
		"data":                     result.Rows, // TODO: remove
		"result":                   result,
		"request":                  rd(r),
		"get_pk_value":             V.get_pk_value,
		"column_display_pk":        V.column_display_pk,
		"column_display_actions":   V.column_display_actions,
		"column_extra_row_actions": nil,
		"list_row_actions":         V.list_row_actions(),
		"actions":                  []string{"delete", "Delete"}, // [('delete', 'Delete')]
		"list_columns":             V.genListFields(),
		"sort":                     q.Sort,
		// not func, return current sort field name
		// in template, `sort url` is: ?sort={index}
		// transform `index` to `column name`
		"sort_column": func() string {
			if q.Sort != "" {
				idx := cast.ToInt(q.Sort)
				return V.column_name(idx)
			}
			return ""
		}(),
		"sort_desc":              q.Desc,
		"search":                 q.Search,
		"column_searchable_list": V.column_searchable_list,
		"search_placeholder":     strings.Join(V.column_searchable_list, ","),
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

// Because `default_page_size`, should place here, not query.go
func (V *ModelView) queryFrom(r *http.Request) *Query {
	base, _ := V.GetBlueprint().GetUrl(".index")
	q := Query{default_page_size: V.page_size, PageSize: V.page_size,
		base: base}
	r.ParseForm()
	uv := r.Form

	form.NewDecoder().Decode(&q, uv)
	for k, v := range uv {
		if lo.IndexOf([]string{"page", "page_size", "sort", "desc", "search"}, k) != -1 {
			continue
		}
		q.args = append(q.args, k, v[0])
	}
	return &q
}

func (V *ModelView) get_column_index(name string) int {
	if _, i, ok := lo.FindIndexOf(V.genListFields(), func(f *Field) bool {
		return f.DBName == name
	}); ok {
		return i
	}
	return -1
}

func (V *ModelView) column_name(i int) string {
	list := V.genListFields()
	if i < len(list) {
		return list[i].DBName
	}
	return ""
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

func (V *ModelView) list_row_actions() []Action {
	actions := []Action{}
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

	if r.Method == http.MethodPost {
		// trigger ParseMultipartForm
		continue_editing := r.PostFormValue("_continue_editing")

		one := V.parseForm(r.PostForm)
		if err := V.create(one); err == nil {
			// Flash(r, gettext("Record was successfully created."), "success")

			// "_add_another"
			// "_continue_editing"
			if continue_editing != "" {
				V.redirect(w, r, must(V.Blueprint.GetUrl(".edit_view", nil, "id", one["id"])))
				return
			}

			if r.PostFormValue("_add_another") != "" {
				V.redirect(w, r, must(V.Blueprint.GetUrl(".create_view", nil)))
				return
			}

			V.redirect(w, r, q.Get("url"))
			return
		}

	}

	V.Render(w, r, "model_create.gotmpl", nil, map[string]any{
		"request":    rd(r),
		"form":       V.form(nil, csrf.Token(r)),
		"cancel_url": "TODO:cancel_url",
		"form_opts": map[string]any{
			"widget_args": nil, "form_rules": nil,
		},
		"action": nil,
	})
}

func (V *ModelView) redirect(w http.ResponseWriter, r *http.Request, url string) {
	if url == "" {
		url = must(V.Blueprint.GetUrl(".index_view", nil))
	}
	http.Redirect(w, r, url, http.StatusFound)
}

func (V *ModelView) editHandler(w http.ResponseWriter, r *http.Request) {
	q := V.queryFrom(r)

	if !V.can_edit {
		V.redirect(w, r, q.Get("url"))
		return
	}

	one, err := V.getOne(q.Get("id")) // force "id"
	if err != nil {
		// TODO: make this work
		// Flash(r, V.admin.gettext("Record does not exist."), "danger")

		V.redirect(w, r, q.Get("url"))
		return
	}

	V.Render(w, r, "model_edit.gotmpl", nil, map[string]any{
		"model":           one,
		"form":            V.form(one, csrf.Token(r)),
		"details_columns": V.genListFields(),
		"request":         rd(r),
	})
}

// Model().Where(pk field = pk value).Delete()
func (V *ModelView) deleteHandler(w http.ResponseWriter, r *http.Request) {
	q := V.queryFrom(r)
	redirect := func() {
		url := q.Get("url")
		if url == "" {
			url = must(V.Blueprint.GetUrl(".index_view", nil))
		}
		http.Redirect(w, r, url, http.StatusFound)
	}

	if !V.can_delete {
		redirect()
		return
	}

	err := V.deleteOne(q.Get("id"))
	if err != nil {
		// flash_errors(form, message='Failed to delete record. %(error)s')
	} else {
		// flash(ngettext('Record was successfully deleted.',
		//                      '%(count)s records were successfully deleted.',
		//                      count, count=count), 'success')
	}
	redirect()
}

// Model().Where(pk field = pk value).First()
func (V *ModelView) detailHandler(w http.ResponseWriter, r *http.Request) {
	q := V.queryFrom(r)
	redirect := func() {
		url := q.Get("url")
		if url == "" {
			url = must(V.Blueprint.GetUrl(".index_view", nil))
		}
		http.Redirect(w, r, url, http.StatusFound)
	}

	if !V.can_view_details {
		redirect()
		return
	}

	one, err := V.getOne(q.Get("id"))
	if err != nil {
		// Flash(r, V.admin.gettext("Record does not exist."), "danger")

		redirect()
		return
	}

	V.Render(w, r, "model_details.gotmpl", nil, map[string]any{
		"model":           one, // TODO: rename 'model' to 'row'
		"details_columns": V.genListFields(),
		"request":         rd(r),
	})
}

// list_form_pk=a1d13310-7c10-48d5-b63b-3485995ad6a4&currency=USD
// Record was successfully saved.
// Model().Where().Update(currency=USD)
func (V *ModelView) ajaxUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(405)
		return
	}

	if len(V.column_editable_list) == 0 {
		w.WriteHeader(404)
		return
	}

	// form
	r.ParseForm()
	rowid := r.Form.Get("list_form_pk")

	row := Row{}
	for k, v := range r.Form {
		if k == "list_form_pk" {
			continue
		}
		row[k] = v[0]
	}

	// TODO: validate

	// update_model
	if err := V.update(rowid, row); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(V.admin.gettext("Failed to update record. %s", err)))
		return
	}
	w.Write([]byte(V.admin.gettext("Record was successfully saved.")))
}

func (V *ModelView) ajaxLookup(w http.ResponseWriter, r *http.Request)    {}
func (V *ModelView) actionHandler(w http.ResponseWriter, r *http.Request) {}
func (V *ModelView) exportHandler(w http.ResponseWriter, r *http.Request) {
	q := V.queryFrom(r)
	result := V.list(q)
	if result.Error != nil {
		panic(result.Error)
	}

	fn := fmt.Sprintf("attachment;filename=%s-%s.csv", V.name(),
		time.Now().Format(time.DateOnly))
	w.Header().Add("content-disposition", fn)
	w.Header().Add("content-type", "text/csv")

	cw := csv.NewWriter(w)

	// header
	header := lo.Map(V.genListFields(), func(f *Field, _ int) string {
		return f.DBName
	})
	cw.Write(header)

	for _, row := range result.Rows {
		line := lo.Map(V.form(row, csrf.Token(r)).Fields, func(f *Field, _ int) string {
			return cast.ToString(row.GetDisplayValue(f))
		})
		cw.Write(line)
	}
	cw.Flush()
}

// request to dict, like flask.request
func rd(r *http.Request) map[string]any {
	return map[string]any{
		"method": r.Method,
		"url":    r.URL.String(),
		"args":   r.URL.Query(),
	}
}

// create form: no primarykey
// edit form: hidden primarykey
func (V *ModelView) form(one Row, token string) *modelForm {
	list := []*Field{HiddenField(token)}
	if one == nil {
		// create
		if V.createFormFields == nil {
			V.createFormFields = V.genFormFields(true)
		}
		list = append(list, V.createFormFields...)
	} else {
		if V.editFormFields == nil {
			V.editFormFields = V.genFormFields(false)
		}
		list = append(list, V.editFormFields...)
	}

	return ModelForm(list, one)
}

// Generate inline edit form in list view
func (V *ModelView) list_form(field *Field, row Row) template.HTML {
	return InlineEdit(V.Model, field, row)
}
func (V *ModelView) delete_form() *modelForm {
	return &modelForm{Fields: []*Field{}}
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
			return V.Blueprint.GetUrl(V.Blueprint.Endpoint+".index_view", nil)
		},
		"get_flashed_messages": func() []map[string]any {
			// return GetFlashedMessages(r)
			return nil
		},
		"get_url": func(endpoint string, args ...any) string {
			return must(V.Blueprint.GetUrl(endpoint, args...))
		},
		"pager_url": func(page int) string {
			return must(V.Blueprint.GetUrl(".index_view", nil, "page", page))
		},
		"csrf_token":  func() string { return csrf.Token(r) },
		"list_form":   V.list_form,
		"delete_form": V.delete_form,
		"is_editable": V.is_editable,
		"clear_search_url": func() string {
			q := V.queryFrom(r)
			q.Search = ""
			return must(V.Blueprint.GetUrl(".index_view", queryToPairs(q.toValues())...))
		},
	}, funcs)

	fs = append(fs, "templates/"+name)
	if err := parseTemplate("view", V.admin.funcs(fm), fs...).
		ExecuteTemplate(w, name, V.dict(r, data)); err != nil {
		panic(err)
	}
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
		idx := cast.ToInt(q.Sort)
		column_name := V.column_name(idx)

		ndb = ndb.Order(clause.OrderByColumn{
			Column: clause.Column{Name: column_name},
			Desc:   q.Desc,
		})
	}

	// filter or search
	if q.Search != "" && len(V.column_searchable_list) > 0 {
		for i, c := range V.column_searchable_list {
			if i == 0 {
				ndb = ndb.Where(
					fmt.Sprintf("%s like ?", c),
					"%"+q.Search+"%")
			} else {
				ndb = ndb.Or(
					fmt.Sprintf("%s like ?", c),
					"%"+q.Search+"%")
			}
		}
	}
	return ndb
}

func (V *ModelView) list(q *Query) *Result {
	res := Result{Query: q}

	var total int64
	if err := V.applyQuery(V.admin.DB, q, true).
		Model(V.Model.new()).
		Count(&total).Error; err != nil {
		res.Error = err
		return &res
	}
	res.Total = total

	ptr := V.newSlice()
	db := V.applyQuery(V.admin.DB, q, false)
	if err := V.applyJoins(db).
		Find(ptr.Interface()).Error; err != nil {
		res.Error = err
		return &res
	}

	len := ptr.Elem().Len()

	res.Rows = make([]Row, len)
	for i := 0; i < len; i++ {
		o := ptr.Elem().Index(i).Interface()
		res.Rows[i] = V.intoRow(o)
	}
	return &res
}

func (V *ModelView) getOne(rowid string) (Row, error) {
	ptr := V.Model.new()

	db := V.applyJoins(V.admin.DB)
	if err := db.Where(V.where(rowid)).First(ptr).Error; err != nil {
		return nil, err
	}
	return V.intoRow(ptr), nil
}

func (V *ModelView) update(rowid string, row map[string]any) error {
	ptr := V.Model.new()

	if rc := V.admin.DB.Model(ptr).
		Where(V.where(rowid)).
		Updates(row); rc.Error != nil || rc.RowsAffected != 1 {
		return rc.Error
	}
	return nil
}

func (V *ModelView) deleteOne(rowid string) error {
	ptr := V.Model.new()

	rc := V.admin.DB.Delete(ptr, V.where(rowid))
	return rc.Error
}

// row -> Model().Create() RETURNING *
// TOOD: select2 value with __None
func (V *ModelView) create(row map[string]any) error {
	ptr := V.Model.new()

	if rc := V.admin.DB.Model(ptr).
		Clauses(clause.Returning{}). // RETURNING *
		Create(row); rc.Error != nil || rc.RowsAffected != 1 {
		return rc.Error
	}
	return nil
}
