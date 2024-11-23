package gadmin

import (
	"bytes"
	"html/template"
	"log"
	"slices"

	"github.com/samber/lo"
	"gorm.io/gorm/schema"
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
	return template.HTML(w.String()) // TODO: HTML safe
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

func (F *model_form) setValue(one row) {
	for _, f := range F.Fields {
		f["value"] = one[f.name()]
	}
}

func (f model_form) dict() map[string]any {
	return map[string]any{
		"action":     "", // empty
		"hidden_tag": false,
		"fields":     f.Fields,
		"csrf_token": true,
	}
}

type input_field struct {
	Id    string
	Name  string
	Type  string
	Value any
	Data  map[string]any
}

func InputField(typo, name string, value any, data map[string]any) *input_field {
	data_names := []string{"url", "json", "role", "multiple",
		"separator", "allow-blank", "date-format", "placeholder",
		"minimum-input-length"}
	for name := range data {
		if !slices.Contains(data_names, name) {
			log.Printf("warning data name %s", name)
		}
	}

	return &input_field{
		Id:    name,
		Name:  name,
		Type:  typo,
		Value: value,
		Data:  data,
	}
}

var inputTemplate = template.Must(template.New("input").Parse(
	`<input{{ range $k,$v :=. -}}
{{- if eq $k "required" }} required{{ else }} {{$k}}="{{$v}}{{- end -}}"
{{- end }} />`))

var inlineEditTemplate = template.Must(template.New("input").Parse(
	`<a{{range $k,$v :=.}} {{$k}}="{{$v}}"{{end}}>{{.value}}</a>`))

func (F *input_field) intoHtml() template.HTML {
	args := map[template.HTMLAttr]any{
		"id":    F.Name,
		"name":  F.Name,
		"type":  F.Type,
		"value": F.Value,
	}
	//
	if F.Type == "text" {
		args["class"] = "form-control"
	}

	for k, v := range F.Data {
		args[template.HTMLAttr("data-"+k)] = v
	}

	w := bytes.Buffer{}
	inputTemplate.Execute(&w, args)
	return template.HTML(w.String())
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

type Choice struct {
	Text  string
	Value any
}

type Column struct {
	Name        string
	Description string
	Required    bool
	Choices     []Choice
	Type        string
	DataType    schema.DataType
	Label       string
	Widget      *input_field
	Errors      string
	PrimaryKey  bool
}

func (C *Column) dict() map[string]any {
	return map[string]any{
		"name":        C.Name,
		"description": C.Description,
		"required":    C.Required,
		"choices":     C.Choices,
		"type":        C.Type,
		"data_type":   C.DataType,
		"label":       C.Label,
		"widget":      C.Widget,
		"errors":      C.Errors,
		"primary_key": C.PrimaryKey,
	}
}
