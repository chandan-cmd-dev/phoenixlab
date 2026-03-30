package services

import (
	"PhoenixLab/models"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/xuri/excelize/v2"
)

type ImportService struct{}

type ImportResult struct {
	TotalRows int
	Imported  int
	Skipped   int
	Errors    []string
	SheetName string
}

func (s *ImportService) ImportExcel(filePath string, branchID int, userID int) ([]ImportResult, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	var results []ImportResult
	for _, sheetName := range f.GetSheetList() {
		result := s.importSheet(f, sheetName, branchID, userID)
		results = append(results, result)
	}
	return results, nil
}

func (s *ImportService) importSheet(f *excelize.File, sheetName string, branchID int, userID int) ImportResult {
	result := ImportResult{SheetName: sheetName}

	rows, err := f.GetRows(sheetName)
	if err != nil || len(rows) < 2 {
		result.Errors = append(result.Errors, fmt.Sprintf("Sheet '%s': cannot read rows or empty", sheetName))
		return result
	}

	brand := detectBrand(sheetName)

	headerRow, dataStartIdx := detectHeaders(rows, brand)
	if headerRow == nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Sheet '%s': cannot detect header row", sheetName))
		return result
	}

	colMap := mapColumns(headerRow, brand)

	o := orm.NewOrm()
	audit := AuditService{}

	for i := dataStartIdx; i < len(rows); i++ {
		row := rows[i]
		result.TotalRows++

		sn := getCellValue(row, colMap["sn"])
		if sn == "" {
			result.Skipped++
			continue
		}

		ticket := &models.Ticket{
			SerialNumber:   strings.TrimSpace(sn),
			Brand:          brand,
			BranchId:       branchID,
			CreatedBy:      userID,
			WarrantyStatus: "in_warranty",
			Priority:       "normal",
			ReceivedAt:     time.Now(),
		}

		if dateStr := getCellValue(row, colMap["date"]); dateStr != "" {
			if t, err := parseFlexDate(dateStr); err == nil {
				ticket.ReceivedAt = t
			}
		}

		ticket.Upc = strings.TrimSpace(getCellValue(row, colMap["upc"]))

		if brand == "HP" {
			ticket.IrNumber = strings.TrimSpace(getCellValue(row, colMap["ir"]))
		}

		ticket.IssueDescription = strings.TrimSpace(getCellValue(row, colMap["issue"]))

		ticket.DiagnosticCode = strings.TrimSpace(getCellValue(row, colMap["diagnostic"]))

		ticket.NeededPart = strings.TrimSpace(getCellValue(row, colMap["needed_part"]))

		if desc := strings.TrimSpace(getCellValue(row, colMap["problem_desc"])); desc != "" {
			ticket.ProblemDescription = desc
		}

		caseStatus := strings.TrimSpace(getCellValue(row, colMap["status"]))
		ticket.Status = mapExcelStatus(caseStatus)

		ticket.MachinePurchasePrice = parseCurrency(getCellValue(row, colMap["machine_price"]))

		ticket.PartNumber = strings.TrimSpace(getCellValue(row, colMap["part_number"]))

		ticket.PartsCost = parseCurrency(getCellValue(row, colMap["part_price"]))

		ticket.LabourCost = parseCurrency(getCellValue(row, colMap["labour"]))

		ticket.PoNumber = strings.TrimSpace(getCellValue(row, colMap["po_number"]))

		ticket.CaseNumber = strings.TrimSpace(getCellValue(row, colMap["case_number"]))

		ticket.WorkOrderNumber = strings.TrimSpace(getCellValue(row, colMap["work_order"]))

		if custInfo := strings.TrimSpace(getCellValue(row, colMap["customer_info"])); custInfo != "" {
			ticket.CustomerName = custInfo
		}

		ticket.ReturnPart = isYes(getCellValue(row, colMap["return_part"]))

		ticket.PartArrivedFixed = isYes(getCellValue(row, colMap["part_fixed"]))

		defShipped := strings.TrimSpace(getCellValue(row, colMap["defective_shipped"]))
		if defShipped != "" {
			ticket.DefectivePartShipped = defShipped
		}

		ticket.CaseFinished = isYes(getCellValue(row, colMap["case_finished"]))

		ticket.AssignedTo = userID
		_, err := o.Insert(ticket)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Row %d (SN: %s): %v", i+1, sn, err))
			result.Skipped++
			continue
		}
		o.Raw("UPDATE tickets SET assigned_to = NULL WHERE id = ?", ticket.Id).Exec()

		audit.Log("ticket", ticket.Id, "create", "", "", "Imported from Excel: "+sheetName, userID, "")
		result.Imported++
	}

	return result
}

