package api

import (
	"net/http"
	"strings"

	"github.com/samber/lo"
)

// Tree liked structure
type MenuItem struct {
	Category string // parent item Name
	Name     string
	Url      string
	Icon     string
	Class    string
	Children []MenuItem
	IsActive bool
}

func (M MenuItem) dict() map[string]any {
	return map[string]any{
		"name":  M.Name,
		"icon":  M.Icon,
		"class": M.Class,
		"children": lo.Map(M.Children, func(m MenuItem, _ int) map[string]any {
			return m.dict()
		}),
	}
}

type View interface {
	// Add custom handler, eg: /admin/{model}/path
	Expose(path string, h http.HandlerFunc)

	// CreateBluePrint()

	// Generate URL for the endpoint.
	// In model view, return {model}/{action}
	GetUrl(ep string, args ...string) string

	GetBlueprint() *Blueprint
	GetMenu() *MenuItem

	// Override this method if you want dynamically hide or show
	// administrative views from Flask-Admin menu structure
	IsVisible() bool

	// Override this method to add permission checks.
	IsAccessible() bool
	Render(w http.ResponseWriter, template string, data map[string]any)
}

type BaseView struct {
	*Blueprint
	menu MenuItem
}

func NewView(menu MenuItem) *BaseView {
	return &BaseView{menu: menu}
}

// Expose "/test" create Blueprint{Endpoint: "test", Path: "/test"}
func (V *BaseView) Expose(path string, h http.HandlerFunc) {
	// TODO: better way to generate default `endpoint`
	ep := strings.ToLower(strings.ReplaceAll(path, "/", ""))

	V.Blueprint.Register(
		&Blueprint{Endpoint: ep, Path: path, Handler: h},
	)
}

func (V *BaseView) GetBlueprint() *Blueprint                                           { return V.Blueprint }
func (V *BaseView) GetMenu() *MenuItem                                                 { return &V.menu }
func (V *BaseView) IsVisible() bool                                                    { return true }
func (V *BaseView) IsAccessible() bool                                                 { return true }
func (V *BaseView) Render(w http.ResponseWriter, template string, data map[string]any) {}
