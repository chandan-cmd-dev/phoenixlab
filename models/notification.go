package models

import (
	"time"
)

type Notification struct {
	Id          int       `orm:"auto;pk" json:"id"`
	RecipientId int       `orm:"column(recipient_id)" json:"recipient_id"`
	ActorId     int       `orm:"column(actor_id)" json:"actor_id"`
	ActorName   string    `orm:"null;size(100)" json:"actor_name"`
	TicketId    int       `orm:"column(ticket_id)" json:"ticket_id"`
	Action      string    `orm:"size(50)" json:"action"`
	Message     string    `orm:"type(text)" json:"message"`
	IsRead      bool      `orm:"default(false)" json:"is_read"`
	CreatedAt   time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
	ReadAt      time.Time `orm:"null;type(timestamptz)" json:"read_at"`
}

func (n *Notification) TableName() string {
	return "notifications"
}
