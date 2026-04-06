package models

import (
	"time"
)

type TicketWorkflow struct {
	Id           int       `orm:"auto;pk" json:"id"`
	TicketId     int       `orm:"column(ticket_id)" json:"ticket_id"`
	WorkflowType string    `orm:"size(50)" json:"workflow_type"`
	CurrentStep  string    `orm:"size(50)" json:"current_step"`
	StepData     string    `orm:"type(jsonb);default({})" json:"step_data"`
	StartedAt    time.Time `orm:"type(timestamptz)" json:"started_at"`
	CompletedAt  time.Time `orm:"null;type(timestamptz)" json:"completed_at"`
	StartedBy    int       `orm:"column(started_by)" json:"started_by"`
	CreatedAt    time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
	UpdatedAt    time.Time `orm:"auto_now;type(timestamptz)" json:"updated_at"`

	Ticket  *Ticket `orm:"-" json:"ticket,omitempty"`
	Starter *User   `orm:"-" json:"starter,omitempty"`
}

func (w *TicketWorkflow) TableName() string {
	return "ticket_workflows"
}

var WorkflowDefinitions = map[string][]string{
	"manufacturer_rma": {
		"rma_initiated",
		"box_requested",
		"box_received",
		"shipped_to_manufacturer",
		"manufacturer_processing",
		"returned_from_manufacturer",
		"workflow_complete",
	},
}

var WorkflowStepLabels = map[string]string{
	"rma_initiated":              "RMA Initiated",
	"box_requested":              "Box Requested",
	"box_received":               "Box Received",
	"shipped_to_manufacturer":    "Shipped to Manufacturer",
	"manufacturer_processing":    "Manufacturer Processing",
	"returned_from_manufacturer": "Returned from Manufacturer",
	"workflow_complete":          "Workflow Complete",
}

var WorkflowTypeLabels = map[string]string{
	"manufacturer_rma": "Manufacturer RMA",
}

func (w *TicketWorkflow) GetSteps() []string {
	if steps, ok := WorkflowDefinitions[w.WorkflowType]; ok {
		return steps
	}
	return nil
}

func (w *TicketWorkflow) CanAdvance() bool {
	steps := w.GetSteps()
	if steps == nil {
		return false
	}
	for i, s := range steps {
		if s == w.CurrentStep {
			return i < len(steps)-1
		}
	}
	return false
}

func (w *TicketWorkflow) NextStep() string {
	steps := w.GetSteps()
	if steps == nil {
		return ""
	}
	for i, s := range steps {
		if s == w.CurrentStep && i < len(steps)-1 {
			return steps[i+1]
		}
	}
	return ""
}

func (w *TicketWorkflow) IsComplete() bool {
	return !w.CompletedAt.IsZero()
}

func (w *TicketWorkflow) StepIndex() int {
	steps := w.GetSteps()
	for i, s := range steps {
		if s == w.CurrentStep {
			return i
		}
	}
	return -1
}
