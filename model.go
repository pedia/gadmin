package gadmin

import (
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/fatih/camelcase"
	"github.com/fatih/structs"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"gorm.io/gorm/schema"
)

type Row map[string]any

func (r Row) Get(c Column) any {
	return r[c.DBName]
}

type Model struct {
	typo    reflect.Type
	schema  *schema.Schema
	columns []Column
	pk      Column // TODO: multiple primary keys
	// index column
}

var schemaStore = sync.Map{}

func NewModel(m any) *Model {
	s := must(schema.Parse(m, &schemaStore, schema.NamingStrategy{}))

	columns := lo.Map(s.Fields, func(field *schema.Field, _ int) Column {
		return NewColumn(field)
	})

	// TODO: multiple primary key
	pk, _ := lo.Find(columns, func(c Column) bool {
		return c.PrimaryKey
	})

	return &Model{
		typo:    s.ModelType,
		schema:  s,
		columns: columns,
		pk:      pk,
	}
}

// TODO: more field type
func field2widget(field *schema.Field) *field {
	table := map[reflect.Kind]string{
		reflect.String: "text",
		// reflect.:"password",
		// reflect.:"hidden",
		reflect.Bool: "checkbox",
		// reflect.:"radio",
		// reflect.:"file",
		// reflect.:"submit",
	}

	return NewField([]lo.Entry[string, any]{
		{Key: "type", Value: table[field.FieldType.Kind()]},
	})
}

// Convert CamelCase to snake_case
func (m *Model) name() string { return strcase.SnakeCase(m.schema.Name) }

func (m *Model) label() string { return m.schema.Name }

// new t
func (m *Model) new() any {
	return reflect.New(m.typo).Interface()
}

// new []t
func (m *Model) newSlice() reflect.Value {
	return reflect.New(reflect.SliceOf(m.typo))
}

// Parse form into map[string]any
func (m *Model) parseForm(uv url.Values) Row {
	res := map[string]any{}
	for _, col := range m.columns {
		name := col.Name
		if uv.Has(name) {
			res[name] = uv.Get(name)
		}
	}
	return res
}

// Convert struct value to row
func (M *Model) intoRow(a any) Row {
	m := structs.Map(a)

	res := map[string]any{}
	for _, c := range M.columns {
		res[c.DBName] = m[c.Name] // LastName -> last_name
	}
	return res
}

func (m *Model) find(name string) *Column {
	if col, ok := lo.Find(m.columns, func(col Column) bool {
		return col.DBName == name
	}); ok {
		return &col
	}
	return nil
}

// Return all field can be sorted
// exclude relationship fields
func (m *Model) sortable_list() []string {
	cols := lo.Filter(m.columns, func(col Column, _ int) bool {
		_, ok := m.schema.Relationships.Relations[col.Label]
		return !ok
	})
	return lo.Map(cols, func(col Column, _ int) string {
		return col.Name
	})
}
func (m *Model) get_pk_value(row Row) any {
	return row.Get(m.pk)
}

type Choice struct {
	Text  string
	Value any
}

type Column struct {
	Name        string
	DBName      string
	Description string
	Required    bool
	Choices     []Choice
	Type        string
	Label       string
	Widget      *field
	Error       string
	PrimaryKey  bool
	Value       any
}

func NewColumn(field *schema.Field) Column {
	return Column{
		Name:        field.Name,   // EnumChoiceField
		DBName:      field.DBName, // enum_choice_field
		Description: field.Comment,
		Required:    field.NotNull,
		Choices:     nil,
		Type:        string(field.DataType),
		Label:       strings.Join(camelcase.Split(field.Name), " "), // Enum Choice Field
		Widget:      field2widget(field),
		Error:       "",
		PrimaryKey:  field.PrimaryKey,
	}
}
