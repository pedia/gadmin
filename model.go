package gadmin

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/camelcase"
	"github.com/samber/lo"
	"github.com/stoewer/go-strcase"
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

var schemaStore = sync.Map{}

func newModel(m any) *model {
	s := must[*schema.Schema](schema.Parse(m, &schemaStore, schema.NamingStrategy{}))

	columns := lo.Map(s.Fields, func(field *schema.Field, _ int) column {
		return column{
			"name":        field.DBName,
			"description": field.Comment,
			"required":    field.NotNull,
			"choices":     nil,
			"type":        "StringField", // TODO:
			"label":       strings.Join(camelcase.Split(field.Name), " "),
			"widget":      field2widget(field),
			"errors":      nil,
			"primary_key": field.PrimaryKey}
	})

	pk, _ := lo.Find(columns, func(c column) bool {
		return c["primary_key"].(bool)
	})

	return &model{
		typo:    reflect.TypeOf(m),
		schema:  s,
		columns: columns,
		pk:      pk, // TODO: remove
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

// Convert CamelCase to snake_case
func (m *model) name() string { return strcase.SnakeCase(m.schema.Name) }

func (m *model) label() string { return m.schema.Name }

// new t
func (m *model) new() any {
	return reflect.New(m.typo).Interface()
}

// new []t
func (m *model) newSlice() reflect.Value {
	return reflect.New(reflect.SliceOf(m.typo))
}

// Convert value to row
func (m *model) intoRow(ctx context.Context, a any) row {
	v := reflect.ValueOf(a)

	r := row{}
	for _, f := range m.schema.Fields {
		i, _ := f.ValueOf(ctx, v)
		r[f.DBName] = i
	}
	return r
}

func (m *model) find(name string) column {
	if col, ok := lo.Find(m.columns, func(col column) bool {
		return col.name() == name
	}); ok {
		return col
	}
	return nil
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

// apply query
func (m *model) apply(db *gorm.DB, q *Query, count_only bool) *gorm.DB {
	ndb := db
	limit := lo.Ternary(q.PageSize != 0, q.PageSize, q.default_page_size)
	if !count_only {
		ndb = ndb.Limit(limit)

		if q.Page > 0 {
			ndb = ndb.Offset(limit * q.Page)
		}

		if q.Sort != "" {
			column_index := must[int](strconv.Atoi(q.Sort))
			column_name := m.columns[column_index].name()

			ndb = ndb.Order(clause.OrderByColumn{
				Column: clause.Column{Name: column_name},
				Desc:   q.Desc,
			})
		}
	}

	// filter or search
	return ndb
}

func (m *model) get_list(ctx context.Context, db *gorm.DB, q *Query) (int64, []row, error) {
	var total int64
	if err := m.apply(db, q, true).
		Model(m.new()).
		Count(&total).Error; err != nil {
		return 0, nil, err
	}

	ptr := m.newSlice()
	if err := m.apply(db, q, false).
		Find(ptr.Interface()).Error; err != nil {
		return 0, nil, err
	}

	// better way?
	len := ptr.Elem().Len()
	res := make([]row, len)
	for i := 0; i < len; i++ {
		item := ptr.Elem().Index(i).Interface()
		res[i] = m.intoRow(ctx, item)
	}
	return total, res, nil
}

func (m *model) get(ctx context.Context, db *gorm.DB, pk any) (row, error) {
	ptr := m.new()
	if err := db.First(ptr, pk).Error; err != nil {
		return nil, err
	}
	return m.intoRow(ctx, ptr), nil
}

func (m *model) update(db *gorm.DB, pk any, row row) error {
	ptr := m.new()

	if rc := db.Model(ptr).
		Where(map[string]any{m.pk.name(): pk}).
		Updates(row); rc.Error != nil || rc.RowsAffected != 1 {
		return rc.Error
	}
	return nil
}
