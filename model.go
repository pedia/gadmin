package gadmin

import (
	"net/url"
	"reflect"
	"strings"
	"sync"

	"github.com/fatih/camelcase"
	"github.com/fatih/structs"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"gorm.io/gorm/schema"
)

type Row map[string]any

func (r Row) Get(f *Field) any {
	return r[f.DBName]
}

// TODO: return DisplayValue for nil ptr, *time.Time format
func (r Row) GetDiplayValue(f *Field) any {
	return r[f.DBName]
}

type Model struct {
	schema *schema.Schema
	Fields []*Field
}

var schemaStore = sync.Map{}

func NewModel(m any) *Model {
	// TODO: option SingularTable
	s := must(schema.Parse(m, &schemaStore, schema.NamingStrategy{SingularTable: true}))
	fs := lo.Map(s.Fields, func(field *schema.Field, _ int) *Field {
		return &Field{Field: field, Label: strings.Join(camelcase.Split(field.Name), " ")}
	})
	if len(fs) > 4 {
		fs[4].Description = "Some important" // TODO: hack
	}
	return &Model{schema: s, Fields: fs}
}

// like:
// user: schema.NamingStrategy{SingularTable: true}
// users: schema.NamingStrategy{SingularTable: false}
// used in Blueprint's {Path/Endpoint}
func (m *Model) name() string { return m.schema.Table }

func (m *Model) label() string { return strings.Join(camelcase.Split(m.schema.Name), " ") }

// new t
func (m *Model) new() any {
	return reflect.New(m.schema.ModelType).Interface()
}

// new []t
func (m *Model) newSlice() reflect.Value {
	return reflect.New(reflect.SliceOf(m.schema.ModelType))
}

// Parse form into map[string]any, only fields in current model
func (m *Model) parseForm(uv url.Values) Row {
	row := Row{}
	for _, col := range m.schema.Fields {
		name := col.Name
		if uv.Has(name) {
			row[name] = uv.Get(name)
		}
	}
	return row
}

// Convert struct value to row
func (M *Model) intoRow(a any) Row {
	m := structs.Map(a)

	res := map[string]any{}
	for _, c := range M.schema.Fields {
		res[c.DBName] = m[c.Name] // LastName -> last_name
	}
	// TODO: join
	return res
}

func (m *Model) find(name string) *Field {
	if f, ok := lo.Find(m.Fields, func(field *Field) bool {
		return field.DBName == name
	}); ok {
		return f
	}
	return nil
}

// Return all field can be sorted
// exclude relationship fields
func (m *Model) sortableColumns() []string {
	return m.schema.DBNames
}

// in detail/edit url is: id=pk1,pk2
func (m *Model) get_pk_value(row Row) string {
	vs := []string{}
	for _, pkf := range m.schema.PrimaryFields {
		v := row[pkf.DBName]
		vs = append(vs, cast.ToString(v))
	}
	return strings.Join(vs, ",")
}

// return where condition map
func (m *Model) where(rowid string) map[string]string {
	vs := strings.Split(rowid, ",")
	res := map[string]string{}
	if len(m.schema.PrimaryFields) == len(vs) {
		for i := range vs {
			res[m.schema.PrimaryFields[0].DBName] = vs[i]
		}
	}
	return res
}

type Choice struct {
	Value any
	Label string
}

type Field struct {
	*schema.Field
	Label       string
	Choices     []Choice
	Description string
	TextAreaRow int
	TimeFormat  string
	Hidden      bool // csrf token
}
