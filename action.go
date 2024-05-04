package gadmin

type action map[string]any

func view_row_action() action {
	return action(map[string]any{
		"name":          "view_row_action",
		"template_name": "row_actions.view_row",
		"title":         gettext("View Record"),
	})
}
func edit_row_action() action {
	return action(map[string]any{
		"name":  "edit_row_action",
		"title": gettext("Edit Record"),
	})
}
func delete_row_action() action {
	return action(map[string]any{
		"name":       "delete_row_action",
		"title":      gettext("Delete Record"),
		"csrf_token": "",
		"id":         "", // HiddenField(validators=[InputRequired()]).Render(value)
		"url":        "",
	})
}

// title
// icon_class
// endpoint
// id_arg
// url_args
