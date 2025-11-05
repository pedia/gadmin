package gadmin

import (
	"strings"

	"github.com/samber/lo"
)

// Tree liked structure
type Menu struct {
	Category string // parent item Name, TODO: remove, use Name
	Name     string // label
	Path     string // url linked to, default to /{name}

	Icon  string
	Class string

	IsActive     bool
	IsVisible    bool
	IsAccessible bool

	Children []*Menu
}

func (M *Menu) EnsureValid() {
	if M.Path == "" {
		panic("invalid path")
	}
	if !strings.HasPrefix(M.Path, "/") {
		panic("invalid path")
	}
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
func (M *Menu) AddMenu(m *Menu) {
	if m.Path == "" {
		m.Path = "/" + strings.ToLower(m.Name)
	}

	m.EnsureValid()

	parent := M.find(m.Category)
	if parent == nil {
		// create stub
		parent = &Menu{
			Category: m.Category,
			Name:     m.Category,
		}
		M.Children = append(M.Children, parent)
	}

	parent.Children = append(parent.Children, m)
}

func (M *Menu) find(cate string) *Menu {
	c, _ := lo.Find(M.Children, func(m *Menu) bool {
		return m.Category == cate
	})
	return c
}
