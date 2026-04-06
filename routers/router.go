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

	web.Router("/", &controllers.DashboardController{}, "get:Index")
}
