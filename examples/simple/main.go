package main

import "gadmin"

type User struct {
	Id                       string `gorm:"primaryKey"`
	Type                     string
	EnumChoiceField          string `a:"enum:1=first,2=second;"`
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

	mv := gadmin.NewModalView(User{}, admin.DB).
		SetColumnList([]string{"type", "first_name", "last_name", "email", "ip_address", "currency", "timezone", "phone_number"}).
		SetColumnEditableList([]string{"first_name", "type", "currency", "timezone"})
	admin.AddView(mv)
	admin.Run()
}
