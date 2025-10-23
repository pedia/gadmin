package gadmin

import (
	"bytes"
	"html/template"

	"github.com/samber/lo"
)

type XEditableWidget struct {
	model  *Model
	column Column
}

// <a data-csrf="" data-pk="5ad19739-80a1-4b0e-9b6d-ab7e264bd4eb"
// data-role="x-editable" data-type="text" data-url="./ajax/update/"
// data-value="EUR" href="#" id="currency" name="currency">EUR</a>
func (xw *XEditableWidget) html(r Row) template.HTML {
	args := map[template.HTMLAttr]any{
		"data-value": r.Get(xw.column),
		"data-role":  "x-editable",
		"data-url":   "./ajax/update/",
		"data-pk":    xw.model.get_pk_value(r),
		"data-csrf":  "", // TODO:
		"data-type":  "text",
		"id":         xw.column.DBName,
		"name":       xw.column.DBName,
		"href":       "#",
	}

	if xw.column.Choices != nil {
		args["data-type"] = "select2"
		args["data-source"] = xw.column.Choices // TODO: choices to dict
	}

	tmpl, err := template.New("test").
		Parse(`<a{{range $k,$v :=.args}} {{$k}}="{{$v}}"{{end}}>{{.display_value}}</a>`)
	if err != nil {
		panic(err)
	}
	w := bytes.Buffer{}
	tmpl.Execute(&w, map[string]any{
		"args":          args,
		"display_value": r.Get(xw.column),
	})
	return template.HTML(w.String()) // TODO: HTML safe
}

type base_form struct {
	Args      Row
	Validates []string
}

type ModelForm struct {
	Fields      []Column
	Prefix      string
	ExtraFields []string
}

func (F *ModelForm) setValue(one Row) {
	for _, f := range F.Fields {
		f.Value = one[f.Name]
	}
}

func (f ModelForm) dict() map[string]any {
	return map[string]any{
		"action":     "", // empty
		"hidden_tag": false,
		"fields":     f.Fields,
		"csrf_token": true,
	}
}

var inputTemplate *template.Template
var inlineEditTemplate *template.Template

func init() {
	inputTemplate = template.Must(template.New("input").Parse(
		`<input
	{{- range $k,$v :=. -}}
		{{- if eq $k "required" }}required {{ else }} {{$k}}="{{$v}}"{{ end -}}
	{{- end }} />`))

	inlineEditTemplate = template.Must(template.New("input").Parse(
		`<a{{range $k,$v :=.}} {{$k}}="{{$v}}"{{end}}>{{.value}}</a>`))
}

type field struct {
	Entries []lo.Entry[string, any]
}

func NewField(es []lo.Entry[string, any]) *field {
	return &field{Entries: es}
}

func (F *field) render(t *template.Template) template.HTML {
	args := map[template.HTMLAttr]any{}
	for _, e := range F.Entries {
		args[template.HTMLAttr(e.Key)] = e.Value
	}

	w := bytes.Buffer{}
	t.Execute(&w, args)
	return template.HTML(w.String())
}

func (F *field) intoFormHtml() template.HTML {
	return F.render(inputTemplate)
}
func (F *field) intoInlineEditHtml() template.HTML {
	return F.render(inlineEditTemplate)
}

func NewTextField(id, value string, es ...lo.Entry[string, any]) *field {
	ps := []lo.Entry[string, any]{
		{Key: "class", Value: "form-control"},
		{Key: "id", Value: id},
		{Key: "name", Value: id},
		{Key: "type", Value: "text"},
		{Key: "value", Value: value}}
	ps = append(ps, es...)
	return &field{Entries: ps}
}

func NewHiddenField(id, value string, es ...lo.Entry[string, any]) *field {
	ps := []lo.Entry[string, any]{
		{Key: "id", Value: id},
		{Key: "name", Value: id},
		{Key: "type", Value: "hidden"},
		{Key: "value", Value: value}}
	ps = append(ps, es...)
	return &field{Entries: ps}
}

// data-csrf
// data-pk
// data-role
// data-source
// data-type
// data-url
// data-value
// href
// id
// name
