package gadmin

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type column map[string]any

func (c column) name() string  { return c["name"].(string) }
func (c column) label() string { return c["label"].(string) }

type row map[string]any

func (r row) get(name string) any { return r[name] }

type model struct {
	typo    reflect.Type
	schema  *schema.Schema
	columns []column
	pk      column // TODO: multiple primary keys
}

func new_model(m any) *model {
	s := must[*schema.Schema](schema.Parse(m, &schemaStore, schema.NamingStrategy{}))

	columns := lo.Map(s.Fields, func(field *schema.Field, _ int) column {
		return column{
			"id":          field.DBName, //
			"name":        field.DBName, //
			"description": field.Comment,
			"required":    field.NotNull,
			"choices":     nil,
			"type":        "StringField", // TODO:
			"label":       strings.Join(camelcase.Split(field.Name), " "),
			"widget":      field2widget(field),
			"errors":      nil,
			"primary_key": field.PrimaryKey,
		}
	})

	pk, _ := lo.Find(columns, func(c column) bool {
		return c["primary_key"].(bool)
	})

	return &model{
		typo:    reflect.TypeOf(m),
		schema:  s,
		columns: columns,
		pk:      pk,
	}
}

// TODO: more field type
func field2widget(field *schema.Field) map[string]any {
	table := map[reflect.Kind]string{
		reflect.String: "text",
		// reflect.:"password",
		// reflect.:"hidden",
		reflect.Bool: "checkbox",
		// reflect.:"radio",
		// reflect.:"file",
		// reflect.:"submit",
	}

	return map[string]any{
		"input_type": table[field.FieldType.Kind()],
	}
}

// new t
func (m *model) new() any {
	return reflect.New(m.typo).Interface()
}

// new []t
func (m *model) new_slice() reflect.Value {
	return reflect.New(reflect.SliceOf(m.typo))
}

func (m *model) into_row(a any) row {
	ctx := context.TODO()
	v := reflect.ValueOf(a)

	r := row{}
	for _, f := range m.schema.Fields {
		i, _ := f.ValueOf(ctx, v)
		r[f.DBName] = i
	}
	return r
}

// Return all field can be sorted
// exclude relationship fields
func (m *model) sortable_list() []string {
	cols := lo.Filter(m.columns, func(col column, _ int) bool {
		_, ok := m.schema.Relationships.Relations[col.label()]
		return !ok
	})
	return lo.Map(cols, func(col column, _ int) string {
		return col.name()
	})
}
func (m *model) get_pk_value(row row) any {
	return row.get(m.pk.name())
}

type query struct {
	page  int
	limit int
	sort  Order
	// search string
	// filters []
}

func (q *query) sort_desc() int {
	return q.sort.desc()
}
func (q *query) sort_column() string {
	return q.sort.name()
}
func (q *query) apply(db *gorm.DB) *gorm.DB {
	n := db.Limit(q.limit)

	if q.page > 0 {
		n = n.Offset(q.limit * q.page)
	}

	if q.sort_column() != "" {
		n = n.Order(clause.OrderByColumn{
			Column: clause.Column{Name: q.sort_column()},
			Desc:   q.sort_desc() == 1,
		})
	}
	return n
}

func Query() *query {
	return &query{
		page:  1,
		limit: 10,
		sort:  Asc(""),
	}
}

func (m *model) get_list(db *gorm.DB, q *query) (int, []row, error) {
	var total int64
	if err := q.apply(db).Model(m.new()).Count(&total).Error; err != nil {
		return 0, nil, err
	}

	ptr := m.new_slice()
	if err := q.apply(db).Find(ptr.Interface()).Error; err != nil {
		return 0, nil, err
	}

	// better way?
	len := ptr.Elem().Len()
	res := make([]row, len)
	for i := 0; i < len; i++ {
		item := ptr.Elem().Index(i).Interface()
		res[i] = m.into_row(item)
	}
	return int(total), res, nil
}

func (m *model) get(db *gorm.DB, pk any) (row, error) {
	ptr := m.new()
	if err := db.First(ptr, pk).Error; err != nil {
		return nil, err
	}
	return m.into_row(ptr), nil
}

func (m *model) update(db *gorm.DB, pk any, row row) error {
	ptr := m.new()

	if rc := db.Model(ptr).
		Where(fmt.Sprintf("%s=?", m.pk["name"]), pk).
		Updates(map[string]any(row)); rc.Error != nil || rc.RowsAffected != 1 {
		return rc.Error
	}
	return nil
}

type Order interface {
	name() string
	desc() int
}

type _order struct {
	_name string
	_desc int
}

func (bo *_order) name() string {
	return bo._name
}

func (bo *_order) desc() int {
	return bo._desc
}

func Asc(name string) Order {
	return &_order{_name: name}
}
func Desc(name string) Order {
	return &_order{_name: name, _desc: 1}
}
