package gadmin

import (
	"html/template"
	"net/url"
	"testing"
	"time"

	"github.com/go-playground/form/v4"
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

	r1 := m.into_row(af)
	is.Len(r1, 6)
	is.True(r1["zen_of_screaming"].(bool))
	is.Equal("a foo", r1["bar"])

	is.Equal(3, m.get_pk_value(r1))

	r2 := m.into_row(&af)
	is.Len(r2, 6)
	is.True(r2["zen_of_screaming"].(bool))
	is.Equal("a foo", r2["bar"])
}

func TestWidget(t *testing.T) {
	is := assert.New(t)

	m := new_model(FooBar{})
	r := m.into_row(af)
	is.Equal(3, m.get_pk_value(r))

	x := XEditableWidget{model: m, column: m.columns[1]}
	is.Equal(template.HTML(
		`<a data-csrf="" data-pk="3" data-role="x-editable" data-type="text" data-url="./ajax/update/" data-value="a foo" href="#" id="bar" name="bar">a foo</a>`),
		x.html(r))
}

func TestPlaygroundForm(t *testing.T) {
	is := assert.New(t)

	e := form.NewEncoder()

	type list_form struct {
		list_form_pk any
		CamelCase    string
	}

	is.Equal("%5Ba%5D=1", must[url.Values](e.Encode(map[string]any{
		"a": 1,
	})).Encode())

	is.Equal("CamelCase=abc", must[url.Values](e.Encode(list_form{
		list_form_pk: "33",
		CamelCase:    "abc",
	})).Encode())
}

func TestUrl(t *testing.T) {
	is := assert.New(t)
	admin := NewAdmin("Example: Simple Views", nil)

	u0, err := admin.urlFor("foo", ".index_view", map[string]any{})
	is.Nil(err)
	is.Equal("/admin/foo/", u0)

	u1, err := admin.urlFor("foo", ".index_view", map[string]any{"page_size": 10})
	is.Nil(err)
	is.Equal("/admin/foo/?page_size=10", u1)

	u2, err := admin.urlFor("", ".create_view", map[string]any{"hello": "world"})
	is.Nil(err)
	is.Equal("new/?hello=world", u2)
}
