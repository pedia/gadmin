package gadmin

import (
	"bytes"
	"html/template"

	"github.com/samber/lo"
)

// form html for create model
// one field html for inline form

// <a data-csrf="" data-pk="5ad19739-80a1-4b0e-9b6d-ab7e264bd4eb"
// data-role="x-editable" data-type="text" data-url="./ajax/update/"
// data-value="EUR" href="#" id="currency" name="currency">EUR</a>

// <a data-csrf="" data-pk="a5c6dc6d-b999-4db9-b409-ac8973624a3f"
// data-role="x-editable"
// data-source="[{&#34;text&#34;: &#34;&#34;, &#34;value&#34;: &#34;__None&#34;}, {&#34;text&#34;: &#34;Admin&#34;, &#34;value&#34;: &#34;admin&#34;}, {&#34;text&#34;: &#34;Content writer&#34;, &#34;value&#34;: &#34;content-writer&#34;}, {&#34;text&#34;: &#34;Editor&#34;, &#34;value&#34;: &#34;editor&#34;}, {&#34;text&#34;: &#34;Regular user&#34;, &#34;value&#34;: &#34;regular-user&#34;}]"
// data-type="select2" data-url="./ajax/update/"
// data-value="regular-user" href="#" id="type" name="type">regular-user</a>
func InlineEdit(model *Model, field *Field, row Row) template.HTML {
	args := map[template.HTMLAttr]any{
		"data-value": row.Get(field),
		"data-role":  "x-editable",
		"data-url":   "./ajax/update/",
		"data-pk":    model.get_pk_value(row),
		"data-csrf":  "", // TODO:
		"data-type":  "text",
		"id":         field.DBName,
		"name":       field.DBName,
		"href":       "#",
	}

	if field.Choices != nil {
		args["data-type"] = "select2"
		args["data-source"] = field.Choices // TODO: choices to dict
	}

	w := bytes.Buffer{}
	inlineEditTemplate.Execute(&w, map[string]any{
		"args":          args,
		"display_value": row.Get(field),
	})
	return template.HTML(w.String()) // TODO: HTML safe
}

// input type = hidden/text/checkbox/file
func FormEdit(model *Model, row Row) {}

var formTemplate *template.Template
var inputTemplate *template.Template
var inlineEditTemplate *template.Template

func init() {
	formTemplate = template.Must(template.ParseFiles("templates/form.gotmpl"))

	inputTemplate = template.Must(template.New("input").Parse(
		`<input
	{{- range $k,$v :=. -}}
		{{- if eq $k "required" }}required {{ else }} {{$k}}="{{$v}}"{{ end -}}
	{{- end }} />`))

	inlineEditTemplate = template.Must(template.New("input").Parse(
		`<a{{range $k,$v :=.args}} {{$k}}="{{$v}}"{{end}}>{{.display_value}}</a>`))

}

// list_forms[key]
//

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

type modelForm struct {
	Model *Model
	Row   Row
}

func ModelForm(model *Model, rows ...Row) *modelForm {
	mf := &modelForm{Model: model}
	if len(rows) > 0 {
		mf.Row = rows[0]
	}
	return mf
}

func (f *modelForm) dict() map[string]any {
	return map[string]any{}
}

func (f *modelForm) Html() template.HTML {
	w := bytes.Buffer{}
	formTemplate = template.Must(template.ParseFiles("templates/form.gotmpl"))
	if err := formTemplate.Execute(&w, f); err != nil {
		panic(err)
	}
	return template.HTML(w.String())
}

// type Widget struct
