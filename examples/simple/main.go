package main

import (
	"gadmin"
	"net/http"
)

type MyView struct {
	*gadmin.BaseView
}

func NewMyView() *MyView {
	v := &MyView{
		BaseView: gadmin.NewView(gadmin.Menu{Name: "view1", Category: "Test"}),
	}
	v.Expose("/", v.indexHandler)
	return v
}

func (M *MyView) indexHandler(w http.ResponseWriter, r *http.Request) {
	M.Render(w, r, "templates/myadmin.gotml", nil, nil)
}

func main() {
	admin := gadmin.NewAdmin("Example: Simple Views", nil)
	// admin.AddView(NewMyView())
	// admin.AddView(gadmin.NewView(gadmin.MenuItem{Category: "Test", Name: "View2"}))
	admin.Run()
}
