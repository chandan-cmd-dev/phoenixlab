package controllers

import (
	"PhoenixLab/models"
	"PhoenixLab/services"
)

type DashboardController struct {
	BaseController
}

func (c *DashboardController) Index() {
	c.RequireRole("technician", "admin", "super_admin", "viewer")

	user := c.GetCurrentUser()
	branchScope := c.GetBranchScope()

	ticketService := services.TicketService{}
	stats, err := ticketService.GetStats(branchScope, user.Role)
	if err != nil {
		c.FlashError("Failed to load statistics: " + err.Error())
		c.Redirect("/auth/login", 302)
		return
	}

	filters := map[string]string{"limit": "10"}
	recentTickets, err := ticketService.GetByBranch(branchScope, filters)
	if err != nil {
		recentTickets = []*models.Ticket{} // Empty slice if error
	}

	kpiCards := []map[string]interface{}{
		{
			"Label":  "Open Tickets",
			"Value":  stats["open"],
			"Up":     false,
			"Change": "0",
		},
		{
			"Label":  "In Repair",
			"Value":  stats["in_repair"],
			"Up":     false,
			"Change": "0",
		},
		{
			"Label":  "Resolved This Week",
			"Value":  stats["resolved"],
			"Up":     false,
			"Change": "0",
		},
		{
			"Label":  "Overdue",
			"Value":  stats["overdue"],
			"Up":     false,
			"Change": "0",
		},
	}

	var recentTicketsData []map[string]interface{}
	for _, ticket := range recentTickets {
		ticketData := map[string]interface{}{
			"Id":               ticket.Id,
			"SerialNumber":     ticket.SerialNumber,
			"Brand":            ticket.Brand,
			"Model":            ticket.Model,
			"IssueDescription": ticket.IssueDescription,
			"Status":           ticket.Status,
			"AssignedToName":   "",
			"TATHours":         ticket.GetTATHours(),
		}

		if ticket.Assigned != nil {
			ticketData["AssignedToName"] = ticket.Assigned.Name
		}

		recentTicketsData = append(recentTicketsData, ticketData)
	}

	c.Data["kpiCards"] = kpiCards
	c.Data["recentTickets"] = recentTicketsData
	c.Data["stats"] = stats
	c.Data["title"] = "Dashboard"
	c.SetActivePage("dashboard")
	c.GetFlashMessages()
	c.TplName = "dashboard/index.html"
}
