package gadmin

import (
	"net/http/httptest"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func TestApi(t *testing.T) {
	is := assert.New(t)

	db := must[*gorm.DB](gorm.Open(
		sqlite.Open("db.sqlite"),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			Logger:         logger.Default.LogMode(logger.Info),
		}))

	_ = clause.Associations

	A := NewAdmin("Test Site", db)

	fv := NewModelView(Foo{})
	is.NotEmpty(fv.GetBlueprint().Children)

	is.Equal("/foo/", fv.GetUrl(".index_view", nil))
	is.Equal("/foo/action", fv.GetUrl(".action_view", nil))
	is.Equal("/foo/action?a=b", fv.GetUrl(".action_view", nil, "a", "b"))

	A.AddView(fv)
	is.NotNil(fv.admin)

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
	A.AddView(NewModelView(Company{}, "Association"))
	A.AddView(NewModelView(Employee{}, "Association"))

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
	A.AddView(NewModelView(CreditCard{}, "Association"))
	A.AddView(NewModelView(User{}, "Association"))

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
	A.AddView(NewModelView(Address{}, "Association"))
	A.AddView(NewModelView(Account{}, "Association"))

	// many to many https://gorm.io/docs/many_to_many.html
	type Language struct {
		gorm.Model
		Name string
	}
	type Student struct {
		gorm.Model
		Languages []Language `gorm:"many2many:student_language"`
	}
	A.AddView(NewModelView(Language{}, "Association"))
	A.AddView(NewModelView(Student{}, "Association"))

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
	A.AddView(NewModelView(Toy{}, "Association"))
	A.AddView(NewModelView(Dog{}, "Association"))

	is.Equal("/admin/foo/", must[string](A.UrlFor("", "foo.index")))
	is.Equal("/admin/foo/", must[string](A.UrlFor("foo", ".index")))

	is.Equal("/admin/foo/?a=1", must[string](A.UrlFor("", "foo.index", "a", 1)))
	is.Equal("/admin/foo/?page=3", must[string](A.UrlFor("foo", ".index", "page", 3)))

	is.Equal("/admin/foo/export?export_type=csv", fv.GetUrl(".export", nil, "export_type", "csv"))
	is.Equal("/admin/foo/?page_size=0", fv.GetUrl(".index_view", nil, "page_size", 0)) // bad page_size

	paths := []string{
		"/admin/foo/",
		"/admin/foo/new",
		"/admin/company/",
		"/admin/company/new",
		"/admin/employee/",
		"/admin/employee/new",
		"/admin/credit_card/",
		"/admin/credit_card/new",
		"/admin/user/",
		"/admin/user/new",
		"/admin/address/",
		"/admin/address/new",
		"/admin/account/",
		"/admin/account/new",
		"/admin/language/",
		"/admin/language/new",
		"/admin/student/",
		"/admin/student/new",
		"/admin/dog/",
		"/admin/dog/new",
		"/admin/toy/",
		"/admin/toy/new",
	}
	for _, path := range paths {
		r := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		A.ServeHTTP(w, r)
		is.Equal(200, w.Code)
	}
}