func detectBrand(sheetName string) string {
	lower := strings.ToLower(sheetName)
	if strings.Contains(lower, "hp") {
		return "HP"
	}
	if strings.Contains(lower, "lenovo") {
		return "Lenovo"
	}
	if strings.Contains(lower, "dell") {
		return "Dell"
	}
	return "Other"
}

func detectHeaders(rows [][]string, brand string) ([]string, int) {
	for i := 0; i < len(rows) && i < 3; i++ {
		row := rows[i]
		for _, cell := range row {
			lower := strings.ToLower(strings.TrimSpace(cell))
			if lower == "sn" || lower == "sn " || strings.Contains(lower, "serial") || lower == "date" || lower == "date " {
				return row, i + 1
			}
		}
	}
	return nil, 0
}

func mapColumns(headers []string, brand string) map[string]int {
	m := make(map[string]int)
	keys := []string{"date", "sn", "upc", "ir", "issue", "diagnostic", "needed_part",
		"problem_desc", "status", "machine_price", "part_number", "part_price",
		"labour", "po_number", "case_number", "work_order", "customer_info",
		"return_part", "part_fixed", "defective_shipped", "case_finished"}
	for _, k := range keys {
		m[k] = -1
	}

	if brand == "HP" {
		for i, h := range headers {
			lower := strings.ToLower(strings.TrimSpace(h))
			switch {
			case lower == "date" || lower == "date ":
				m["date"] = i
			case lower == "sn" || lower == "sn ":
				m["sn"] = i
			case lower == "upc":
				m["upc"] = i
			case lower == "ir":
				m["ir"] = i
			case lower == "issue":
				m["issue"] = i
			case strings.Contains(lower, "diagnostic"):
				m["diagnostic"] = i
			case strings.Contains(lower, "needed part"):
				m["needed_part"] = i
			case strings.Contains(lower, "case status") || strings.Contains(lower, "status"):
				m["status"] = i
			case strings.Contains(lower, "machine purchase") || strings.Contains(lower, "machine price"):
				m["machine_price"] = i
			case strings.Contains(lower, "part") && strings.Contains(lower, "number") && !strings.Contains(lower, "po"):
				m["part_number"] = i
			case strings.Contains(lower, "part") && (strings.Contains(lower, "retail") || strings.Contains(lower, "price")):
				m["part_price"] = i
			case strings.Contains(lower, "labor") || strings.Contains(lower, "labour") || strings.Contains(lower, "reimbursement"):
				m["labour"] = i
			case strings.Contains(lower, "po number") || strings.Contains(lower, "po"):
				m["po_number"] = i
			case strings.Contains(lower, "case number") && !strings.Contains(lower, "status"):
				m["case_number"] = i
			case strings.Contains(lower, "customer service") || strings.Contains(lower, "service order") || strings.Contains(lower, "work order"):
				m["work_order"] = i
			case strings.Contains(lower, "return") && strings.Contains(lower, "part"):
				m["return_part"] = i
			case strings.Contains(lower, "arrived") || strings.Contains(lower, "fixed"):
				m["part_fixed"] = i
			case strings.Contains(lower, "defective") || strings.Contains(lower, "shipped"):
				m["defective_shipped"] = i
			case strings.Contains(lower, "case") && strings.Contains(lower, "结束"):
				m["case_finished"] = i
			}
		}
		if m["sn"] == -1 {
			m["date"] = 0
			m["sn"] = 1
			m["upc"] = 2
			m["ir"] = 3
			m["issue"] = 4
			m["diagnostic"] = 5
			m["needed_part"] = 6
			m["status"] = 7
			m["machine_price"] = 8
			m["part_number"] = 9
			m["part_price"] = 10
			m["labour"] = 11
			m["po_number"] = 12
			m["case_number"] = 13
			m["work_order"] = 14
			m["return_part"] = 15
			m["part_fixed"] = 16
			m["defective_shipped"] = 17
			m["case_finished"] = 18
		}
	} else {
		for i, h := range headers {
			lower := strings.ToLower(strings.TrimSpace(h))
			switch {
			case lower == "date" || lower == "date ":
				m["date"] = i
			case lower == "sn" || lower == "sn ":
				m["sn"] = i
			case lower == "upc":
				m["upc"] = i
			case strings.Contains(lower, "case status") || (lower == "status"):
				m["status"] = i
			case lower == "issue":
				m["issue"] = i
			case strings.Contains(lower, "diagnostic"):
				m["diagnostic"] = i
			case strings.Contains(lower, "problem"):
				m["problem_desc"] = i
			case strings.Contains(lower, "needed part"):
				m["needed_part"] = i
			case strings.Contains(lower, "customer"):
				m["customer_info"] = i
			case strings.Contains(lower, "part") && strings.Contains(lower, "number") && !strings.Contains(lower, "po"):
				m["part_number"] = i
			case strings.Contains(lower, "machine") && strings.Contains(lower, "price"):
				m["machine_price"] = i
			case strings.Contains(lower, "part") && (strings.Contains(lower, "retail") || strings.Contains(lower, "price")):
				m["part_price"] = i
			case strings.Contains(lower, "labor") || strings.Contains(lower, "labour") || strings.Contains(lower, "reimbursement"):
				m["labour"] = i
			case strings.Contains(lower, "case number") || strings.Contains(lower, "case") && !strings.Contains(lower, "status") && !strings.Contains(lower, "结束"):
				m["case_number"] = i
			case strings.Contains(lower, "work order"):
				m["work_order"] = i
			case strings.Contains(lower, "return") || (strings.Contains(lower, "归还") && strings.Contains(lower, "part")):
				m["return_part"] = i
			case strings.Contains(lower, "arrived") || strings.Contains(lower, "fixed") || strings.Contains(lower, "送达"):
				m["part_fixed"] = i
			case strings.Contains(lower, "defective") || strings.Contains(lower, "shipped") || strings.Contains(lower, "寄出"):
				m["defective_shipped"] = i
			case strings.Contains(lower, "结束"):
				m["case_finished"] = i
			}
		}
		if m["sn"] == -1 {
			m["date"] = 0
			m["sn"] = 1
			m["upc"] = 2
			m["status"] = 3
			m["issue"] = 4
			m["diagnostic"] = 5
			m["problem_desc"] = 6
			m["needed_part"] = 7
			m["customer_info"] = 8
			m["part_number"] = 9
			m["machine_price"] = 10
			m["part_price"] = 11
			m["labour"] = 12
			m["case_number"] = 13
			m["work_order"] = 14
			m["return_part"] = 15
			m["part_fixed"] = 16
			m["defective_shipped"] = 17
			m["case_finished"] = 18
		}
	}
	return m
}

