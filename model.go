package gadm

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fatih/camelcase"
	"github.com/samber/lo"
	"github.com/spf13/cast"
	"gorm.io/gorm/schema"
)

type Row struct {
	// db -> Row
	v      any // struct ptr
	rv     reflect.Value
	fields []*Field
	// form -> map -> db
	// difficult to set struct field, use map instead
	m map[string]any
}

func newRow(v any, fields []*Field) *Row {
	row := &Row{v: v, fields: fields, m: map[string]any{}}

	rv := reflect.Indirect(reflect.ValueOf(v))

	row.v = rv.Interface()
	row.rv = rv
	return row
}

func fieldValue(a any, name string) any {
	rv := reflect.ValueOf(a)
	for rv.Kind() == reflect.Pointer {
		v := rv.Elem()
		if !v.IsValid() {
			return nil
		}
		rv = reflect.ValueOf(v.Interface())
	}

	if rv.Kind() == reflect.Struct {
		if v := rv.FieldByName(name); v.IsValid() {
			return v.Interface()
		}
	} else if rv.Kind() == reflect.Slice {
		return rv.Interface()
	}
	return nil
}

func (r *Row) Set(field *Field, v any) {
	// How to set struct field?
	r.m[field.DBName] = v
}

func (r *Row) Get(f *Field) any {
	v := r.rv.FieldByName(f.Name)
	if v.IsValid() {
		return v.Interface()
	}
	return r.rv.FieldByName(f.Name).Interface()
}

func (r *Row) GetDisplayValue(f *Field) any {
	v := r.Get(f)

	vf := &Field{Field: f.Field, Value: v}
	return vf.Display()
}

func (r *Row) GetPkValue(f *Field) string {
	if v := r.Get(f); v != nil {
		nf := &Field{Field: f.Field, Value: v}
		return nf.GetPkValue()
	}
	return ""
}

func (r *Row) Of(f *Field) *Field {
	return &Field{Field: f.Field, Value: r.Get(f)}
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
	return &Model{schema: s, Fields: fs}
}

// like:
// user: schema.NamingStrategy{SingularTable: true}
// users: schema.NamingStrategy{SingularTable: false}
// used in Blueprint's {Path/Endpoint}
func (m *Model) name() string { return m.schema.Table }

func (m *Model) endpoint() string { return strings.ReplaceAll(m.schema.Table, "_", "") }

func (m *Model) label() string { return strings.Join(camelcase.Split(m.schema.Name), " ") }

// new t
func (m *Model) new() any {
	return reflect.New(m.schema.ModelType).Interface()
}

// new []t
func (m *Model) newSlice() reflect.Value {
	return reflect.New(reflect.SliceOf(m.schema.ModelType))
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

// TODO: null
// in detail/edit url is: id=pk1,pk2
func (m *Model) get_pk_value(row *Row) string {
	vs := []string{}
	for _, pkf := range m.schema.PrimaryFields {
		v := row.Get(&Field{Field: pkf})
		vs = append(vs, cast.ToString(v))
	}
	return strings.Join(vs, ",")
}

// single primarykey, rowid: id
// multiple primarykey, rowid like: pk1,pk2
// return where condition map
func (m *Model) where(rowid string) map[string]string {
	vs := strings.Split(rowid, ",")
	res := map[string]string{}
	if len(m.schema.PrimaryFields) == len(vs) {
		for i := range vs {
			res[m.schema.PrimaryFields[i].DBName] = vs[i]
		}
	}
	return res
}

type Choice struct {
	Value any    `json:"value"`
	Label string `json:"text"` // in flask
}

type Field struct {
	*schema.Field
	Label       string
	Choices     []Choice
	Description string
	TextAreaRow int
	TimeFormat  string
	Readonly    bool // primary key
	Hidden      bool // csrf token
	Value       any
}

func (f *Field) Endpoint() string {
	if f.Schema == nil {
		panic("not refer field")
	}
	ns := schema.NamingStrategy{SingularTable: true}
	tn := ns.TableName(f.Schema.Table)
	return strings.ReplaceAll(tn, "_", "")
}
func (f *Field) IsSlice() bool {
	rv := reflect.ValueOf(f.Value)
	return rv.Kind() == reflect.Slice
}
func (f *Field) IsStruct() bool {
	return f.DBName == "" && f.Schema != nil && !f.IsSlice()
}
func (f *Field) GetPkValue() string {
	vs := []string{}
	for _, pkf := range f.Schema.PrimaryFields {
		v := fieldValue(f.Value, pkf.Name)
		vs = append(vs, cast.ToString(v))
	}
	return strings.Join(vs, ",")
}

func isNil(object interface{}) bool {
	if object == nil {
		return true
	}

	rv := reflect.ValueOf(object)
	return slices.Contains(
		[]reflect.Kind{reflect.Chan, reflect.Func, reflect.Interface,
			reflect.Map, reflect.Ptr, reflect.Slice, reflect.UnsafePointer},
		rv.Kind()) && rv.IsNil()
}

// Edit or display in HTML
// Empty string means nil = sql null, null.String.Valid = false
func (f *Field) Display() string {
	// refer field
	if f.DBName == "" && f.DataType == "" {
		if !isNil(f.Value) {
			if str, ok := f.Value.(fmt.Stringer); ok && str != nil {
				return str.String()
			}
		}
		return ""
	}

	switch v := f.Value.(type) {
	case driver.Valuer:
		dv, _ := v.Value()
		if dv == nil {
			return ""
		}
		return cast.ToString(dv)
	case string:
		return v
	case *string:
		if v != nil {
			return *v
		}
		return ""
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, []byte:
		return cast.ToString(v)
	case bool:
		return f.displayBool(v)
	case *bool:
		if v != nil {
			return f.displayBool(*v)
		}
		return ""
	case time.Time:
		return f.displayTime(v)
	case *time.Time:
		if v != nil {
			return f.displayTime(*v)
		}
		return ""
	default:
		fmt.Printf("todo: %v %t", v, v)
	}
	return ""
}

func (f *Field) displayBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func (f *Field) displayTime(t time.Time) string {
	switch f.TimeFormat {
	default:
		return t.Format(time.DateOnly)
	case "YYYY-MM-DD":
		return t.Format(time.DateOnly)
	case "YYYY-MM-DD HH:mm:ss":
		return t.Format(time.DateTime)
	case "HH:mm:ss":
		return t.Format(time.TimeOnly)
	}
}
