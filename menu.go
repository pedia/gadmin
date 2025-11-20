package gadm

import (
	"log"
	"strings"

	"github.com/samber/lo"
)

// Tree liked structure
type Menu struct {
	Category string // tree liked
	Name     string
	Path     string

	Icon  string
	Class string

	IsActive     bool // TODO:
	IsVisible    bool
	IsAccessible bool

	Children []*Menu
}

// TODO: AddCategory/AddLink/AddMenuItem
func (M *Menu) AddMenu(i *Menu, category ...string) {
	if i.Category == "" {
		if i.Path == "" && !strings.HasPrefix(i.Path, "/") {
			np := "/" + strings.ToLower(i.Name)
			log.Printf(`menu(%s) path '%s' invalid, fixed to '%s'`, i.Name, i.Path, np)
			i.Path = np
		}
	}

	parent := M.find(firstOr(category, ""))
	if parent != nil {
		parent.Children = append(parent.Children, i)
	} else {
		// stub, create a new stub or self is stub
		stub := &Menu{Name: i.Category, Category: i.Category}
		stub.Children = append(stub.Children, i)
		M.Children = append(M.Children, stub)
	}
}

func (M *Menu) find(cate string) *Menu {
	if M.Category == cate {
		return M
	}

	c, _ := lo.Find(M.Children, func(m *Menu) bool {
		return m.Category == cate
	})
	return c
}
