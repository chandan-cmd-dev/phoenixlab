package models

import (
	"strconv"
	"time"
)

type SheetRowLink struct {
	Id               int       `orm:"auto;pk" json:"id"`
	ConnectionId     int       `orm:"column(connection_id)" json:"connection_id"`
	SheetRowUid      string    `orm:"column(sheet_row_uid);size(500)" json:"sheet_row_uid"`
	TicketId         int       `orm:"column(ticket_id)" json:"ticket_id"`
	ContentHash      string    `orm:"null;size(100)" json:"content_hash"`
	BaselineSnapshot string    `orm:"null;type(jsonb)" json:"baseline_snapshot"`
	StampedUid       string    `orm:"null;column(stamped_uid);size(100)" json:"stamped_uid"`
	LastPushedAt     time.Time `orm:"null;type(timestamptz)" json:"last_pushed_at"`
	LastPulledAt     time.Time `orm:"null;type(timestamptz)" json:"last_pulled_at"`
	CreatedAt        time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
	UpdatedAt        time.Time `orm:"auto_now;type(timestamptz)" json:"updated_at"`
}

func (l *SheetRowLink) TableName() string {
	return "sheet_row_links"
}

func (l *SheetRowLink) StampValue() string {
	return "PLR-" + strconv.Itoa(l.Id)
}
