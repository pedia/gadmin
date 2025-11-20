package sqla

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
)

type AllTyped struct {
	ID          uint `gorm:"primaryKey;autoincrement"`
	Name        string
	Email       *string `gorm:"not null"`
	Age         uint8   `gorm:"default:23"`
	IsNormal    bool
	IsFalse     bool  `gorm:"default:false"`
	Valid       *bool `gorm:"default:true"`
	NotNone     *bool `gorm:"not null"`
	Type        string
	Long        string
	Badge       sql.NullString
	BadgeId     sql.NullInt64 `gorm:"default:42"`
	Birthday    *time.Time
	ActivatedAt sql.NullTime
	CreatedAt   time.Time       `gorm:"autoCreateTime"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime:nano"`
	Decimal     decimal.Decimal `gorm:"default:3.14"`
	PtrDecimal  *decimal.Decimal
	Bytes       []byte      `gorm:"size:32"`
	Favorite    null.String `gorm:"default:book"`
	LastLogin   null.Time
}

// belongs to https://gorm.io/docs/belongs_to.html
type Company struct {
	Id   uint `gorm:"primaryKey;autoincrement"`
	Name string
}
type Employee struct {
	Id        uint `gorm:"primaryKey;autoincrement"`
	Name      string
	CompanyId null.Int
	Company   *Company `gorm:"foreignKey:CompanyId"`
}

func (c *Company) String() string { return c.Name }

// has one https://gorm.io/docs/has_one.html
type CreditCard struct {
	Id     uint   `gorm:"primaryKey;autoincrement"`
	Number string `gorm:"size:32"`
	UserID uint
}
type User struct {
	Id         uint `gorm:"primaryKey;autoincrement"`
	Name       string
	CreditCard CreditCard `gorm:"foreignKey:UserID"`
}

// has many https://gorm.io/docs/has_many.html
type Address struct {
	Id        uint `gorm:"primaryKey;autoincrement"`
	Number    string
	AccountID uint
}
type Account struct {
	Id        uint `gorm:"primaryKey;autoincrement"`
	Name      string
	Addresses []Address `gorm:"foreignKey:AccountID"`
}

// many to many https://gorm.io/docs/many_to_many.html
type Language struct {
	Id   uint `gorm:"primaryKey;autoincrement"`
	Name string
}
type Student struct {
	Id        uint `gorm:"primaryKey;autoincrement"`
	Name      string
	Languages []Language `gorm:"many2many:student_language"`
}

// polymorphic https://gorm.io/docs/polymorphism.html
type Toy struct {
	Id        uint `gorm:"primaryKey;autoincrement"`
	Name      string
	OwnerID   int
	OwnerType string
}
type Dog struct {
	Id   uint `gorm:"primaryKey;autoincrement"`
	Name string
	Toys []Toy `gorm:"polymorphic:Owner"`
}

var Models = []any{&AllTyped{},
	Company{}, &Employee{},
	User{}, &CreditCard{},
	Account{}, &Address{},
	Language{}, &Student{},
	Toy{}, &Dog{}}

func ptr[T any](t T) *T {
	return &t
}

var Samples = []any{
	AllTyped{ID: 3, Name: "foo", Email: ptr("foo@a.com"), Age: 42, IsNormal: true,
		Valid: ptr(true), NotNone: ptr(false), Birthday: ptr(time.Now()),
		Badge: sql.NullString{String: "9527", Valid: true}},
	AllTyped{ID: 4, Name: "bar", Email: ptr("bar@a.com"), Age: 21, IsNormal: false,
		Valid: ptr(true), NotNone: ptr(false), Birthday: ptr(time.Now()),
		Badge: sql.NullString{String: "3699", Valid: true}},
	Company{Name: "talk ltd", Id: 31},
	Company{Name: "chat ltd", Id: 32},
	Employee{Name: "Alice", CompanyId: null.NewInt(31, true)},
	Employee{Name: "Bob", CompanyId: null.NewInt(31, true)},
	User{Name: "Alice", Id: 1}, &CreditCard{Number: "2392423948234", UserID: 1},
	Account{Name: "Alice", Id: 1}, &Address{AccountID: 1, Number: "29-1"},
	Language{Name: "french"}, &Student{Name: "alice"},
	Dog{Name: "dog1", Toys: []Toy{{Name: "toy1"}, {Name: "toy2"}}},
}
