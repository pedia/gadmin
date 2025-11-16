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
	vat := gadm.NewModelView(sqla.AllTyped{}, db)
	vat.SetColumnDescriptions(map[string]string{
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
		SetFormColumns("name").
		SetColumnSearchableList("name", "email", "bdge").
		SetColumnEditableList("name", "email", "age", "is_normal", "valid", "type", "long", "badge",
			"birthday", "activated_at", "created_at", "updated_at", "decimal", "bytes", "favorite", "last_login")
	a.AddView(vat)

	a.AddView(gadm.NewModelView(sqla.Company{}, db, "BelongsTo"))
	ve := gadm.NewModelView(sqla.Employee{}, db, "BelongsTo").
		Joins("Company")
	a.AddView(ve)

	a.AddView(gadm.NewModelView(sqla.CreditCard{}, db, "HasOne"))
	vu := gadm.NewModelView(sqla.User{}, db, "HasOne").
		Joins("CreditCard")
	a.AddView(vu)

	a.AddView(gadm.NewModelView(sqla.Address{}, db, "HasMany"))
	va := gadm.NewModelView(sqla.Account{}, db, "HasMany").
		Preloads("Addresses")
	a.AddView(va)

	a.AddView(gadm.NewModelView(sqla.Toy{}, db, "Polymorphism"))
	vt := gadm.NewModelView(sqla.Dog{}, db, "Polymorphism").
		Preloads("Toys")
	a.AddView(vt)

	a.BaseView.Menu.AddMenu(&gadm.Menu{Category: "Other", Name: "Other", Path: "/other"})
	a.BaseView.Menu.AddMenu(&gadm.Menu{Name: "Tree", Path: "/tree"}, "Other")
	a.BaseView.Menu.AddMenu(&gadm.Menu{Name: "Links", Path: "/links", Children: []*gadm.Menu{
		{Name: "Back Home", Path: "/"},
		{Name: "External Link", Path: "http://www.example.com/"},
	}}, "Other")

	// TODO: replace index handler /admin/

	for _, p := range sqla.Samples {
		db.Create(p)
	}

	a.Run()
}
