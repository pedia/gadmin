package gadmin

import (
	"bytes"
	"html/template"

	"github.com/samber/lo"
	"gorm.io/gorm/schema"
)

// form html for create model
// one field html for inline form

// <a data-csrf="" data-pk="5ad19739-80a1-4b0e-9b6d-ab7e264bd4eb"
// data-role="x-editable" data-type="text" data-url="./ajax/update/"
// data-value="EUR" href="#" id="currency" name="currency">EUR</a>

// data-type: text,textarea,select2,combodate,number
// data-role: elect2-ajax,x-editable,x-editable-boolean,x-editable-combodate,x-editable-select2-multiple

// <a data-csrf="" data-format="YYYY-MM-DD" data-pk="" data-role="x-editable-combodate" data-template="YYYY-MM-DD" data-type="combodate" data-url="./ajax/update/" data-value="" href="#" id="born_date" name="born_date"></a>
// <a data-csrf="" data-pk="" data-role="x-editable-boolean" data-source="[{&#34;text&#34;: &#34;No&#34;, &#34;value&#34;: &#34;&#34;}, {&#34;text&#34;: &#34;Yes&#34;, &#34;value&#34;: &#34;1&#34;}]" data-type="select2" data-url="./ajax/update/" data-value="" href="#" id="valid" name="valid"></a>
// <a data-csrf="" data-pk="" data-role="x-editable" data-source="[{&#34;text&#34;: &#34;&#34;, &#34;value&#34;: &#34;__None&#34;}, {&#34;text&#34;: &#34;Admin&#34;, &#34;value&#34;: &#34;admin&#34;}, {&#34;text&#34;: &#34;Content writer&#34;, &#34;value&#34;: &#34;content-writer&#34;}, {&#34;text&#34;: &#34;Editor&#34;, &#34;value&#34;: &#34;editor&#34;}, {&#34;text&#34;: &#34;Regular user&#34;, &#34;value&#34;: &#34;regular-user&#34;}]" data-type="select2" data-url="./ajax/update/" data-value="editor" href="#" id="type" name="type">editor</a>
// <a data-csrf="" data-pk="" data-role="x-editable" data-type="text" data-url="./ajax/update/" data-value="EUR" href="#" id="currency" name="currency">EUR</a>
// <a data-csrf="" data-pk="" data-role="x-editable" data-type="number" data-url="./ajax/update/" data-value="49" href="#" id="dialling_code" name="dialling_code">49</a>
func InlineEdit(model *Model, field *Field, row Row) template.HTML {
	args := map[template.HTMLAttr]any{
		"data-value": row.Get(field),
		"data-role":  "x-editable", // x-editable-boolean, x-editable-combodate data-template
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

type modelForm struct {
	Fields []*Field
	Row    Row
}

func ModelForm(fields []*Field, rows ...Row) *modelForm {
	form := &modelForm{Fields: fields}
	if len(rows) > 0 {
		form.Row = rows[0]
	}
	return form
}

func (f *modelForm) patchValue() []*valueField {
	return lo.Map(f.Fields, func(field *Field, _ int) *valueField {
		return &valueField{field, f.Row.Get(field)}
	})
}

func (f *modelForm) Html() template.HTML {
	w := bytes.Buffer{}
	// TODO: debug only
	formTemplate = template.Must(template.ParseFiles("templates/form.gotmpl"))

	// TODO: rename form_all to render_form
	if err := formTemplate.ExecuteTemplate(&w, "form_all", f.patchValue()); err != nil {
		panic(err)
	}
	return template.HTML(w.String())
}

type valueField struct {
	*Field
	Value any
}

func (f *valueField) Html() template.HTML {
	if len(f.Choices) > 0 {
		return f.render("field_select2")
	}

	switch f.DataType {
	case schema.Bool:
		return f.render("field_checkbox")
	case schema.Int, schema.Uint, schema.Float:
		return f.render("field_number")
	case schema.String:
		if f.TextAreaRow != 0 {
			return f.render("field_textarea")
		}
		return f.render("field_text")
	case schema.Time:
		return f.render("field_time")
	case schema.Bytes:
	}
	return ""
}

func (f *valueField) render(tmpl string) template.HTML {
	w := bytes.Buffer{}
	if err := formTemplate.ExecuteTemplate(&w, tmpl, f); err != nil {
		panic(err)
	}
	return template.HTML(w.String())
}
