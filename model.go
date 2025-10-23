package gadmin

import (
	"context"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/camelcase"
	"github.com/fatih/structs"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Row map[string]any

func (r Row) Get(c Column) any {
	return r[c.Name]
}

type Model struct {
	typo    reflect.Type
	schema  *schema.Schema
	columns []Column
	pk      Column // TODO: multiple primary keys
	// index column
}

var schemaStore = sync.Map{}

func newModel(m any) *Model {
	s := must(schema.Parse(m, &schemaStore, schema.NamingStrategy{}))

	columns := lo.Map(s.Fields, func(field *schema.Field, _ int) Column {
		return newColumn(field)
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
	for k, v := range m {
		res[strings.ToLower(k)] = v
	}
	return res
}

func (m *Model) find(name string) *Column {
	if col, ok := lo.Find(m.columns, func(col Column) bool {
		return col.Name == name
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

// apply query
func (m *Model) apply(db *gorm.DB, q *Query, count_only bool) *gorm.DB {
	ndb := db
	limit := lo.Ternary(q.PageSize != 0, q.PageSize, q.default_page_size)
	if !count_only {
		ndb = ndb.Limit(limit)

		if q.Page > 0 {
			ndb = ndb.Offset(limit * q.Page)
		}

		if q.Sort != "" {
			column_index := must(strconv.Atoi(q.Sort))
			column_name := m.columns[column_index].Name

			ndb = ndb.Order(clause.OrderByColumn{
				Column: clause.Column{Name: column_name},
				Desc:   q.Desc,
			})
		}
	}

	// filter or search
	return ndb
}

// Joins/InnerJoins/Preload
func (m *Model) get_one(ctx context.Context, db *gorm.DB, pk any) (Row, error) {
	// TODO: multiple primary keys `a,b`
	ptr := m.new()
	if err := db.First(ptr, pk).Error; err != nil {
		return nil, err
	}
	return m.intoRow(ptr), nil
}

func (m *Model) update(db *gorm.DB, pk any, row map[string]any) error {
	ptr := m.new()

	if rc := db.Model(ptr).
		Where(map[string]any{m.pk.Name: pk}).
		Updates(row); rc.Error != nil || rc.RowsAffected != 1 {
		return rc.Error
	}
	return nil
}

// row -> Model().Create() RETURNING *
func (m *Model) create(db *gorm.DB, row map[string]any) error {
	ptr := m.new()

	if rc := db.Model(ptr).
		Clauses(clause.Returning{}). // RETURNING *
		Create(row); rc.Error != nil || rc.RowsAffected != 1 {
		return rc.Error
	}
	return nil
}

type Choice struct {
	Text  string
	Value any
}

type Column struct {
	Name        string
	Description string
	Required    bool
	Choices     []Choice
	Type        string
	Label       string
	Widget      *field
	Errors      string
	PrimaryKey  bool
	Value       any
}

func newColumn(field *schema.Field) Column {
	return Column{
		Name:        field.DBName,
		Description: field.Comment,
		Required:    field.NotNull,
		Choices:     nil,
		Type:        string(field.DataType),
		Label:       strings.Join(camelcase.Split(field.Name), " "),
		Widget:      field2widget(field),
		Errors:      "",
		PrimaryKey:  field.PrimaryKey,
	}
}
