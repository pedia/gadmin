package gadmin

import (
	"bytes"
	"html/template"
)

type XEditableWidget struct {
	model  *model
	column column
}

// <a data-csrf="" data-pk="a1d13310-7c10-48d5-b63b-3485995ad6a4" data-role="x-editable"
// data-source="[{"text": "", "value": "__None"}, {"text": "Admin", "value": "admin"}, {"text": "Content writer", "value": "content-writer"}, {"text": "Editor", "value": "editor"}, {"text": "Regular user", "value": "regular-user"}]"
// data-type="select2" data-url="./ajax/update/" data-value="editor" href="#"
// id="type" name="type">editor</a>
func (xw *XEditableWidget) html(r row) template.HTML {
	args := map[template.HTMLAttr]any{
		"data-value": r.get(xw.column.label()), // TODO: column.name()
		"data-role":  "x-editable",
		"data-type":  "select2",
		"data-url":   "./ajax/update/",
		"data-pk":    xw.model.get_pk_value(r),
		"data-csrf":  "",

		"id":   xw.column.name(),
		"name": xw.column.name(),
		"href": "#",
	}
	tmpl, err := template.New("test").
		Parse(`<a{{range $k,$v :=.args}} {{$k}}="{{$v}}"{{end}}>{{.display_value}}</a>`)
	if err != nil {
		panic(err)
	}
	w := &bytes.Buffer{}
	tmpl.Execute(w, map[string]any{
		"args":          args,
		"display_value": r.get(xw.column.label()),
	})
	return template.HTML(w.String())
}

type form struct {
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
