package models

import (
	"encoding/json"
	"time"
)

type WarrantyStatus string
type TicketStatus string

const (
	WarrantyIn       WarrantyStatus = "in_warranty"
	WarrantyOut      WarrantyStatus = "out_of_warranty"
	WarrantyExtended WarrantyStatus = "extended_warranty"

	StatusOpen         TicketStatus = "open"
	StatusDiagnosing   TicketStatus = "diagnosing"
	StatusPartsOrdered TicketStatus = "parts_ordered"
	StatusPartApplied  TicketStatus = "part_applied"
	StatusInRepair     TicketStatus = "in_repair"
	StatusQC           TicketStatus = "qc_check"
	StatusResolved     TicketStatus = "resolved"
	StatusClosed       TicketStatus = "closed"
	StatusOnHold       TicketStatus = "on_hold"
	StatusCancelled    TicketStatus = "cancelled"
)

type Ticket struct {
	Id                   int       `orm:"auto;pk" json:"id"`
	SerialNumber         string    `orm:"size(100)" json:"serial_number" form:"serial_number" validate:"required"`
	IrNumber             string    `orm:"null;size(100)" json:"ir_number" form:"ir_number"`
	Upc                  string    `orm:"null;size(100)" json:"upc" form:"upc"`
	Model                string    `orm:"null;size(100)" json:"model" form:"model"`
	Brand                string    `orm:"null;size(100)" json:"brand" form:"brand"`
	BranchId             int       `orm:"column(branch_id)" json:"branch_id" form:"branch_id"`
	WarrantyStatus       string    `orm:"size(30)" json:"warranty_status" form:"warranty_status" validate:"required"`
	IssueDescription     string    `orm:"null;type(text)" json:"issue_description" form:"issue_description"`
	IssueCategory        string    `orm:"null;size(30)" json:"issue_category" form:"issue_category"`
	AssignedTo           int       `orm:"null;column(assigned_to)" json:"assigned_to" form:"assigned_to"`
	Priority             string    `orm:"size(20);default(normal)" json:"priority" form:"priority"`
	CustomerRepair       bool      `orm:"default(false)" json:"customer_repair" form:"customer_repair"`
	Status               string    `orm:"size(30);default(open)" json:"status" form:"status"`
	ParentTicketId       int       `orm:"null;column(parent_ticket_id)" json:"parent_ticket_id" form:"parent_ticket_id"`
	LinkedTicketId       int       `orm:"null;column(linked_ticket_id)" json:"linked_ticket_id" form:"linked_ticket_id"`
	DiagnosticCode       string    `orm:"null;size(200)" json:"diagnostic_code" form:"diagnostic_code"`
	NeededPart           string    `orm:"null;size(200)" json:"needed_part" form:"needed_part"`
	ProblemDescription   string    `orm:"null;type(text)" json:"problem_description" form:"problem_description"`
	MachinePurchasePrice float64   `orm:"digits(10);decimals(2)" json:"machine_purchase_price" form:"machine_purchase_price"`
	PartNumber           string    `orm:"null;size(200)" json:"part_number" form:"part_number"`
	PartsCost            float64   `orm:"digits(10);decimals(2)" json:"parts_cost" form:"parts_cost"`
	LabourCost           float64   `orm:"digits(10);decimals(2)" json:"labour_cost" form:"labour_cost"`
	PoNumber             string    `orm:"null;size(100)" json:"po_number" form:"po_number"`
	CaseNumber           string    `orm:"null;size(100)" json:"case_number" form:"case_number"`
	WorkOrderNumber      string    `orm:"null;size(100)" json:"work_order_number" form:"work_order_number"`
	CourierName          string    `orm:"null;size(100)" json:"courier_name" form:"courier_name"`
	CourierTracking      string    `orm:"null;size(200)" json:"courier_tracking" form:"courier_tracking"`
	TrackingLink         string    `orm:"null;size(500)" json:"tracking_link" form:"tracking_link"`
	ReturnPart           bool      `orm:"default(false)" json:"return_part" form:"return_part"`
	ReturnTracking       string    `orm:"null;size(200)" json:"return_tracking" form:"return_tracking"`
	PartArrivedFixed     bool      `orm:"default(false)" json:"part_arrived_fixed" form:"part_arrived_fixed"`
	DefectivePartShipped string    `orm:"null;size(20)" json:"defective_part_shipped" form:"defective_part_shipped"`
	CaseFinished         bool      `orm:"default(false)" json:"case_finished" form:"case_finished"`
	CustomerName         string    `orm:"null;size(150)" json:"customer_name" form:"customer_name"`
	CustomerEmail        string    `orm:"null;size(200)" json:"customer_email" form:"customer_email"`
	CustomerPhone        string    `orm:"null;size(50)" json:"customer_phone" form:"customer_phone"`
	RmaNumber            string    `orm:"null;size(100)" json:"rma_number" form:"rma_number"`
	RmaRefNumber         string    `orm:"null;size(100)" json:"rma_ref_number" form:"rma_ref_number"`
	RmaStatus            string    `orm:"null;size(30)" json:"rma_status" form:"rma_status"`
	OdooTicketId         string    `orm:"null;size(100)" json:"odoo_ticket_id"`
	OdooSyncedAt         time.Time `orm:"null;type(timestamptz)" json:"odoo_synced_at"`
	ReceivedAt           time.Time `orm:"type(timestamptz)" json:"received_at"`
	ResolvedAt           time.Time `orm:"null;type(timestamptz)" json:"resolved_at"`
	DueDate              time.Time `orm:"null;type(date)" json:"due_date" form:"due_date"`
	CreatedBy            int       `orm:"column(created_by)" json:"created_by"`
	Version              int       `orm:"default(1)" json:"version" form:"version"`
	Notes                string    `orm:"null;type(text)" json:"notes" form:"notes"`
	CustomFields         string    `orm:"null;type(jsonb)" json:"custom_fields"`
	CreatedAt            time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
	UpdatedAt            time.Time `orm:"auto_now;type(timestamptz)" json:"updated_at"`

	Branch       *Branch `orm:"-" json:"branch,omitempty"`
	Assigned     *User   `orm:"-" json:"assigned_user,omitempty"`
	Creator      *User   `orm:"-" json:"creator,omitempty"`
	ParentTicket *Ticket `orm:"-" json:"parent_ticket,omitempty"`
	LinkedTicket *Ticket `orm:"-" json:"linked_ticket,omitempty"`
}

