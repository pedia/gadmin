package gadmin

import (
	"errors"
	"html/template"
	"net/url"
	"testing"
	"time"

	"github.com/go-playground/form/v4"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestBaseMust(t *testing.T) {
	is := assert.New(t)

	fr := func() (int, error) {
		return 42, nil
	}
	is.Equal(42, must[int](fr()))

	fe := func() (int, error) {
		return 1, errors.New("sth. wrong")
	}
	is.Panics(func() { must[int](fe()) })

	ft := func() (int, bool) {
		return 42, true
	}
	is.Equal(42, must[int](ft()))

	ff := func() (int, bool) {
		return 42, false
	}
	is.Panics(func() { must[int](ff()) })
}

func TestBaseConvert(t *testing.T) {
	is := assert.New(t)

	myMap := map[string]any{
		"name":    "John Doe",
		"age":     30,
		"active":  true,
		"numbers": []int{1, 2, 3}, // This will be skipped as it's not a string
	}

	urlValues := map_into_values(myMap)
	is.Equal("active=true&age=30&name=John+Doe", urlValues.Encode())
}

type foo struct {
	Id             int `gorm:"primaryKey"`
	Bar            string
	ZenOfScreaming bool
	RawDate        time.Time
	WhenHappend    null.Time
	LowestPrice    decimal.Decimal
}

var af = foo{Id: 3, Bar: "a foo", ZenOfScreaming: true,
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

	m := new_model(foo{})
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
