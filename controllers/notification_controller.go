package controllers

import (
	"PhoenixLab/services"
	"strconv"
)

type NotificationController struct {
	BaseController
}

func (c *NotificationController) List() {
	c.RequireRole("super_admin")

	page, _ := c.GetInt("page", 1)
	if page < 1 {
		page = 1
	}
	pageSize := 30

	notificationService := services.NotificationService{}
	notifications, total, err := notificationService.ListForUser(c.GetCurrentUser().Id, page, pageSize)
	if err != nil {
		c.FlashError("Failed to load notifications: " + err.Error())
		c.Redirect("/dashboard", 302)
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	c.Data["notifications"] = notifications
	c.Data["total"] = total
	c.Data["page"] = page
	c.Data["totalPages"] = totalPages
	c.Data["title"] = "Notifications"
	c.SetActivePage("notifications")
	c.GetFlashMessages()
	c.TplName = "notifications/list.html"
}

func (c *NotificationController) Read() {
	c.RequireRole("super_admin")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.Redirect("/notifications", 302)
		return
	}

	notificationService := services.NotificationService{}
	notificationService.MarkRead(id, c.GetCurrentUser().Id)

	if ticketID := c.GetString("ticket"); ticketID != "" {
		if _, err := strconv.Atoi(ticketID); err == nil {
			c.Redirect("/tickets/"+ticketID, 302)
			return
		}
	}
	c.Redirect("/notifications", 302)
}

func (c *NotificationController) ReadAll() {
	c.RequireRole("super_admin")

	notificationService := services.NotificationService{}
	notificationService.MarkAllRead(c.GetCurrentUser().Id)
	c.Redirect("/notifications", 302)
}
