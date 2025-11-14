package gadm

import (
	"database/sql"
	"fmt"
	"gadm/examples/sqla"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func views(db *gorm.DB) []*ModelView {
	return []*ModelView{
		NewModelView(sqla.AllTyped{}, db),
		NewModelView(sqla.Company{}, db, "Association"),
		NewModelView(sqla.Employee{}, db, "Association"),
		NewModelView(sqla.CreditCard{}, db, "Association"),
		NewModelView(sqla.User{}, db, "Association"),
		NewModelView(sqla.Address{}, db, "Association"),
		NewModelView(sqla.Account{}, db, "Association"),
		NewModelView(sqla.Language{}, db, "Association"),
		NewModelView(sqla.Student{}, db, "Association"),
		NewModelView(sqla.Toy{}, db, "Association"),
		NewModelView(sqla.Dog{}, db, "Association"),
	}
}

func typeds() []sqla.AllTyped {
	e1 := "foo@foo.com"
	d1 := time.Date(2024, 10, 1, 0, 0, 0, 0, time.Local)
	e2 := "bar@foo.com"
	d2 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	return []sqla.AllTyped{
		{ID: 3, Name: "foo", Email: &e1, Age: 42, IsNormal: true, Birthday: &d1,
			Badge: sql.NullString{String: "9527", Valid: true}},
		{ID: 4, Name: "bar", Email: &e2, Age: 21, IsNormal: false, Birthday: &d2,
			Badge: sql.NullString{String: "3699", Valid: true}},
	}
}

func TestModel(t *testing.T) {
	is := assert.New(t)

	m := NewModel(sqla.AllTyped{})
	is.Equal("all_typed", m.name())
	is.Equal("All Typed", m.label())
	is.Equal("alltyped", m.path())

	is.Equal("ID", m.Fields[0].Label)
	is.Equal("id", m.Fields[0].DBName)
	is.Equal("ID", m.Fields[0].Name)
	is.Equal("Email", m.Fields[2].Label)
	is.Equal("Activated At", m.Fields[10].Label)
	is.Equal("activated_at", m.Fields[10].DBName)
	is.Equal("ActivatedAt", m.Fields[10].Name)

	r1 := m.intoRow(typeds()[0])
	is.Equal("foo", r1["name"])
	is.True(r1["is_normal"].(bool))

	is.Equal("3,foo", m.get_pk_value(r1))
	is.Equal(map[string]string{"id": "3", "name": "foo"}, m.where("3,foo"))
}

func TestWidget(t *testing.T) {
	// is := assert.New(t)
	m := NewModel(sqla.AllTyped{})
	ModelForm(m.Fields, "xx")
}

type ModelTestSuite struct {
	suite.Suite
	is        *assert.Assertions
	admin     *Admin
	typedView *ModelView
}

func (ts *ModelTestSuite) SetupTest() {
	ts.is = assert.New(ts.T())

	db, _ := gorm.Open(sqlite.Open(":memory:"),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			Logger:         logger.Default.LogMode(logger.Info),
		})
	ts.admin = NewAdmin("Test Site")

	var c int64
	tx := db.Model(&sqla.Company{}).Count(&c)
	if tx.Error != nil || c == 0 {
		db.AutoMigrate(sqla.Models...)

		// e1 := "foo@foo.com"
		// d1 := time.Date(2024, 10, 1, 0, 0, 0, 0, time.Local)
		// e2 := "bar@foo.com"
		// d2 := time.Date(2024, 3, 1, 0, 0, 0, 0, time.Local)
		// fs := []Typed{
		// 	{Name: "foo", Email: &e1, Age: 42, Normal: true, Birthday: &d1,
		// 		MemberNumber: sql.NullString{String: "9527", Valid: true}},
		// 	{Name: "bar", Email: &e2, Age: 21, Normal: false, Birthday: &d2,
		// 		MemberNumber: sql.NullString{String: "3699", Valid: true}},
		// }
		// db.Create(&fs)

		samples := []any{
			&sqla.Company{Name: "talk ltd"},
			&sqla.Company{Name: "chat ltd"},
			&sqla.Employee{Name: "Alice", CompanyId: 1},
			&sqla.Employee{Name: "Bob", CompanyId: 1},
		}
		for _, o := range samples {
			tx := db.Create(o)
			if tx.Error != nil {
				panic(tx.Error)
			}
		}
	}

	for _, v := range views(db) {
		ts.admin.AddView(v)
	}

	ts.typedView = ts.admin.FindView("alltyped").(*ModelView)
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, new(ModelTestSuite))
}

