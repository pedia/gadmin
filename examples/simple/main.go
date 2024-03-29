package main

import "gadmin"

type User struct {
	Id                       string `gorm:"primaryKey"`
	Type                     string
	EnumChoiceField          string
	SqlaUtilsChoiceField     string
	SqlaUtilsEnumChoiceField int
	FirstName                string
	LastName                 string
	Email                    string
	Website                  string
	IpAddress                string
	Currency                 string
	Timezone                 string
	DiallingCode             int
	LocalPhoneNumber         string
	FeaturedPostId           int
}

func main() {
	admin := gadmin.NewAdmin("Example: Simple Views")
	admin.AddView(gadmin.NewView("Test", "view1"))
	admin.AddView(gadmin.NewView("Test", "view2"))
	admin.AddView(gadmin.NewModalView(User{}, admin.DB))
	admin.Run()
}
