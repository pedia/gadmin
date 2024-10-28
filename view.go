package gadmin

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/Masterminds/sprig/v3"
)

type View interface {
	// Add custom handler, eg: /admin/{model}/path
	Expose(path string, h http.HandlerFunc)

	// CreateBluePrint()

	// Generate URL for the endpoint.
	// In model view, return {model}/{action}
	GetUrl(ep string, q *Query, args ...any) string

	GetBlueprint() *Blueprint
	GetMenu() *MenuItem

	// Override this method if you want dynamically hide or show
	// administrative views from Flask-Admin menu structure
	IsVisible() bool

	// Override this method to add permission checks.
	IsAccessible() bool
	Render(w http.ResponseWriter, template string, data map[string]any)

	dict(others ...map[string]any) map[string]any
}

type BaseView struct {
	*Blueprint
	menu  MenuItem
	admin *Admin
}

func NewView(menu MenuItem) *BaseView {
	return &BaseView{menu: menu}
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
		return must[string](V.admin.GetUrl(ep, uv))
	}
	return must[string](V.Blueprint.GetUrl(ep, uv))
}

func (V *BaseView) GetBlueprint() *Blueprint { return V.Blueprint }
func (V *BaseView) GetMenu() *MenuItem       { return &V.menu }
func (V *BaseView) IsVisible() bool          { return true }
func (V *BaseView) IsAccessible() bool       { return true }
func (V *BaseView) Render(w http.ResponseWriter, template string, data map[string]any) {
	w.Header().Add("content-type", ContentTypeUtf8Html)
	bases := []string{
		"templates/layout.gotmpl",
		"templates/master.gotmpl",
		"templates/base.gotmpl",
		"templates/lib.gotmpl",
		"templates/model_layout.gotmpl",
		"templates/actions.gotmpl",
		"templates/model_row_actions.gotmpl",
	}
	bases = append(bases, "templates/"+template)
	if err := V.createTemplate(bases...).Lookup(template).Execute(w, data); err != nil {
		panic(err)
	}
}

func (V *BaseView) dict(others ...map[string]any) map[string]any {
	o := map[string]any{
		"category":           V.menu.Category,
		"name":               V.menu.Name,
		"extra_css":          []string{},
		"extra_js":           []string{}, // "a.js", "b.js"}
		"admin":              V.admin.dict(),
		"admin_fluid_layout": true,
	}

	if len(others) > 0 {
		merge(o, others[0])
	}
	return o
}

func templateFuncs(model string, admin *Admin) template.FuncMap {
	fm := merge(sprig.FuncMap(), Funcs)
	merge(fm, template.FuncMap{
		"admin_static_url": admin.staticURL, // used
		"get_url": func(endpoint string, args ...any) (string, error) {
			return admin.UrlFor(model, endpoint, args...)
		},
		"marshal":    admin.marshal, // test
		"config":     admin.config,  // used
		"gettext":    admin.gettext, //
		"csrf_token": func() string { return "xxxx-csrf-token" },
		// escape safe
		"safehtml": func(s string) template.HTML { return template.HTML(s) },
		"comment": func(format string, args ...any) template.HTML {
			return template.HTML(
				"<!-- " + fmt.Sprintf(format, args...) + " -->",
			)
		},
		"safejs": func(s string) template.JS { return template.JS(s) },
		"json": func(v any) (template.JS, error) {
			bs, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return template.JS(string(bs)), nil
		},
	})
	return fm
}

func (V *BaseView) createTemplate(fs ...string) *template.Template {
	tx := template.Must(template.New("all").
		Option("missingkey=error").
		Funcs(templateFuncs(V.Endpoint, V.admin)).
		ParseFiles(fs...))
	// log.Println(tx.DefinedTemplates())
	return tx
}
