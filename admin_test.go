package gadmin

import (
	"database/sql"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func TestApi(t *testing.T) {
	is := assert.New(t)

	db := must[*gorm.DB](gorm.Open(sqlite.Open("db.sqlite"),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			Logger:         logger.Default.LogMode(logger.Info),
		}))

	_ = clause.Associations

	A := NewAdmin("Test Site", db)

	type Foo struct {
		ID           uint `gorm:"primaryKey"`
		Name         string
		Email        *string
		Age          uint8
		Normal       bool
		Valid        *bool `gorm:"default:true"`
		Birthday     *time.Time
		MemberNumber sql.NullString
		ActivatedAt  sql.NullTime
		CreatedAt    time.Time `gorm:"autoCreateTime"`
		UpdatedAt    time.Time `gorm:"autoUpdateTime:nano"`
	}
	fv := NewModelView(Foo{})
	is.Len(fv.GetBlueprint().Children, 9)

	is.Equal("/foo/", fv.GetUrl(".index_view"))
	is.Equal("/foo/action", fv.GetUrl(".action_view"))
	is.Equal("/foo/action?a=b", fv.GetUrl(".action_view", "a", "b"))

	A.AddView(fv)

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

	is.Equal("/admin/foo/", A.UrlFor("foo.index"))
}
