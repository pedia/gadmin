package gadmin

import (
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

type View interface {
	// Add custom handler, eg: /admin/{model}/path
	Expose(path string, h http.HandlerFunc)

	// CreateBluePrint()

	// Generate URL for the endpoint.
	// In model view, return {model}/{action}
	GetUrl(ep string, q *Query, args ...any) string

	GetBlueprint() *Blueprint
	GetMenu() *Menu

	// Override this method if you want dynamically hide or show
	// administrative views from Flask-Admin menu structure
	IsVisible() bool

	// Override this method to add permission checks.
	IsAccessible() bool
	Render(http.ResponseWriter, *http.Request, string, template.FuncMap, map[string]any)
}

type BaseView struct {
	*Blueprint
	menu  Menu
	admin *Admin
}

func NewView(menu Menu) *BaseView {
	return &BaseView{Blueprint: &Blueprint{}, menu: menu}
}

// Expose "/test" create Blueprint{Endpoint: "test", Path: "/test"}
func (V *BaseView) Expose(path string, h http.HandlerFunc) {
	// TODO: better way to generate default `endpoint`
	ep := strings.ToLower(strings.ReplaceAll(path, "/", ""))

	V.Blueprint.Add(
		&Blueprint{Endpoint: ep, Path: path, Handler: h},
	)
}

// TODO: move query into `ModelView`
func (V *BaseView) GetUrl(ep string, q *Query, args ...any) string {
	var uv url.Values
	if q != nil {
		uv = q.withArgs(args...).toValues()
	} else {
		uv = pairToQuery(args...)
	}

	if strings.HasPrefix(ep, ".") {
		ep = V.Endpoint + ep
	}
	if V.admin != nil {
		return must(V.admin.GetUrl(ep, uv))
	}
	return must(V.Blueprint.GetUrl(ep, uv))
}

func (V *BaseView) GetBlueprint() *Blueprint { return V.Blueprint }
func (V *BaseView) GetMenu() *Menu           { return &V.menu }
func (V *BaseView) IsVisible() bool          { return true }
func (V *BaseView) IsAccessible() bool       { return true }

func (V *BaseView) Render(w http.ResponseWriter, r *http.Request, name string, funcs template.FuncMap, data map[string]any) {
	w.Header().Add("content-type", ContentTypeUtf8Html)
	fs := []string{
		"templates/actions.gotmpl",
		"templates/base.gotmpl",
		"templates/layout.gotmpl",
		"templates/lib.gotmpl",
		"templates/master.gotmpl",
	}

	fm := V.admin.funcs(funcs)

	fm["get_flashed_messages"] = func() []map[string]any {
		return GetFlashedMessages(r)
	}

	if err := createTemplate(fs, fm).
		ExecuteTemplate(w, name, V.dict(r, data)); err != nil {
		panic(err)
	}
}

func (V *BaseView) dict(r *http.Request, others ...map[string]any) map[string]any {
	// TODO: remove r
	o := map[string]any{
		"category":           V.menu.Category,
		"name":               V.menu.Name,
		"extra_css":          []string{},
		"extra_js":           []string{}, // "a.js", "b.js"}
		"admin":              V.admin.dict(),
		"admin_fluid_layout": true,
		"csrf_token":         NewCSRF(CurrentSession(r)).GenerateToken,
	}

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func createTemplate(fs []string, funcs template.FuncMap) *template.Template {
	return template.Must(template.New("all").
		Option("missingkey=error").
		Funcs(funcs).
		ParseFiles(fs...))
}
