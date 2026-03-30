package models

import (
	"time"
)

type TicketPart struct {
	Id          int       `orm:"auto;pk" json:"id"`
	TicketId    int       `orm:"column(ticket_id)" json:"ticket_id"`
	PartNumber  string    `orm:"size(100)" json:"part_number" form:"part_number"`
	Description string    `orm:"null;type(text)" json:"description" form:"description"`
	Quantity    int       `orm:"default(1)" json:"quantity" form:"quantity"`
	UnitCost    float64   `orm:"digits(10);decimals(2)" json:"unit_cost" form:"unit_cost"`
	Status      string    `orm:"size(20);default(pending)" json:"status" form:"status"`
	OrderedAt   time.Time `orm:"null;type(timestamptz)" json:"ordered_at"`
	ReceivedAt  time.Time `orm:"null;type(timestamptz)" json:"received_at"`
	CreatedAt   time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`

	Ticket *Ticket `orm:"-" json:"ticket,omitempty"`
}

func (t *TicketPart) TableName() string {
	return "ticket_parts"
}

func (t *TicketPart) GetTotalCost() float64 {
	return float64(t.Quantity) * t.UnitCost
}