func (t *Ticket) TableName() string {
	return "tickets"
}

func (t *Ticket) IsOverdue() bool {
	if t.DueDate.IsZero() || t.Status == string(StatusClosed) || t.Status == string(StatusResolved) || t.Status == string(StatusCancelled) {
		return false
	}
	return time.Now().After(t.DueDate)
}

func (t *Ticket) GetTATHours() float64 {
	if t.ResolvedAt.IsZero() {
		return 0
	}
	return t.ResolvedAt.Sub(t.ReceivedAt).Hours()
}

func (t *Ticket) GetAllowedTransitions() []string {
	transitions := map[string][]string{
		string(StatusOpen):         {string(StatusDiagnosing), string(StatusPartsOrdered), string(StatusInRepair), string(StatusQC), string(StatusResolved), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusDiagnosing):   {string(StatusPartsOrdered), string(StatusPartApplied), string(StatusInRepair), string(StatusQC), string(StatusResolved), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusPartsOrdered): {string(StatusPartApplied), string(StatusInRepair), string(StatusQC), string(StatusResolved), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusPartApplied):  {string(StatusInRepair), string(StatusQC), string(StatusResolved), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusInRepair):     {string(StatusQC), string(StatusResolved), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusQC):           {string(StatusResolved), string(StatusInRepair), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusResolved):     {string(StatusClosed), string(StatusOpen)},
		string(StatusOnHold):       {string(StatusOpen), string(StatusDiagnosing), string(StatusPartsOrdered), string(StatusPartApplied), string(StatusInRepair), string(StatusQC), string(StatusResolved), string(StatusClosed)},
		string(StatusCancelled):    {string(StatusOpen)},
		string(StatusClosed):       {string(StatusOpen)},
	}
	if allowed, ok := transitions[t.Status]; ok {
		return allowed
	}
	return nil
}

// GetCustomFields decodes the JSON CustomFields bag into a map. It never
// returns nil.
func (t *Ticket) GetCustomFields() map[string]string {
	m := make(map[string]string)
	if t.CustomFields == "" {
		return m
	}
	_ = json.Unmarshal([]byte(t.CustomFields), &m)
	if m == nil {
		m = make(map[string]string)
	}
	return m
}

// SetCustomField merges a single key into the CustomFields bag.
func (t *Ticket) SetCustomField(key, value string) {
	m := t.GetCustomFields()
	m[key] = value
	if b, err := json.Marshal(m); err == nil {
		t.CustomFields = string(b)
	}
}

func (t *Ticket) CanTransitionTo(newStatus string) bool {
	current := t.Status

	validTransitions := map[string][]string{
		string(StatusOpen):         {string(StatusDiagnosing), string(StatusPartsOrdered), string(StatusInRepair), string(StatusQC), string(StatusResolved), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusDiagnosing):   {string(StatusPartsOrdered), string(StatusPartApplied), string(StatusInRepair), string(StatusQC), string(StatusResolved), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusPartsOrdered): {string(StatusPartApplied), string(StatusInRepair), string(StatusQC), string(StatusResolved), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusPartApplied):  {string(StatusInRepair), string(StatusQC), string(StatusResolved), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusInRepair):     {string(StatusQC), string(StatusResolved), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusQC):           {string(StatusResolved), string(StatusInRepair), string(StatusClosed), string(StatusOnHold), string(StatusCancelled)},
		string(StatusResolved):     {string(StatusClosed), string(StatusOpen)},
		string(StatusOnHold):       {string(StatusOpen), string(StatusDiagnosing), string(StatusPartsOrdered), string(StatusPartApplied), string(StatusInRepair), string(StatusQC), string(StatusResolved), string(StatusClosed)},
		string(StatusCancelled):    {string(StatusOpen)},
		string(StatusClosed):       {string(StatusOpen)},
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == newStatus {
			return true
		}
	}

	return false
}
