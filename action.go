package gadmin

type action map[string]any

var view_row_action = action(map[string]any{
	"name":          "view_row_action",
	"template_name": "row_actions.view_row",
	"title":         gettext("View Record"),
})

// {'title': 'Edit Record', 'template_name': 'row_actions.edit_row'}
var edit_row_action = action(map[string]any{
	"name":          "edit_row_action",
	"title":         gettext("Edit Record"),
	"template_name": "row_actions.edit_row",
})

// {'title': 'Delete Record', 'template_name': 'row_actions.delete_row'}
var delete_row_action = action(map[string]any{
	"name":          "delete_row_action",
	"title":         gettext("Delete Record"),
	"csrf_token":    "",
	"id":            "", // HiddenField(validators=[InputRequired()]).Render(value)
	"url":           "",
	"template_name": "row_actions.delete_row",
})

// title
// icon_class
// endpoint
// id_arg
// url_args
