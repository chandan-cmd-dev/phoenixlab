package controllers

import (
	"PhoenixLab/models"
	"PhoenixLab/services"
	"encoding/json"
	"html/template"
	"strings"
)

func chartJSON(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		return template.JS("[]")
	}
	return template.JS(b)
}

func titleCase(s string) string {
	parts := strings.Fields(s)
	for i, p := range parts {
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

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
		recentTickets = []*models.Ticket{}
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

	analyticsService := services.AnalyticsService{}
	monthly, err := analyticsService.GetMonthlyStats(branchScope, 6)
	if err != nil {
		monthly = []services.MonthlyStats{}
	}
	monthLabels := []string{}
	monthReceived := []int{}
	monthResolved := []int{}
	for _, m := range monthly {
		monthLabels = append(monthLabels, m.Month)
		monthReceived = append(monthReceived, m.Received)
		monthResolved = append(monthResolved, m.Resolved)
	}

	statusOrder := []string{"open", "diagnosing", "parts_ordered", "part_applied", "in_repair", "qc_check", "resolved", "closed", "on_hold", "cancelled"}
	statusLabels := []string{}
	statusCounts := []int{}
	for _, st := range statusOrder {
		if stats[st] > 0 {
			statusLabels = append(statusLabels, titleCase(strings.ReplaceAll(st, "_", " ")))
			statusCounts = append(statusCounts, stats[st])
		}
	}

	c.Data["monthLabels"] = chartJSON(monthLabels)
	c.Data["monthReceived"] = chartJSON(monthReceived)
	c.Data["monthResolved"] = chartJSON(monthResolved)
	c.Data["statusLabels"] = chartJSON(statusLabels)
	c.Data["statusCounts"] = chartJSON(statusCounts)

	c.Data["kpiCards"] = kpiCards
	c.Data["recentTickets"] = recentTicketsData
	c.Data["stats"] = stats
	c.Data["title"] = "Dashboard"
	c.SetActivePage("dashboard")
	c.GetFlashMessages()
	c.TplName = "dashboard/index.html"
}
