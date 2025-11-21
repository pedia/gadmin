package gadm

import (
	"encoding/json"
	"html/template"

	"github.com/spf13/cast"
	"gorm.io/gorm/schema"
)

// form html for create model
// one field html for inline form

// <a data-csrf="" data-pk="5ad19739-80a1-4b0e-9b6d-ab7e264bd4eb"
// data-role="x-editable" data-type="text" data-url="./ajax/update/"
// data-value="EUR" href="#" id="currency" name="currency">EUR</a>

// data-type: text,textarea,select2,combodate,number
// data-role: select2-ajax,x-editable,x-editable-boolean,x-editable-combodate,x-editable-select2-multiple
// TODO: select2-ajax
// <a data-csrf="" data-pk="2" data-role="x-editable" data-rows="5" data-type="textarea" data-url="./ajax/update/" data-value="" href="#" id="text" name="text"></a>
// <a data-csrf="" data-format="YYYY-MM-DD" data-pk="" data-role="x-editable-combodate" data-template="YYYY-MM-DD" data-type="combodate" data-url="./ajax/update/" data-value="" href="#" id="born_date" name="born_date"></a>
// <a data-csrf="" data-pk="" data-role="x-editable-boolean" data-source="[{&#34;text&#34;: &#34;No&#34;, &#34;value&#34;: &#34;&#34;}, {&#34;text&#34;: &#34;Yes&#34;, &#34;value&#34;: &#34;1&#34;}]" data-type="select2" data-url="./ajax/update/" data-value="" href="#" id="valid" name="valid"></a>
// <a data-csrf="" data-pk="" data-role="x-editable" data-source="[{&#34;text&#34;: &#34;&#34;, &#34;value&#34;: &#34;__None&#34;}, {&#34;text&#34;: &#34;Admin&#34;, &#34;value&#34;: &#34;admin&#34;}, {&#34;text&#34;: &#34;Content writer&#34;, &#34;value&#34;: &#34;content-writer&#34;}, {&#34;text&#34;: &#34;Editor&#34;, &#34;value&#34;: &#34;editor&#34;}, {&#34;text&#34;: &#34;Regular user&#34;, &#34;value&#34;: &#34;regular-user&#34;}]" data-type="select2" data-url="./ajax/update/" data-value="editor" href="#" id="type" name="type">editor</a>
// <a data-csrf="" data-pk="" data-role="x-editable" data-type="text" data-url="./ajax/update/" data-value="EUR" href="#" id="currency" name="currency">EUR</a>
// <a data-csrf="" data-pk="" data-role="x-editable" data-type="number" data-url="./ajax/update/" data-value="49" href="#" id="dialling_code" name="dialling_code">49</a>
func InlineEdit(gt *groupTempl, token string, model *Model, field *Field, row *Row) template.HTML {
	dv := field.Display()
	args := map[template.HTMLAttr]any{
		"data-value": dv,
		"data-role":  "x-editable", // x-editable-boolean, x-editable-combodate data-template
		"data-url":   "ajax/update",
		"data-pk":    row.GetPkValue(),
		"data-csrf":  token,
		"data-type":  "text",
		"id":         field.DBName,
		"name":       field.DBName,
		"href":       "#",
	}

	if field.Choices != nil {
		args["data-type"] = "select2"
		args["data-source"] = jsonify(field.Choices)
	}
	if field.TextAreaRow > 0 {
		args["data-type"] = "textarea"
		args["data-row"] = cast.ToString(field.TextAreaRow)
	}

	switch field.DataType {
	case schema.Time:
		args["data-type"] = "combodate"
		args["data-template"] = "YYYY-MM-DD" // TODO
		args["data-role"] = "x-editable-combodate"
	case schema.Int, schema.Uint, schema.Float:
		args["data-type"] = "number"
	case schema.Bool:
		args["data-type"] = "select2"
		args["data-role"] = "x-editable-boolean"
		args["data-source"] = `[{"text": "False", "value": "false"},{"text": "True", "value": "true"}]`
	}

	return gt.Execute("inline_field", map[string]any{
		"args":          args,
		"display_value": dv,
		"field":         field,
	})
}

func jsonify(a any) string {
	if bs, err := json.Marshal(a); err == nil {
		return string(bs)
	}
	return ""
}

type modelForm struct {
	Fields    []*Field
	Row       *Row
	CSRFToken string
}

func NewForm(fs []*Field, row *Row, csrfToken string) *modelForm {
	return &modelForm{Fields: fs, Row: row, CSRFToken: csrfToken}
}
