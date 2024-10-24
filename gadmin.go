package gadmin

import (
	"fmt"
	"strings"
)

// Like Flask.url_for
func (*Admin) urlFor(model, endpoint string, args map[string]any) (string, error) {
	// endpoint to path
	if pos := strings.Index(endpoint, "."); pos != 0 {
		model = endpoint[:pos]
		endpoint = endpoint[pos:]
	}
	path, ok := map[string]string{
		".index_view":   "",
		".create_view":  "new",
		".details_view": "details",
		".action_view":  "action",
		".execute_view": "execute",
		".edit_view":    "edit",
		".delete_view":  "delete",
		".export":       "export",
	}[endpoint]
	if !ok {
		return "", fmt.Errorf(`endpoint "%s" not found`, endpoint)
	}

	// apply custom args
	uv := anyMapToQuery(args)
	if model != "" {
		path = "/admin/" + model + "/"
	} else {
		if path != "" {
			path = path + "/" // TODO: use URL.JoinPath
		}
	}
	if len(uv) > 0 {
		return path + "?" + uv.Encode(), nil
	}
	return path, nil
}
