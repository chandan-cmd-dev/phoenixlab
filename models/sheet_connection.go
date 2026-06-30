package models

import (
	"strings"
	"time"
)

type SheetConnection struct {
	Id                  int       `orm:"auto;pk" json:"id"`
	SpreadsheetId       string    `orm:"size(200)" json:"spreadsheet_id" form:"spreadsheet_id"`
	SpreadsheetTitle    string    `orm:"null;size(500)" json:"spreadsheet_title"`
	TabName             string    `orm:"null;size(200)" json:"tab_name" form:"tab_name"`
	Brand               string    `orm:"null;size(100)" json:"brand" form:"brand"`
	BranchId            int       `orm:"column(branch_id)" json:"branch_id" form:"branch_id"`
	Status              string    `orm:"size(20);default(draft)" json:"status"`
	SyncDirection       string    `orm:"size(20);default(two_way)" json:"sync_direction" form:"sync_direction"`
	ConflictPolicy      string    `orm:"size(20);default(review)" json:"conflict_policy" form:"conflict_policy"`
	HeaderRow           int       `orm:"default(0)" json:"header_row" form:"header_row"`
	IdentityKey         string    `orm:"size(300);default(SerialNumber,IssueDescription)" json:"identity_key" form:"identity_key"`
	AutoSyncEnabled     bool      `orm:"default(false)" json:"auto_sync_enabled" form:"auto_sync_enabled"`
	SyncIntervalMinutes int       `orm:"default(15)" json:"sync_interval_minutes" form:"sync_interval_minutes"`
	LastAutoRunAt       time.Time `orm:"null;type(timestamptz)" json:"last_auto_run_at"`
	LastAutoStatus      string    `orm:"null;size(20)" json:"last_auto_status"`
	LastAutoMessage     string    `orm:"null;type(text)" json:"last_auto_message"`
	LastSyncedAt        time.Time `orm:"null;type(timestamptz)" json:"last_synced_at"`
	LastPushedAt        time.Time `orm:"null;type(timestamptz)" json:"last_pushed_at"`
	CreatedBy           int       `orm:"column(created_by)" json:"created_by"`
	CreatedAt           time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
	UpdatedAt           time.Time `orm:"auto_now;type(timestamptz)" json:"updated_at"`

	Branch *Branch `orm:"-" json:"branch,omitempty"`
}

func (c *SheetConnection) TableName() string {
	return "sheet_connections"
}

func (c *SheetConnection) IdentityFields() []string {
	var out []string
	for _, f := range strings.Split(c.IdentityKey, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			out = append(out, f)
		}
	}
	if len(out) == 0 {
		out = []string{"SerialNumber"}
	}
	return out
}

func (c *SheetConnection) EffectiveIntervalMinutes() int {
	if c.SyncIntervalMinutes < 5 {
		return 5
	}
	return c.SyncIntervalMinutes
}
