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
		BaseView: gadmin.NewView(gadmin.Menu{Name: "View3", Category: "Test"}),
	}
	v.Expose("/", v.indexHandler)
	return v
}

func (M *MyView) indexHandler(w http.ResponseWriter, r *http.Request) {
	M.Render(w, r, "examples/simple/templates/myadmin.gotmpl", nil, nil)
}

func main() {
	admin := gadmin.NewAdmin("Example: Simple Views", nil)

	v := gadmin.NewView(gadmin.Menu{Name: "View1", Category: "Test"})
	v.Expose("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("raw view"))
	})
	admin.AddView(v)
	admin.AddView(gadmin.NewView(gadmin.Menu{Category: "Test", Name: "View2"}))
	admin.AddView(NewMyView())
	admin.Run()
}
