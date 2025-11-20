package gadm

import (
	"encoding/gob"
	"html/template"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
)

type View interface {
	// Add custom handler, eg: /admin/{endpoint}/{path}
	Expose(path string, h http.HandlerFunc)

	GetBlueprint() *Blueprint
	GetMenu() *Menu

	// Override this method if you want dynamically hide or show
	// administrative views from Flask-Admin menu structure
	IsVisible() bool

	// Override this method to add permission checks.
	IsAccessible() bool

	Render(http.ResponseWriter, *http.Request, string, template.FuncMap, map[string]any)

	setAdmin(*Admin)
}

type BaseView struct {
	Blueprint *Blueprint
	Menu      Menu
	admin     *Admin

	gt *groupTempl
}

func NewView(menu Menu) *BaseView {
	return &BaseView{Blueprint: &Blueprint{Path: menu.Path},
		Menu: menu,
		gt: NewGroupTempl("templates/actions.gotmpl",
			"templates/base.gotmpl",
			"templates/layout.gotmpl",
			"templates/lib.gotmpl", // move to ModelView
			"templates/master.gotmpl"),
	}
}

// Expose "/test" create Blueprint{Endpoint: "test", Path: "/test"}
// Expose "/test" create Blueprint{Endpoint: "test", Path: "/test"}
func (V *BaseView) Expose(path string, h http.HandlerFunc) {
	// generate default `endpoint`
	ep := strings.ToLower(strings.ReplaceAll(path, "/", ""))

	V.Blueprint.AddChild(
		&Blueprint{Endpoint: ep, Path: path, Handler: h},
	)
}

// func (V *BaseView) GetUrl(ep string, q *Query, args ...any) string {
// 	var uv url.Values
// 	if q != nil {
// 		uv = q.withArgs(args...).toValues()
// 	} else {
// 		uv = pairsToQuery(args...)
// 	}
// 	_ = uv

// 	if strings.HasPrefix(ep, ".") {
// 		ep = V.Blueprint.Endpoint + ep
// 	}
// 	if V.admin != nil {
// 		return must(V.admin.GetUrl(ep, q, args...))
// 	}
// 	return must(V.Blueprint.GetUrl(ep, queryToPairs(uv)...))
// }

func (V *BaseView) GetBlueprint() *Blueprint { return V.Blueprint }
func (V *BaseView) GetMenu() *Menu           { return &V.Menu }
func (V *BaseView) IsVisible() bool          { return true }
func (V *BaseView) IsAccessible() bool       { return true }

func (V *BaseView) Render(w http.ResponseWriter, r *http.Request, fn string, funcs template.FuncMap, data map[string]any) {
	fm := V.admin.funcs(funcs)
	fm["get_flashed_messages"] = func() []any {
		return V.admin.Session(r).Flashes()
	}
	fm["pager_url"] = func() string { return "TODO" }
	fm["csrf_token"] = func() string { return csrf.Token(r) }

	if err := V.gt.Render(w, fn, fm, V.dict(r, data)); err != nil {
		panic(err)
	}
}

func (V *BaseView) setAdmin(admin *Admin) { V.admin = admin }

// category: success, danger, error, info
func (V *BaseView) AddFlash(r *http.Request, flash flash) {
	sess := V.admin.Session(r)
	sess.AddFlash(flash)
}

func (V *BaseView) dict(r *http.Request, others ...map[string]any) map[string]any {
	// TODO: remove r
	o := map[string]any{
		"path":               r.URL.Path,
		"category":           V.Menu.Category,
		"name":               V.Menu.Name,
		"extra_css":          []string{},
		"extra_js":           []string{}, // "a.js", "b.js"}
		"admin":              V.admin.dict(),
		"admin_fluid_layout": true,
		"csrf_token":         func() string { return csrf.Token(r) },
	}

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func parseTemplate(name string, funcs template.FuncMap, fn ...string) *template.Template {
	return template.Must(template.New(name).
		Option("missingkey=error").
		Funcs(funcs).
		ParseFiles(fn...))
}

func init() {
	gob.Register(flash{})
}

type flash struct {
	Data     string
	Category string
}

// flash category: success(green), info(blue), danger(red)
func Flash(data, category string) flash { return flash{data, category} }
func FlashSuccess(data string) flash    { return flash{data, "success"} }
func FlashInfo(data string) flash       { return flash{data, "info"} }
func FlashError(err error) flash        { return flash{err.Error(), "danger"} }
func FlashDanger(data string) flash     { return flash{data, "danger"} }

type Action struct {
	Name         string
	Title        string
	Desc         string
	Confirmation string
	URL          string
	ReturnURL    string
	CSRFToken    string
}
