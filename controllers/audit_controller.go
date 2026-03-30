package controllers

import (
	"PhoenixLab/services"
	"strconv"
)

type AuditController struct {
	BaseController
}

func (c *AuditController) Log() {
	c.RequireRole("admin", "super_admin")

	page, _ := c.GetInt("page", 1)
	pageSize := 50

	filters := map[string]string{}
	if v := c.GetString("entity"); v != "" {
		filters["entity"] = v
	}
	if v := c.GetString("action"); v != "" {
		filters["action"] = v
	}
	if v := c.GetString("user"); v != "" {
		filters["user"] = v
	}

	auditService := services.AuditService{}
	logs, total, err := auditService.GetAll(c.GetBranchScope(), c.GetCurrentUser().Role, page, pageSize)
	if err != nil {
		c.FlashError("Failed to load audit log: " + err.Error())
		c.Redirect("/dashboard", 302)
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	c.Data["logs"] = logs
	c.Data["total"] = total
	c.Data["page"] = page
	c.Data["totalPages"] = totalPages
	c.Data["limit"] = page * pageSize
	c.Data["offset"] = (page-1)*pageSize + 1
	c.Data["filters"] = filters
	c.Data["title"] = "Audit Log"
	c.SetActivePage("audit")
	c.GetFlashMessages()
	c.TplName = "audit/log.html"
}

func (c *AuditController) ForTicket() {
	c.RequireRole("technician", "admin", "super_admin")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.RenderError(400, "Invalid ticket ID")
		return
	}

	auditService := services.AuditService{}
	logs, err := auditService.GetForTicket(id)
	if err != nil {
		c.RenderError(500, "Failed to load audit log")
		return
	}

	c.Data["json"] = map[string]interface{}{
		"logs":  logs,
		"total": strconv.Itoa(len(logs)),
	}
	c.ServeJSON()
}
