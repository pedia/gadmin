package gadmin

import (
	"encoding/json"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
	"gorm.io/gorm"
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
	pk      string // TODO: multiple primary keys
}

func new_model(m any) *model {
	s := must[*schema.Schema](schema.Parse(m, &schemaStore, schema.NamingStrategy{}))

	pkf, ok := lo.Find(s.Fields, func(f *schema.Field) bool {
		return f.PrimaryKey
	})
	var pk string
	if ok {
		pk = pkf.Name
	}

	columns := lo.Map(s.Fields, func(field *schema.Field, _ int) column {
		return column{
			"id":          field.DBName, //
			"name":        field.DBName, //
			"description": field.Comment,
			"required":    field.NotNull,
			"choices":     nil,
			"type":        "StringField", // TODO:
			"label":       field.Name,    // TODO: FooBar to Foo Bar
			"widget":      field2widget(field),
			"errors":      nil,
		}
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
	r := row{}
	for _, col := range m.columns {
		col.name()
	}
	return r
}

func (m *model) get_pk_value(row row) any {
	return row.get(m.pk)
}
func (m *model) get_list(db *gorm.DB) ([]row, error) {
	ptr := m.new_slice()
	if err := db.Limit(10).Find(ptr.Interface()).Error; err != nil {
		return nil, err
	}

	// better way?
	len := ptr.Elem().Len()
	res := make([]any, len)
	for i := 0; i < len; i++ {
		item := ptr.Elem().Index(i).Interface()
		res[i] = item
	}

	var rs []row
	mapstructure.Decode(ptr.Interface(), &rs)

	bs := must[[]byte](json.Marshal(res))
	var ms []row
	if err := json.Unmarshal(bs, &ms); err != nil {
		return nil, err
	}
	return ms, nil
}

func (m *model) get(db *gorm.DB, pk any) (row, error) {
	ptr := m.new()
	if err := db.First(ptr, pk).Error; err != nil {
		return nil, err
	}

	var r row
	if err := mapstructure.Decode(ptr, &r); err != nil {
		return nil, err
	}
	return r, nil
}

func (m *model) update(db *gorm.DB, pk any, row row) error {
	ptr := m.new()

	if err := db.Model(ptr).
		Where("?=?", m.pk, pk).
		Updates(map[string]any(row)).
		Error; err != nil {
		return err
	}
	return nil
}
