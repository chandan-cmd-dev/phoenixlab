package services

import (
	"PhoenixLab/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type AnalyticsService struct{}

type MonthlyStats struct {
	Month    string `json:"month"`
	Received int    `json:"received"`
	Resolved int    `json:"resolved"`
}

type TechnicianStats struct {
	TechnicianId   int    `json:"technician_id"`
	TechnicianName string `json:"technician_name"`
	Open           int    `json:"open"`
	Resolved       int    `json:"resolved"`
	Total          int    `json:"total"`
}

func (s *AnalyticsService) GetMonthlyStats(branchID int, months int) ([]MonthlyStats, error) {
	o := orm.NewOrm()
	var results []MonthlyStats

	from := time.Now().AddDate(0, -months, 0)

	query := `
		SELECT
			TO_CHAR(received_at, 'Mon YYYY') as month,
			COUNT(*) as received,
			COUNT(CASE WHEN status IN ('resolved','closed') THEN 1 END) as resolved
		FROM tickets
		WHERE received_at >= $1
	`
	args := []interface{}{from}
	if branchID > 0 {
		query += " AND branch_id = $2"
		args = append(args, branchID)
	}
	query += " GROUP BY TO_CHAR(received_at, 'Mon YYYY'), DATE_TRUNC('month', received_at) ORDER BY DATE_TRUNC('month', received_at)"

	_, err := o.Raw(query, args...).QueryRows(&results)
	return results, err
}

func (s *AnalyticsService) GetTechnicianStats(branchID int) ([]TechnicianStats, error) {
	o := orm.NewOrm()
	var results []TechnicianStats

	query := `
		SELECT
			u.id as technician_id,
			u.name as technician_name,
			COUNT(CASE WHEN t.status NOT IN ('resolved','closed','cancelled') THEN 1 END) as open,
			COUNT(CASE WHEN t.status IN ('resolved','closed') THEN 1 END) as resolved,
			COUNT(*) as total
		FROM users u
		LEFT JOIN tickets t ON t.assigned_to = u.id
		WHERE u.role = 'technician' AND u.is_active = true
	`
	args := []interface{}{}
	if branchID > 0 {
		query += " AND u.branch_id = $1"
		args = append(args, branchID)
	}
	query += " GROUP BY u.id, u.name ORDER BY total DESC"

	_, err := o.Raw(query, args...).QueryRows(&results)
	return results, err
}

func (s *AnalyticsService) GetStatusBreakdown(branchID int) (map[string]int, error) {
	o := orm.NewOrm()
	breakdown := make(map[string]int)

	qs := o.QueryTable("tickets")
	if branchID > 0 {
		qs = qs.Filter("BranchId", branchID)
	}

	statuses := []string{"open", "diagnosing", "parts_ordered", "in_repair", "qc_check", "resolved", "closed", "on_hold", "cancelled"}
	for _, status := range statuses {
		count, _ := qs.Filter("Status", status).Count()
		breakdown[status] = int(count)
	}
	return breakdown, nil
}

func (s *AnalyticsService) GetWarrantyBreakdown(branchID int) (map[string]int, error) {
	o := orm.NewOrm()
	breakdown := make(map[string]int)

	qs := o.QueryTable("tickets")
	if branchID > 0 {
		qs = qs.Filter("BranchId", branchID)
	}

	warranties := []string{
		string(models.WarrantyIn),
		string(models.WarrantyOut),
		string(models.WarrantyExtended),
	}
	for _, w := range warranties {
		count, _ := qs.Filter("WarrantyStatus", w).Count()
		breakdown[w] = int(count)
	}
	return breakdown, nil
}

func (s *AnalyticsService) GetAvgTAT(branchID int) (float64, error) {
	o := orm.NewOrm()

	query := `
		SELECT AVG(EXTRACT(EPOCH FROM (resolved_at - received_at)) / 3600) as avg_tat
		FROM tickets
		WHERE status IN ('resolved','closed') AND resolved_at IS NOT NULL
	`
	args := []interface{}{}
	if branchID > 0 {
		query += " AND branch_id = $1"
		args = append(args, branchID)
	}

	var result struct {
		AvgTat float64 `orm:"column(avg_tat)"`
	}
	err := o.Raw(query, args...).QueryRow(&result)
	return result.AvgTat, err
}
