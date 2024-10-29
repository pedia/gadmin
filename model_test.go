package gadmin

import (
	"context"
	"database/sql"
	"html/template"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// Model with all-typed fields
type Foo struct {
	ID           uint `gorm:"primaryKey"`
	Name         string
	Email        *string
	Age          uint8
	Normal       bool
	Valid        *bool `gorm:"default:true"`
	MemberNumber sql.NullString

	Birthday    *time.Time
	ActivatedAt sql.NullTime
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime:nano"`

	Decimal decimal.Decimal
	// TODO: enum
}

func foos() []Foo {
	e1 := "foo@foo.com"
	d1 := time.Date(2024, 10, 1, 0, 0, 0, 0, time.Local)
	e2 := "bar@foo.com"
	d2 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	return []Foo{
		{ID: 3, Name: "foo", Email: &e1, Age: 42, Normal: true, Birthday: &d1,
			MemberNumber: sql.NullString{String: "9527", Valid: true}},
		{ID: 4, Name: "bar", Email: &e2, Age: 21, Normal: false, Birthday: &d2,
			MemberNumber: sql.NullString{String: "3699", Valid: true}},
	}
}

func TestModel(t *testing.T) {
	is := assert.New(t)

	m := newModel(Foo{})

	is.Equal("ID", m.columns[0]["label"])
	is.Equal("Email", m.columns[2]["label"])
	is.Equal("Member Number", m.columns[6]["label"])

	r1 := m.intoRow(context.TODO(), foos()[0])
	is.Equal("foo", r1["name"])
	is.True(r1["normal"].(bool))

	is.Equal(uint(3), m.get_pk_value(r1))

	// m.get_list()
}

func TestWidget(t *testing.T) {
	is := assert.New(t)

	m := newModel(Foo{})

	af := foos()[0]
	r := m.intoRow(context.TODO(), af)
	is.Equal(uint(3), m.get_pk_value(r))

	x := XEditableWidget{model: m, column: m.columns[1]}
	is.Equal(template.HTML(
		`<a data-csrf="" data-pk="3" data-role="x-editable" data-type="text" data-url="./ajax/update/" data-value="foo" href="#" id="name" name="name">foo</a>`),
		x.html(r))
}
