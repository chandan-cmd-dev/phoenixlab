package models

import (
	"time"
)

// SheetConflict records a single mapped field where, since the baseline, BOTH
// the sheet and the DB changed to different values.
type SheetConflict struct {
	Id            int       `orm:"auto;pk" json:"id"`
	ConnectionId  int       `orm:"column(connection_id)" json:"connection_id"`
	LinkId        int       `orm:"column(link_id)" json:"link_id"`
	TicketId      int       `orm:"column(ticket_id)" json:"ticket_id"`
	FieldName     string    `orm:"size(200)" json:"field_name"`
	BaselineValue string    `orm:"null;type(text)" json:"baseline_value"`
	SheetValue    string    `orm:"null;type(text)" json:"sheet_value"`
	DbValue       string    `orm:"null;type(text)" json:"db_value"`
	Status        string    `orm:"size(20);default(open)" json:"status"`
	Resolution    string    `orm:"null;size(20)" json:"resolution"`
	ResolvedBy    int       `orm:"null;column(resolved_by)" json:"resolved_by"`
	ResolvedAt    time.Time `orm:"null;type(timestamptz)" json:"resolved_at"`
	CreatedAt     time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
	UpdatedAt     time.Time `orm:"auto_now;type(timestamptz)" json:"updated_at"`
}

func (c *SheetConflict) TableName() string {
	return "sheet_conflicts"
}
