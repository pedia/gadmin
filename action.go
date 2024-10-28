package gadmin

type action map[string]any

var view_row_action = action(map[string]any{
	"name":          "view",
	"title":         gettext("View Record"),
	"template_name": "row_actions.view_row",
})

var edit_row_action = action(map[string]any{
	"name":          "edit",
	"title":         gettext("Edit Record"),
	"template_name": "row_actions.edit_row",
})

var delete_row_action = action(map[string]any{
	"name":          "delete",
	"title":         gettext("Delete Record"),
	"template_name": "row_actions.delete_row",
	"confirmation":  gettext("Are you sure you want to delete selected records?"),
	"csrf_token":    "",
	"id":            "", // HiddenField(validators=[InputRequired()]).Render(value)
	"url":           "",
})
