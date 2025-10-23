package gadmin

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlueprint(t *testing.T) {
	is := assert.New(t)

	f := Blueprint{
		Name:     "Foo",
		Endpoint: "foo",
		Path:     "/foo",
		Children: map[string]*Blueprint{
			// flask-admin use `admin.index`, in `gadmin` should not use `view.index``
			"index":        {Endpoint: "index", Path: "/"},
			"index_view":   {Endpoint: "index_view", Path: "/"},
			"create_view":  {Endpoint: "create_view", Path: "/new"},
			"details_view": {Endpoint: "details_view", Path: "/details"},
			"action_view":  {Endpoint: "action_view", Path: "/action"},
			"execute_view": {Endpoint: "execute_view", Path: "/execute"},
			"edit_view":    {Endpoint: "edit_view", Path: "/edit"},
			"delete_view":  {Endpoint: "delete_view", Path: "/delete"},
			// not .export_view
			"export": {Endpoint: "export", Path: "/export"},
		},
	}

	// is.PanicsWithError("endpoint 'not' miss in `foo`", func() { f.GetUrl(".not") })
	is.NotPanics(func() { f.GetUrl(".not") })

	is.Equal("/foo/", must(f.GetUrl(".index")))
	is.Equal("/foo/", must(f.GetUrl("foo.index")))
	is.Equal("/foo/edit", must(f.GetUrl("foo.edit_view")))

	is.Equal("/foo/", must(f.GetUrl(".index", url.Values{})))
	is.Equal("/foo/?a=A", must(f.GetUrl(".index", url.Values{"a": []string{"A"}})))
	is.Equal("/foo/?a=B&a=C", must(f.GetUrl(".index", url.Values{"a": []string{"B", "C"}})))

	a := Blueprint{
		Name:     "Admin",
		Endpoint: "admin",
		Path:     "/admin",
		Children: map[string]*Blueprint{
			"foo":   &f,
			"index": {Endpoint: "index", Path: "/"},
		},
	}
	is.Equal("/admin/", must(a.GetUrl(".index")))
	is.Equal("/admin/foo/", must(a.GetUrl("foo.index")))
	is.Equal("/admin/foo/edit", must(a.GetUrl("foo.edit_view")))

	f.Add(&Blueprint{Endpoint: "bar", Path: "/haha"})
	is.Equal("/admin/foo/haha", must(a.GetUrl("foo.bar")))
}
