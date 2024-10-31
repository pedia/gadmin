package gadmin

import "github.com/samber/lo"

// Tree liked structure
type MenuItem struct {
	Category string // parent item Name
	Name     string
	Path     string

	Icon  string
	Class string

	IsActive     bool
	IsVisible    bool
	IsAccessible bool

	Children []*MenuItem
}

func (M *MenuItem) dict() map[string]any {
	return map[string]any{
		"category":      M.Category,
		"name":          M.Name,
		"path":          M.Path,
		"icon":          M.Icon,
		"class":         M.Class,
		"is_active":     M.IsActive,
		"is_visible":    M.IsVisible,
		"is_accessible": M.IsAccessible,
		"children": lo.Map(M.Children, func(m *MenuItem, _ int) map[string]any {
			return m.dict()
		}),
	}
}

type Menu []*MenuItem

func (M *Menu) Add(m *MenuItem) {
	if m.Category != m.Name && m.Category != "" {
		c := M.findByCategory(m.Category)
		// create stub
		if c == nil {
			c = &MenuItem{
				Category: m.Category,
				Name:     m.Category,
				Children: []*MenuItem{},
			}
			*M = append(*M, c)
		}

		c.Children = append(c.Children, m)
		return
	}
	*M = append(*M, m)
}

func (M Menu) findByCategory(cate string) *MenuItem {
	c, _ := lo.Find(M, func(m *MenuItem) bool {
		return m.Category == cate
	})
	return c
}

func (M Menu) dict() []map[string]any {
	return lo.Map(M, func(m *MenuItem, _ int) map[string]any {
		return m.dict()
	})
}
