package main

import (
	"database/sql"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/pedia/gadmin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	db, _ := gorm.Open(sqlite.Open("../db.sqlite"),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			Logger:         logger.Default.LogMode(logger.Info),
		})
	A := gadmin.NewAdmin("Test Site", db)

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

	A.AddView(gadmin.NewModelView(Foo{}))

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
	A.AddView(gadmin.NewModelView(Company{}, "Association"))
	A.AddView(gadmin.NewModelView(Employee{}, "Association"))

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
	A.AddView(gadmin.NewModelView(CreditCard{}, "Association"))
	A.AddView(gadmin.NewModelView(User{}, "Association"))

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
	A.AddView(gadmin.NewModelView(Address{}, "Association"))
	A.AddView(gadmin.NewModelView(Account{}, "Association"))

	// many to many https://gorm.io/docs/many_to_many.html
	type Language struct {
		gorm.Model
		Name string
	}
	type Student struct {
		gorm.Model
		Languages []Language `gorm:"many2many:student_language"`
	}
	A.AddView(gadmin.NewModelView(Language{}, "Association"))
	A.AddView(gadmin.NewModelView(Student{}, "Association"))

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
	A.AddView(gadmin.NewModelView(Toy{}, "Association"))
	A.AddView(gadmin.NewModelView(Dog{}, "Association"))

	A.Run()
}
