package controllers

import (
	"PhoenixLab/services"
	"encoding/json"
)

type AnalyticsController struct {
	BaseController
}

func (c *AnalyticsController) Index() {
	c.RequireRole("admin", "super_admin")

	branchScope := c.GetBranchScope()
	analyticsService := services.AnalyticsService{}

	monthly, _ := analyticsService.GetMonthlyStats(branchScope, 6)

	techStats, _ := analyticsService.GetTechnicianStats(branchScope)

	statusBreakdown, _ := analyticsService.GetStatusBreakdown(branchScope)

	warrantyBreakdown, _ := analyticsService.GetWarrantyBreakdown(branchScope)

	avgTAT, _ := analyticsService.GetAvgTAT(branchScope)

	monthlyLabels := []string{}
	monthlyReceived := []int{}
	monthlyResolved := []int{}
	for _, m := range monthly {
		monthlyLabels = append(monthlyLabels, m.Month)
		monthlyReceived = append(monthlyReceived, m.Received)
		monthlyResolved = append(monthlyResolved, m.Resolved)
	}
	monthlyJSON, _ := json.Marshal(map[string]interface{}{
		"labels":   monthlyLabels,
		"received": monthlyReceived,
		"resolved": monthlyResolved,
	})

	statusLabels := []string{"Open", "Diagnosing", "Parts Ordered", "In Repair", "QC Check", "Resolved", "Closed", "On Hold", "Cancelled"}
	statusKeys := []string{"open", "diagnosing", "parts_ordered", "in_repair", "qc_check", "resolved", "closed", "on_hold", "cancelled"}
	statusValues := []int{}
	for _, k := range statusKeys {
		statusValues = append(statusValues, statusBreakdown[k])
	}
	statusJSON, _ := json.Marshal(map[string]interface{}{
		"labels": statusLabels,
		"values": statusValues,
	})

	c.Data["monthly"] = monthly
	c.Data["techStats"] = techStats
	c.Data["statusBreakdown"] = statusBreakdown
	c.Data["warrantyBreakdown"] = warrantyBreakdown
	c.Data["avgTAT"] = avgTAT
	c.Data["monthlyJSON"] = string(monthlyJSON)
	c.Data["statusJSON"] = string(statusJSON)
	c.Data["title"] = "Analytics"
	c.SetActivePage("analytics")
	c.GetFlashMessages()
	c.TplName = "analytics/index.html"
}
