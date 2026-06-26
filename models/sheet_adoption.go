package models

import (
	"time"
)

// SheetAdoption is a sheet row whose identity matched multiple candidate
// tickets (ambiguous). It is queued for human review; RowDataJson stashes the
// parsed row so the chosen action can be applied later.
type SheetAdoption struct {
	Id             int       `orm:"auto;pk" json:"id"`
	ConnectionId   int       `orm:"column(connection_id)" json:"connection_id"`
	SheetRowUid    string    `orm:"column(sheet_row_uid);size(500)" json:"sheet_row_uid"`
	NaturalKey     string    `orm:"null;size(500)" json:"natural_key"`
	RowDataJson    string    `orm:"null;type(jsonb)" json:"row_data_json"`
	CandidateIds   string    `orm:"null;size(500)" json:"candidate_ids"`
	Status         string    `orm:"size(20);default(open)" json:"status"`
	Resolution     string    `orm:"null;size(20)" json:"resolution"`
	ResultTicketId int       `orm:"null;column(result_ticket_id)" json:"result_ticket_id"`
	ResolvedBy     int       `orm:"null;column(resolved_by)" json:"resolved_by"`
	ResolvedAt     time.Time `orm:"null;type(timestamptz)" json:"resolved_at"`
	CreatedAt      time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
	UpdatedAt      time.Time `orm:"auto_now;type(timestamptz)" json:"updated_at"`

	Candidates []*Ticket `orm:"-" json:"candidates,omitempty"`
}

func (a *SheetAdoption) TableName() string {
	return "sheet_adoptions"
}
