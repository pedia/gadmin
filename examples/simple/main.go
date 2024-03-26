package main

import "gadmin"

func main() {
	admin := gadmin.NewAdmin("Example: Simple Views")
	admin.AddView(gadmin.NewView("view1", "Test"))
	admin.AddView(gadmin.NewView("view2", "Test"))
	admin.Run()
}
