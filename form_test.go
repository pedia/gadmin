package gadmin

import (
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestField(t *testing.T) {
	is := assert.New(t)

	d1 := time.Date(2024, 10, 1, 0, 0, 0, 0, time.Local)
	// f0 := InputField("text", "created_at", d1.Format(time.DateOnly), map[string]any{
	// 	"date-format": "YYYY-MM-DD",
	// 	"role":        "datepicker",
	// })
	// is.Equal(`<input class="form-control" data-date-format="YYYY-MM-DD" data-role="datepicker" id="created_at" name="created_at" type="text" value="2024-10-01" />`,
	// 	string(f0.intoHtml()))

	// f1 := InputField("hidden", "csrf", "longtext", map[string]any{})
	// is.Equal(`<input id="csrf" name="csrf" type="hidden" value="longtext" />`, string(f1.intoHtml()))

	//
	ft := NewTextField("created_at", d1.Format(time.DateTime),
		[]lo.Entry[string, any]{
			{Key: "data-date-format", Value: "YYYY-MM-DD"},
			{Key: "data-role", Value: "datepicker"},
		}...)
	is.Equal(`<input class="form-control" data-date-format="YYYY-MM-DD" data-role="datepicker" id="created_at" name="created_at" type="text" value="2024-10-01 00:00:00" />`,
		string(ft.intoFormHtml()))

	is.Equal(`<input id="csrf" name="csrf" type="hidden" value="longtext" />`,
		string(NewHiddenField("csrf", "longtext").intoFormHtml()))

	is.NotEqual(`<a class="form-control" data-date-format="YYYY-MM-DD" data-role="datepicker" id="created_at" name="created_at">2024-10-01 00:00:00</a>`,
		string(ft.intoInlineEditHtml()))
}
