package gadmin

import "github.com/samber/lo"

// Tree liked structure
type Menu struct {
	Category string // parent item Name
	Name     string
	Path     string

	Icon  string
	Class string

	IsActive     bool
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
func (M *Menu) Add(m *Menu) {
	c := M.find(m.Category)
	if c == nil {
		// create stub
		c = &Menu{
			Category: m.Category,
			Name:     m.Category,
			Children: []*Menu{},
		}
	}

	c.Children = append(c.Children, m)
}

func (M *Menu) find(cate string) *Menu {
	c, _ := lo.Find(M.Children, func(m *Menu) bool {
		return m.Category == cate
	})
	return c
}
