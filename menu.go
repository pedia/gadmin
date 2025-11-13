package gadm

import (
	"log"
	"strings"

	"github.com/samber/lo"
)

// Tree liked structure
type Menu struct {
	Category string // parent item Name
	Name     string // label
	Path     string // linked to url

	Icon  string
	Class string

	IsActive     bool // TODO:
	IsVisible    bool
	IsAccessible bool

	Children []*Menu
}

func (M *Menu) dict() map[string]any {
	return map[string]any{
		"category":      M.Category,
		"name":          M.Name,
		"path":          M.Path,
		"icon":          M.Icon,
		"class":         M.Class,
		"is_active":     M.IsActive,
		"is_visible":    M.IsVisible,
		"is_accessible": M.IsAccessible,
		"children": lo.Map(M.Children, func(c *Menu, _ int) map[string]any {
			return c.dict()
		}),
	}
}

// TODO: AddCategory/AddLink/AddMenuItem
func (M *Menu) AddMenu(i *Menu) {
	if i.Path == "" || !strings.HasPrefix(i.Path, "/") {
		np := "/" + strings.ToLower(i.Name)
		log.Printf(`menu(%s) path '%s' invalid, fixed to '%s'`, i.Name, i.Path, np)
		i.Path = np
	}

	parent := M.find(i.Category)
	if parent != nil {
		parent.Children = append(parent.Children, i)
	} else {
		// self is stub
		M.Children = append(M.Children, i)
	}
}

func (M *Menu) find(cate string) *Menu {
	c, _ := lo.Find(M.Children, func(m *Menu) bool {
		return m.Category == cate
	})
	return c
}
