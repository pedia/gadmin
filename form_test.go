package gadmin

import (
	"errors"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestWidget(t *testing.T) {
	is := assert.New(t)

	type foo struct {
		Id  int `gorm:"primaryKey"`
		Bar string
	}

	col := column{"name": "bar", "label": "Bar"}
	r := row{"Id": 1, "Bar": "a foo"}
	m := new_model(foo{})
	is.Equal(1, m.get_pk_value(r))

	x := XEditableWidget{model: m, column: col}
	is.Equal(template.HTML(
		`<a data-csrf="" data-pk="1" data-role="x-editable" data-type="select2" data-url="./ajax/update/" data-value="a foo" href="#" id="bar" name="bar">a foo</a>`),
		x.html(r))
}
