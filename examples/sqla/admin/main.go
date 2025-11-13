package main

import (
	"gadm"
	"gadm/examples/sqla"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	db, _ := gorm.Open(sqlite.Open("examples/sqla/sample.db"),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{SingularTable: true},
			Logger:         logger.Default.LogMode(logger.Info)})

	a := gadm.NewAdmin("Example: SQLAlchemy")
	vu := gadm.NewModelView(sqla.AllTyped{}, db)
	vu.SetColumnDescriptions(map[string]string{
		"is_normal": "nobody is normal",
		"type":      "3 career",
	}).
		SetTextareaRow(map[string]int{"long": 3}).
		SetFormChoices(map[string][]gadm.Choice{"type": {
			{Value: "admin", Label: "Admin"},
			{Value: "content-writer", Label: "Content writer"},
			{Value: "editor", Label: "Editor"},
			{Value: "regular-user", Label: "Regular user"}},
		}).
		SetCanSetPageSize().
		SetColumnList("name", "email", "age", "is_normal", "valid", "type", "long", "badge",
			"birthday", "activated_at", "created_at", "updated_at", "decimal", "bytes", "favorite", "last_login").
		SetColumnSearchableList("name", "email", "bdge").
		SetColumnEditableList("name", "email", "age", "is_normal", "valid", "type", "long", "badge",
			"birthday", "activated_at", "created_at", "updated_at", "decimal", "bytes", "favorite", "last_login")
	a.AddView(vu)

	vp := gadm.NewModelView(sqla.Company{}, db)
	a.AddView(vp)

	a.AddView(gadm.NewModelView(sqla.Employee{}, db))

	a.BaseView.Menu.AddMenu(&gadm.Menu{Category: "Other", Name: "Other", Path: "/other"})
	a.BaseView.Menu.AddMenu(&gadm.Menu{Category: "Other", Name: "Tree", Path: "/tree"})
	a.BaseView.Menu.AddMenu(&gadm.Menu{Category: "Other", Name: "Links", Path: "/links", Children: []*gadm.Menu{
		{Name: "Back Home", Path: "/"},
		{Name: "External Link", Path: "http://www.example.com/"},
	}})

	// TODO: replace index handler /admin/

	a.Run()
}
