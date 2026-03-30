package models

import (
	"time"
)

type OdooSyncLog struct {
	Id           int       `orm:"auto;pk" json:"id"`
	TicketId     int       `orm:"column(ticket_id)" json:"ticket_id"`
	Direction    string    `orm:"size(20)" json:"direction"`
	OdooTicketId string    `orm:"null;size(100)" json:"odoo_ticket_id"`
	Status       string    `orm:"size(20)" json:"status"`
	Payload      string    `orm:"null;type(jsonb)" json:"payload"`
	ErrorMessage string    `orm:"null;type(text)" json:"error_message"`
	SyncedAt     time.Time `orm:"auto_now_add;type(timestamptz)" json:"synced_at"`

	Ticket *Ticket `orm:"-" json:"ticket,omitempty"`
}

func (o *OdooSyncLog) TableName() string {
	return "odoo_sync_log"
}