func getCellValue(row []string, colIdx int) string {
	if colIdx < 0 || colIdx >= len(row) {
		return ""
	}
	return row[colIdx]
}

func mapExcelStatus(status string) string {
	lower := strings.ToLower(strings.TrimSpace(status))
	switch {
	case lower == "hold" || lower == "on hold":
		return "on_hold"
	case strings.Contains(lower, "part applied"):
		return "part_applied"
	case strings.Contains(lower, "finished") || strings.Contains(lower, "closed"):
		return "closed"
	case lower == "":
		return "open"
	default:
		return "open"
	}
}

func parseFlexDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	formats := []string{
		"1/2/2006",
		"01/02/2006",
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"1/2/06",
		"01/02/06",
	}
	for _, fmt := range formats {
		if t, err := time.Parse(fmt, s); err == nil {
			return t, nil
		}
	}
	if num, err := strconv.ParseFloat(s, 64); err == nil && num > 40000 && num < 60000 {
		base := time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
		return base.AddDate(0, 0, int(num)), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse date: %s", s)
}

var currencyRegex = regexp.MustCompile(`[^0-9.\-]`)

func parseCurrency(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" || strings.ToLower(s) == "n/a" {
		return 0
	}
	cleaned := currencyRegex.ReplaceAllString(s, "")
	if val, err := strconv.ParseFloat(cleaned, 64); err == nil {
		return val
	}
	return 0
}

func isYes(s string) bool {
	return strings.ToLower(strings.TrimSpace(s)) == "yes"
}
