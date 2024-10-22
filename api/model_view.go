package api

import (
	"net/http"
	"reflect"
)

type ModelView struct {
	*BaseView
	model *model
}

func NewModelView(m any, category ...string) *ModelView {
	model := new_model(m)

	cate := reflect.ValueOf(m).Type().Name()
	if len(category) > 0 {
		cate = category[0]
	}

	mv := ModelView{
		BaseView: NewView(MenuItem{
			Category: cate,
			Name:     model.name(),
		}),
		model: model}

	mv.Blueprint = &Blueprint{
		Name:     model.label(),
		Endpoint: model.name(),
		Path:     model.name(),
		Children: map[string]*Blueprint{
			// In flask-admin use `view.index`. Should use `view.index_view` in `gadmin`
			"index":        {Endpoint: "index", Path: "/", Handler: mv.index},
			"index_view":   {Endpoint: "index_view", Path: "/", Handler: mv.index},
			"create_view":  {Endpoint: "create_view", Path: "/new", Handler: mv.index},
			"details_view": {Endpoint: "details_view", Path: "/details", Handler: mv.index},
			"action_view":  {Endpoint: "action_view", Path: "/action", Handler: mv.index},
			"execute_view": {Endpoint: "execute_view", Path: "/execute", Handler: mv.index},
			"edit_view":    {Endpoint: "edit_view", Path: "/edit", Handler: mv.index},
			"delete_view":  {Endpoint: "delete_view", Path: "/delete", Handler: mv.index},
			// not .export_view
			"export": {Endpoint: "export", Path: "/export", Handler: mv.index},
		},
	}

	return &mv
}

func (M *ModelView) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "text/html; charset=utf-8")
	M.Render(w, "index.gotmpl", nil)
}
