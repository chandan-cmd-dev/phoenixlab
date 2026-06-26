package routers

import (
	"PhoenixLab/controllers"

	"github.com/beego/beego/v2/server/web"
)

func init() {
	web.BConfig.WebConfig.Session.SessionOn = true

	web.Router("/auth/login", &controllers.AuthController{}, "get:ShowLogin;post:Login")
	web.Router("/auth/logout", &controllers.AuthController{}, "get:Logout")
	web.Router("/auth/profile", &controllers.AuthController{}, "get:Profile;post:UpdateProfile")

	web.Router("/dashboard", &controllers.DashboardController{}, "get:Index")

	web.Router("/tickets", &controllers.TicketController{}, "get:List;post:Create")
	web.Router("/tickets/new", &controllers.TicketController{}, "get:New")
	web.Router("/tickets/bulk-delete", &controllers.TicketController{}, "post:BulkDelete")
	web.Router("/tickets/:id", &controllers.TicketController{}, "get:Detail;post:Update")
	web.Router("/tickets/:id/edit", &controllers.TicketController{}, "get:Edit")
	web.Router("/tickets/:id/delete", &controllers.TicketController{}, "post:Delete")
	web.Router("/tickets/:id/status", &controllers.TicketController{}, "post:UpdateStatus")
	web.Router("/tickets/:id/assign", &controllers.TicketController{}, "post:Assign")
	web.Router("/tickets/:id/workflow/start", &controllers.TicketController{}, "post:StartWorkflow")
	web.Router("/tickets/:id/workflow/advance", &controllers.TicketController{}, "post:AdvanceWorkflow")

	web.Router("/users", &controllers.UserController{}, "get:List;post:Create")
	web.Router("/users/new", &controllers.UserController{}, "get:New")
	web.Router("/users/:id/edit", &controllers.UserController{}, "get:Edit")
	web.Router("/users/:id/update", &controllers.UserController{}, "post:Update")
	web.Router("/users/:id/toggle", &controllers.UserController{}, "post:ToggleActive")

	web.Router("/audit", &controllers.AuditController{}, "get:Log")
	web.Router("/audit/ticket/:id", &controllers.AuditController{}, "get:ForTicket")

	web.Router("/tickets/:id/comments", &controllers.CommentController{}, "post:Create")

	web.Router("/import", &controllers.ImportController{}, "get:ShowImport;post:DoImport")

	web.Router("/analytics", &controllers.AnalyticsController{}, "get:Index")

	web.Router("/oauth/google/connect", &controllers.OAuthController{}, "get:GoogleConnect")
	web.Router("/oauth/google/callback", &controllers.OAuthController{}, "get:GoogleCallback")
	web.Router("/oauth/google/disconnect", &controllers.OAuthController{}, "post:GoogleDisconnect")

	web.Router("/sheets", &controllers.SheetController{}, "get:List")
	web.Router("/sheets/connect", &controllers.SheetController{}, "get:Connect;post:DoConnect")
	web.Router("/sheets/:id/mapping", &controllers.SheetController{}, "get:Mapping;post:SaveMapping")
	web.Router("/sheets/:id/preview", &controllers.SheetController{}, "get:Preview")
	web.Router("/sheets/:id/import", &controllers.SheetController{}, "post:DoImport")
	web.Router("/sheets/:id/push", &controllers.SheetController{}, "get:PushPreview;post:DoPush")
	web.Router("/sheets/:id/reconcile", &controllers.SheetController{}, "post:DoReconcile")
	web.Router("/sheets/:id/conflicts", &controllers.SheetController{}, "get:Conflicts;post:ResolveConflicts")
	web.Router("/sheets/:id/adoptions", &controllers.SheetController{}, "get:Adoptions;post:ResolveAdoptions")
	web.Router("/sheets/:id/autosync", &controllers.SheetController{}, "post:ToggleAutoSync")
	web.Router("/sheets/:id/delete", &controllers.SheetController{}, "post:Delete")

	web.Router("/", &controllers.DashboardController{}, "get:Index")
}
