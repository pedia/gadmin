package gadm

//
// html:
// <a class="icon" href="/admin/familymember/edit/?id=1%2C.%2CKen&amp;url=%2Fadmin%2Ffamilymember%2F" title="Edit Record">
//   <span class="fa fa-pencil glyphicon glyphicon-pencil"></span>
// </a>
// <form class="icon" method="POST" action="/admin/familymember/delete/">
//   <input id="id" name="id" required type="hidden" value="1,.,Ken">
//   <input id="url" name="url" type="hidden" value="/admin/familymember/">

//   <button onclick="return faHelpers.safeConfirm('Are you sure you want to delete this record?');" title="Delete record">
//     <span class="fa fa-trash glyphicon glyphicon-trash"></span>
//   </button>
// </form>

type Action map[string]any

var view_row_action = Action(map[string]any{
	"name":          "view",
	"title":         gettext("View Record"),
	"template_name": "row_actions.view_row",
})

var edit_row_action = Action(map[string]any{
	"name":          "edit",
	"title":         gettext("Edit Record"),
	"template_name": "row_actions.edit_row",
})
