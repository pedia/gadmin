package gadmin

import (
	"html/template"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

type FooBar struct {
	Id             int `gorm:"primaryKey"`
	Bar            string
	ZenOfScreaming bool
	RawDate        time.Time
	WhenHappend    null.Time
	LowestPrice    decimal.Decimal
}

var af = FooBar{Id: 3, Bar: "a foo", ZenOfScreaming: true,
	RawDate: time.Date(2014, 11, 30, 12, 59, 59, 0, time.Local),
}

func TestModel(t *testing.T) {
	is := assert.New(t)
	m := new_model(af)
	is.NotNil(m)
	is.NotNil(m.schema)
	is.Len(m.columns, 6)
	is.Equal("Bar", m.columns[1]["label"])
	is.Equal("Zen Of Screaming", m.columns[2]["label"])

	r1 := m.intoRow(af)
	is.Len(r1, 6)
	is.True(r1["zen_of_screaming"].(bool))
	is.Equal("a foo", r1["bar"])

	is.Equal(3, m.get_pk_value(r1))

	r2 := m.intoRow(&af)
	is.Len(r2, 6)
	is.True(r2["zen_of_screaming"].(bool))
	is.Equal("a foo", r2["bar"])
}

func TestWidget(t *testing.T) {
	is := assert.New(t)

	m := new_model(FooBar{})
	r := m.intoRow(af)
	is.Equal(3, m.get_pk_value(r))

	x := XEditableWidget{model: m, column: m.columns[1]}
	is.Equal(template.HTML(
		`<a data-csrf="" data-pk="3" data-role="x-editable" data-type="text" data-url="./ajax/update/" data-value="a foo" href="#" id="bar" name="bar">a foo</a>`),
		x.html(r))
}