func (ts *ModelTestSuite) TestRelations() {
	// ve := NewModelView(Employee{}, db, "Association").Joins("Company")
	// S.admin.AddView(ve)
	// r := ve.list(DefaultQuery())
	// S.assert.Nil(r.Error)
	// S.assert.Len(r.Rows, 2)
	// S.assert.Equal(int64(2), r.Total)
}

func (ts *ModelTestSuite) TestAdmin() {
	ts.is.NotNil(ts.admin.FindView("alltyped"))
}

func (ts *ModelTestSuite) TestModelView() {
	v := ts.typedView

	ts.is.NotEmpty(v.GetBlueprint().Children)

	ts.is.Equal("/admin/alltyped/", must(v.Blueprint.GetUrl(".index_view")))
	ts.is.Equal("/admin/alltyped/action", must(v.Blueprint.GetUrl(".action_view")))
	ts.is.Equal("/admin/alltyped/action?a=b", must(v.Blueprint.GetUrl(".action_view", "a", "b")))

	// query
	r1 := httptest.NewRequest("", "/admin/tag/?sort=0&desc=1&page_size=23&page=2", nil)
	q1 := v.queryFrom(r1)
	ts.is.Equal("0", q1.Sort)
	ts.is.Equal(true, q1.Desc)
	ts.is.Equal(23, q1.PageSize)
	ts.is.Equal(2, q1.Page)

	r2 := httptest.NewRequest("", "/admin/tag/?sort=1", nil)
	q2 := v.queryFrom(r2)
	ts.is.Equal("1", q2.Sort)
	ts.is.Equal(false, q2.Desc)
	ts.is.Equal(20, q2.PageSize)
	ts.is.Equal(0, q2.Page)

	r3 := httptest.NewRequest("", "/admin/tag/details?id=6&url=%2Fadmin%2Ftag%2F%3Fdesc%3D1%26sort%3D1", nil)
	q3 := v.queryFrom(r3)
	ts.is.Equal("", q3.Sort)
	ts.is.Equal(false, q3.Desc)
	ts.is.Equal(20, q3.PageSize)
	ts.is.Equal(0, q3.Page)
	ts.is.Equal("6", q3.Get("id"))
	ts.is.Equal("/admin/tag/?desc=1&sort=1", q3.Get("url"))
}

// func (S *ModelTestSuite) TestSession() {
// 	is := assert.New(S.T())
// 	S.admin.Register(&Blueprint{Endpoint: "bar", Path: "/bar",
// 		Handler: func(w http.ResponseWriter, r *http.Request) {
// 			CurrentSession(r).Set("bar", "hello")
// 		}})
// 	r := httptest.NewRequest("GET", "/admin/bar", nil)
// 	w := httptest.NewRecorder()
// 	S.admin.ServeHTTP(w, r)
// 	is.Equal(200, w.Code)
// }

func (ts *ModelTestSuite) TestUrlStatusCode() {
	ts.is.Equal("/admin/alltyped/", must(ts.admin.UrlFor("", "alltyped.index")))

	es := []string{
		"alltyped",
		"company", "employee",
		"creditcard", "user", "address", "account",
		"language", "student", "toy", "dog",
	}

	paths := lo.FlatMap(es, func(e string, _ int) []string {
		return []string{
			fmt.Sprintf("/admin/%s/", e),
			fmt.Sprintf("/admin/%s/new", e),
			// 302 fmt.Sprintf("/admin/%s/edit", e),
			// fmt.Sprintf("/admin/%s/details", e),
			// fmt.Sprintf("/admin/%s/action", e),
			// fmt.Sprintf("/admin/%s/delete", e),
			fmt.Sprintf("/admin/%s/export", e),
		}
	})

	for _, path := range paths {
		r := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		ts.admin.ServeHTTP(w, r)
		ts.is.Equal(200, w.Code, path)
	}
}
