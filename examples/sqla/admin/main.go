package main

import (
	"time"

	"gadmin"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type User struct {
	Id                       string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Type                     string `gorm:"size:100"`
	EnumChoiceField          string `gorm:"size:6"`
	SqlaUtilsChoiceField     string `gorm:"size:255"`
	SqlaUtilsEnumChoiceField int
	FirstName                string     `gorm:"size:100"`
	LastName                 string     `gorm:"size:100"`
	Email                    string     `gorm:"size:255;not null"`
	Valid                    bool       `gorm:"not null"`
	BornDate                 *time.Time `gorm:";"`
	Website                  string
	IpAddress                string `gorm:"size:50"`
	Currency                 string `gorm:"size:3"`
	Timezone                 string `gorm:"size:50"`
	DiallingCode             int
	LocalPhoneNumber         string `gorm:"size:10"`
	FeaturedPostId           int
	FeaturedPost             *Post `gorm:"foreignKey:FeaturedPostId"`
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
	Name string `gorm:"uniqueIndex;size:64"`
}

func main() {
	db, _ := gorm.Open(sqlite.Open("examples/sqla/admin/sample_db.sqlite"),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			Logger:         logger.Default.LogMode(logger.Info)})

	a := gadmin.NewAdmin("Example: SQLAlchemy", db)
	a.AddView(gadmin.NewModelView(User{}))
	a.AddView(gadmin.NewModelView(Tag{})).(*gadmin.ModelView).
		SetTablePrefixHtml(`<h5>dismiss table prefix</h5>`).
		SetColumnEditableList("name")
	a.AddView(gadmin.NewModelView(Post{}))

	a.BaseView.Menu.AddMenu(&gadmin.Menu{Category: "Other", Name: "Other"})
	a.BaseView.Menu.AddMenu(&gadmin.Menu{Category: "Other", Name: "Tree"})
	a.BaseView.Menu.AddMenu(&gadmin.Menu{Category: "Other", Name: "Links", Children: []*gadmin.Menu{
		{Name: "Back Home", Path: "/"},
		{Name: "External Link", Path: "http://www.example.com/"},
	}})

	// TODO: replace index handler /admin/

	a.Run()
}
