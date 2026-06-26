package services

import (
	"PhoenixLab/models"
	"regexp"
	"strings"

	"github.com/beego/beego/v2/client/orm"
)

type SheetMappingService struct{}

// FieldOption is a selectable mapping target for the mapping UI.
type FieldOption struct {
	Value     string
	Label     string
	Transform string
}

// logicalKeyToField maps the import_service logical column keys to real Ticket
// struct field names.
var logicalKeyToField = map[string]string{
	"sn":                "SerialNumber",
	"date":              "ReceivedAt",
	"upc":               "Upc",
	"ir":                "IrNumber",
	"issue":             "IssueDescription",
	"diagnostic":        "DiagnosticCode",
	"needed_part":       "NeededPart",
	"problem_desc":      "ProblemDescription",
	"status":            "Status",
	"machine_price":     "MachinePurchasePrice",
	"part_number":       "PartNumber",
	"part_price":        "PartsCost",
	"labour":            "LabourCost",
	"po_number":         "PoNumber",
	"case_number":       "CaseNumber",
	"work_order":        "WorkOrderNumber",
	"customer_info":     "CustomerName",
	"return_part":       "ReturnPart",
	"part_fixed":        "PartArrivedFixed",
	"defective_shipped": "DefectivePartShipped",
	"case_finished":     "CaseFinished",
}

var currencyFields = map[string]bool{
	"MachinePurchasePrice": true, "PartsCost": true, "LabourCost": true,
}
var boolFields = map[string]bool{
	"ReturnPart": true, "PartArrivedFixed": true, "CaseFinished": true, "CustomerRepair": true,
}
var dateFields = map[string]bool{
	"ReceivedAt": true, "ResolvedAt": true, "DueDate": true,
}

// FieldCatalog lists the ticket fields that may be chosen as mapping targets.
func (s *SheetMappingService) FieldCatalog() []FieldOption {
	fields := []FieldOption{
		{"SerialNumber", "Serial Number", "text"},
		{"IssueDescription", "Issue Description", "text"},
		{"IrNumber", "IR Number", "text"},
		{"Upc", "UPC", "text"},
		{"Model", "Model", "text"},
		{"Brand", "Brand", "text"},
		{"Status", "Status", "status"},
		{"DiagnosticCode", "Diagnostic Code", "text"},
		{"NeededPart", "Needed Part", "text"},
		{"ProblemDescription", "Problem Description", "text"},
		{"MachinePurchasePrice", "Machine Purchase Price", "currency"},
		{"PartNumber", "Part Number", "text"},
		{"PartsCost", "Parts Cost", "currency"},
		{"LabourCost", "Labour Cost", "currency"},
		{"PoNumber", "PO Number", "text"},
		{"CaseNumber", "Case Number", "text"},
		{"WorkOrderNumber", "Work Order Number", "text"},
		{"CustomerName", "Customer Name", "text"},
		{"CustomerEmail", "Customer Email", "text"},
		{"CustomerPhone", "Customer Phone", "text"},
		{"ReturnPart", "Return Part?", "bool"},
		{"PartArrivedFixed", "Part Arrived/Fixed?", "bool"},
		{"DefectivePartShipped", "Defective Part Shipped", "text"},
		{"CaseFinished", "Case Finished?", "bool"},
		{"ReceivedAt", "Received Date", "date"},
		{"DueDate", "Due Date", "date"},
		{"RmaNumber", "RMA Number", "text"},
		{"Notes", "Notes", "text"},
	}
	return fields
}

// IdentityKeyOptions are the fields offered as identity-key choices.
func (s *SheetMappingService) IdentityKeyOptions() []FieldOption {
	return []FieldOption{
		{"SerialNumber", "Serial Number", "text"},
		{"IssueDescription", "Issue Description", "text"},
		{"CaseNumber", "Case Number", "text"},
		{"WorkOrderNumber", "Work Order Number", "text"},
		{"IrNumber", "IR Number", "text"},
	}
}

func transformFor(field string) string {
	switch {
	case field == "Status":
		return "status"
	case currencyFields[field]:
		return "currency"
	case boolFields[field]:
		return "bool"
	case dateFields[field]:
		return "date"
	default:
		return "text"
	}
}

// DetectHeaderRow finds the 0-indexed header row (looks for SN/serial/date),
// falling back to row 0.
func (s *SheetMappingService) DetectHeaderRow(rows [][]string) int {
	for i := 0; i < len(rows) && i < 5; i++ {
		for _, cell := range rows[i] {
			lower := strings.ToLower(strings.TrimSpace(cell))
			if lower == "sn" || strings.Contains(lower, "serial") || lower == "date" {
				return i
			}
		}
	}
	return 0
}

// SuggestMapping auto-maps headers to ticket fields using the brand-aware
// keyword matcher from import_service. Unmatched columns become custom fields.
func (s *SheetMappingService) SuggestMapping(headers []string, brand string) []*models.SheetColumnMapping {
	if brand == "" {
		brand = "Other"
	}
	colMap := mapColumns(headers, brand)

	colToField := make(map[int]string)
	for logicalKey, idx := range colMap {
		if idx < 0 {
			continue
		}
		if field, ok := logicalKeyToField[logicalKey]; ok {

			if _, taken := colToField[idx]; !taken {
				colToField[idx] = field
			}
		}
	}

	mappings := make([]*models.SheetColumnMapping, 0, len(headers))
	for i, h := range headers {
		target := ""
		transform := "text"
		if field, ok := colToField[i]; ok {
			target = field
			transform = transformFor(field)
		} else if strings.TrimSpace(h) == "" {
			target = "ignore"
		} else {
			target = models.CustomFieldPrefix + slugify(h)
		}
		mappings = append(mappings, &models.SheetColumnMapping{
			ColumnIndex: i,
			Header:      h,
			TargetField: target,
			Transform:   transform,
		})
	}
	return mappings
}

// SaveMapping replaces all mappings for a connection.
func (s *SheetMappingService) SaveMapping(connID int, mappings []*models.SheetColumnMapping) error {
	o := orm.NewOrm()
	if _, err := o.Raw("DELETE FROM sheet_column_mappings WHERE connection_id = ?", connID).Exec(); err != nil {
		return err
	}
	for _, m := range mappings {
		m.Id = 0
		m.ConnectionId = connID
		if m.TargetField == "" {
			m.TargetField = "ignore"
		}
		if m.Transform == "" {
			m.Transform = "text"
		}
		if _, err := o.Insert(m); err != nil {
			return err
		}
	}
	return nil
}

func (s *SheetMappingService) LoadMappings(connID int) ([]*models.SheetColumnMapping, error) {
	o := orm.NewOrm()
	var mappings []*models.SheetColumnMapping
	_, err := o.QueryTable("sheet_column_mappings").
		Filter("connection_id", connID).
		OrderBy("ColumnIndex").
		All(&mappings)
	return mappings, err
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "_")
	s = strings.Trim(s, "_")
	if s == "" {
		return "field"
	}
	return s
}
