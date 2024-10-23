package main

import (
	"time"

	"github.com/glebarez/sqlite"
	"github.com/pedia/gadmin"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/google/uuid"
)

type User struct {
	Id                       string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Type                     string
	EnumChoiceField          string `a:"enum:1=first,2=second;"`
	SqlaUtilsChoiceField     string
	SqlaUtilsEnumChoiceField int
	FirstName                string
	LastName                 string
	Email                    string
	Website                  string
	IpAddress                string `gorm:"comment:last logined ip address"`
	Currency                 string
	Timezone                 string
	DiallingCode             int
	LocalPhoneNumber         string
	FeaturedPostId           int
}

type Post struct {
	Id              int    `gorm:"primaryKey"`
	Title           string `gorm:"size:120"`
	Text            string
	Date            time.Time
	BackgroundColor string
	CreatedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UserId          uuid.UUID
	User            *User `gorm:"foreignKey:UserId"`
}

type Tag struct {
	Id   int    `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex"`
}

func main() {
	db, _ := gorm.Open(sqlite.Open("db.sqlite"),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			Logger:         logger.Default.LogMode(logger.Info),
		})

	admin := gadmin.NewAdmin("Example: Simple Views", db)
	// admin.AddView(gadmin.NewView(gadmin.MenuItem{Category: "Test", Name: "View1"}))
	// admin.AddView(gadmin.NewView(gadmin.MenuItem{Category: "Test", Name: "View2"}))

	mv := gadmin.NewModelView(User{}).
		SetColumnList("type", "first_name", "last_name", "email", "ip_address", "currency", "timezone", "phone_number").
		SetColumnEditableList("first_name", "type", "currency", "timezone").
		SetColumnDescriptions(map[string]string{"first_name": "名"}).
		SetCanSetPageSize(true).
		SetPageSize(5).
		SetTablePrefixHtml(`<h4>hello</h4>`)
	admin.AddView(mv)
	admin.AddView(gadmin.NewModelView(Post{}))
	admin.AddView(gadmin.NewModelView(Tag{}))
	admin.Run()
}
