package gadmin

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Foo struct {
	ID           uint `gorm:"primaryKey"`
	Name         string
	Email        *string
	Age          uint8
	Normal       bool
	Valid        *bool `gorm:"default:true"`
	MemberNumber sql.NullString
	Birthday     *time.Time
	ActivatedAt  sql.NullTime
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime:nano"`

	Decimal decimal.Decimal
}

// belongs to https://gorm.io/docs/belongs_to.html
type Company struct {
	Id   int
	Name string
}
type Employee struct {
	Id        int
	Name      string
	CompanyId int
	Company   *Company
}

// has one https://gorm.io/docs/has_one.html
type CreditCard struct {
	gorm.Model
	Number string
	UserID uint
}
type User struct {
	gorm.Model
	CreditCard CreditCard
}

// has many https://gorm.io/docs/has_many.html
type Address struct {
	gorm.Model
	Number    string
	AccountID uint
}
type Account struct {
	gorm.Model
	Addresses []Address
}

// many to many https://gorm.io/docs/many_to_many.html
type Language struct {
	gorm.Model
	Name string
}
type Student struct {
	gorm.Model
	Languages []Language `gorm:"many2many:student_language"`
}

// polymorphic https://gorm.io/docs/polymorphism.html
type Toy struct {
	ID        int
	Name      string
	OwnerID   int
	OwnerType string
}
type Dog struct {
	ID   int
	Name string
	Toys []Toy `gorm:"polymorphic:Owner"`
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

type ModelTestSuite struct {
	suite.Suite
	admin   *Admin
	fooView *ModelView
}

func (S *ModelTestSuite) SetupTest() {
	db, _ := gorm.Open(sqlite.Open("../db.sqlite"),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			Logger:         logger.Default.LogMode(logger.Info),
		})
	A := NewAdmin("Test Site", db)

	var c int64
	db.Model(&Foo{}).Count(&c)
	if c == 0 {
		e1 := "foo@foo.com"
		d1 := time.Date(2024, 10, 1, 0, 0, 0, 0, nil)
		e2 := "bar@foo.com"
		d2 := time.Date(2024, 3, 1, 0, 0, 0, 0, nil)
		fs := []Foo{
			{Name: "foo", Email: &e1, Age: 42, Normal: true, Birthday: &d1,
				MemberNumber: sql.NullString{String: "9527", Valid: true}},
			{Name: "bar", Email: &e2, Age: 21, Normal: false, Birthday: &d2,
				MemberNumber: sql.NullString{String: "3699", Valid: true}},
		}
		db.Create(&fs)
	}

	S.fooView = A.AddView(NewModelView(Foo{})).(*ModelView)

	A.AddView(NewModelView(Company{}, "Association"))
	A.AddView(NewModelView(Employee{}, "Association"))

	A.AddView(NewModelView(CreditCard{}, "Association"))
	A.AddView(NewModelView(User{}, "Association"))

	A.AddView(NewModelView(Address{}, "Association"))
	A.AddView(NewModelView(Account{}, "Association"))

	A.AddView(NewModelView(Language{}, "Association"))
	A.AddView(NewModelView(Student{}, "Association"))

	A.AddView(NewModelView(Toy{}, "Association"))
	A.AddView(NewModelView(Dog{}, "Association"))

	S.admin = A
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, new(ModelTestSuite))
}

func (S *ModelTestSuite) TestModelView() {
	is := assert.New(S.T())

	v := NewModelView(Foo{})

	is.NotEmpty(v.GetBlueprint().Children)

	is.Equal("/foo/", v.GetUrl(".index_view", nil))
	is.Equal("/foo/action", v.GetUrl(".action_view", nil))
	is.Equal("/foo/action?a=b", v.GetUrl(".action_view", nil, "a", "b"))

	is.Equal([]string{"id", "name", "email", "age", "normal", "valid", "member_number", "birthday", "activated_at", "created_at", "updated_at", "decimal"}, v.column_list)
	is.Equal([]string{"id", "name", "email", "age", "normal", "valid", "member_number", "birthday", "activated_at", "created_at", "updated_at", "decimal"}, v.column_sortable_list)

	// query
	r1 := httptest.NewRequest("", "/admin/tag/?sort=0&desc=1&page_size=23&page=2", nil)
	q1 := v.queryFrom(r1)
	is.Equal("0", q1.Sort)
	is.Equal(true, q1.Desc)
	is.Equal(23, q1.PageSize)
	is.Equal(2, q1.Page)

	r2 := httptest.NewRequest("", "/admin/tag/?sort=1", nil)
	q2 := v.queryFrom(r2)
	is.Equal("1", q2.Sort)
	is.Equal(false, q2.Desc)
	is.Equal(0, q2.PageSize)
	is.Equal(0, q2.Page)

	r3 := httptest.NewRequest("", "/admin/tag/details?id=6&url=%2Fadmin%2Ftag%2F%3Fdesc%3D1%26sort%3D1", nil)
	q3 := v.queryFrom(r3)
	is.Equal("", q3.Sort)
	is.Equal(false, q3.Desc)
	is.Equal(0, q3.PageSize)
	is.Equal(0, q3.Page)
	is.Equal("6", q3.Get("id"))
	is.Equal("/admin/tag/?desc=1&sort=1", q3.Get("url"))
}
func (S *ModelTestSuite) TestUrl() {
	is := assert.New(S.T())

	is.Equal("/admin/foo/", lo.Must[string](S.admin.UrlFor("", "foo.index")))
	is.Equal("/admin/foo/", lo.Must[string](S.admin.UrlFor("foo", ".index")))

	is.Equal("/admin/foo/?a=1", lo.Must[string](S.admin.UrlFor("", "foo.index", "a", 1)))
	is.Equal("/admin/foo/?page=3", lo.Must[string](S.admin.UrlFor("foo", ".index", "page", 3)))

	is.Equal("/admin/foo/export?export_type=csv", S.fooView.GetUrl(".export", nil, "export_type", "csv"))
	is.Equal("/admin/foo/?page_size=0", S.fooView.GetUrl(".index_view", nil, "page_size", 0)) // bad page_size

	es := []string{
		"company", "employee",
		"credit_card", "user", "address", "account",
		"language", "student", "toy", "dog",
	}

	paths := lo.FlatMap(es, func(e string, _ int) []string {
		return []string{
			fmt.Sprintf("/admin/%s/", e),
			// fmt.Sprintf("/admin/%s/edit", e),
			// fmt.Sprintf("/admin/%s/details", e),
			// fmt.Sprintf("/admin/%s/action", e),
			// fmt.Sprintf("/admin/%s/delete", e),
			// fmt.Sprintf("/admin/%s/export", e),
		}
	})

	for _, path := range paths {
		r := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		S.admin.ServeHTTP(w, r)
		is.Equal(200, w.Code, path)
	}
}
