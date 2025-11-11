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
	FirstName                string `gorm:"size:100"`
	LastName                 string `gorm:"size:100"`
	Email                    string `gorm:"size:255;not null"`
	Valid                    bool   `gorm:"not null"`
	BornDate                 *time.Time
	Website                  string `gorm:"default:a.io"`
	Bio                      string
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

type PostTags struct {
	PostId int `gorm:"column:post_id"`
	TagId  int `gorm:"column:tag_id"`
}

func main() {
	db, _ := gorm.Open(sqlite.Open("examples/sqla/admin/sample_db.sqlite"),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			Logger:         logger.Default.LogMode(logger.Info)})

	a := gadmin.NewAdmin("Example: SQLAlchemy", db)
	vu := gadmin.NewModelView(User{})
	vu.SetColumnDescriptions(map[string]string{"valid": "user passed verified"}).
		SetTextareaRow(map[string]int{"bio": 3}).
		SetFormChoices(map[string][]gadmin.Choice{"type": {
			{Value: "admin", Label: "Admin"},
			{Value: "content-writer", Label: "Content writer"},
			{Value: "editor", Label: "Editor"},
			{Value: "regular-user", Label: "Regular user"}},
		}).
		SetColumnEditableList("first_name", "type", "currency", "dialling_code", "valid", "born_date")
	a.AddView(vu)

	a.AddView(gadmin.NewModelView(Tag{})).(*gadmin.ModelView).
		SetTablePrefixHtml(`<h5>dismissible prefix, Tag is important</h5>`).
		SetColumnEditableList("name")

	vp := gadmin.NewModelView(Post{})
	vp.SetTextareaRow(map[string]int{"text": 5})
	a.AddView(vp)

	a.AddView(gadmin.NewModelView(PostTags{}))

	a.BaseView.Menu.AddMenu(&gadmin.Menu{Category: "Other", Name: "Other", Path: "/other"})
	a.BaseView.Menu.AddMenu(&gadmin.Menu{Category: "Other", Name: "Tree", Path: "/tree"})
	a.BaseView.Menu.AddMenu(&gadmin.Menu{Category: "Other", Name: "Links", Path: "/links", Children: []*gadmin.Menu{
		{Name: "Back Home", Path: "/"},
		{Name: "External Link", Path: "http://www.example.com/"},
	}})

	// TODO: replace index handler /admin/

	a.Run()
}
