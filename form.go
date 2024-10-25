package gadmin

import (
	"bytes"
	"html/template"

	"github.com/go-playground/form/v4"
)

type XEditableWidget struct {
	model  *model
	column column
}

// <a data-csrf="" data-pk="5ad19739-80a1-4b0e-9b6d-ab7e264bd4eb"
// data-role="x-editable" data-type="text" data-url="./ajax/update/"
// data-value="EUR" href="#" id="currency" name="currency">EUR</a>
func (xw *XEditableWidget) html(r row) template.HTML {
	args := map[template.HTMLAttr]any{
		"data-value": r.get(xw.column.name()),
		"data-role":  "x-editable",
		"data-url":   "./ajax/update/",
		"data-pk":    xw.model.get_pk_value(r),
		"data-csrf":  "", // TODO:
		"data-type":  "text",
		"id":         xw.column.name(),
		"name":       xw.column.name(),
		"href":       "#",
	}

	if xw.column["choices"] != nil {
		args["data-type"] = "select2"
		args["data-source"] = xw.column["choices"]
	}

	tmpl, err := template.New("test").
		Parse(`<a{{range $k,$v :=.args}} {{$k}}="{{$v}}"{{end}}>{{.display_value}}</a>`)
	if err != nil {
		panic(err)
	}
	w := bytes.Buffer{}
	tmpl.Execute(&w, map[string]any{
		"args":          args,
		"display_value": r.get(xw.column.name()),
	})
	return template.HTML(w.String())
}

type base_form struct {
	Args      row
	Validates []string
}

type model_form struct {
	Fields      []column
	Prefix      string
	ExtraFields []string
}

func (f model_form) dict() map[string]any {
	return map[string]any{
		"action":     "", // empty
		"hidden_tag": false,
		"fields":     f.Fields,
		"cancel_url": "TODO:cancel_url",
		"is_modal":   true,
		"csrf_token": true,
	}
}

// list_form_pk = HiddenField(validators=[InputRequired()])
// xx=yy
// widget
func enc() {
	e := form.NewEncoder()
	e.Encode(map[string]any{})
}

type field struct {
	Name string
	// Value       any
	Label       string
	Description string
	Default     any
	Widget      any
	RenderArgs  map[string]any
	Filters     []any
	Validators  []any
}

var inputTemplate = template.Must(template.New("input").Parse(
	`<input{{ range $k,$v :=.args }} {{$k}}="{{$v}}"{{end}} />`))

func (f *field) intoHtml() template.HTML {
	args := map[template.HTMLAttr]any{
		"id":   f.Name,
		"name": f.Name,
	}
	w := bytes.Buffer{}
	inputTemplate.Execute(&w, map[string]any{
		"args": args,
	})
	return template.HTML(w.String())
}

type hidden_field struct {
	*field
}

func (f *hidden_field) intoHtml() template.HTML {
	return ""
}

// type list_form struct {
// 	fields: []{ list_form_pk, xx }
// }
//
// form.Render(field_name string, value any)

// type delete_form struct {
// 	fields: []{ id, url }
// }
