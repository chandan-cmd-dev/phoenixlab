package models

import (
	"time"
)

type Comment struct {
	Id        int       `orm:"auto;pk" json:"id"`
	TicketId  int       `orm:"column(ticket_id)" json:"ticket_id"`
	UserId    int       `orm:"column(user_id)" json:"user_id"`
	Body      string    `orm:"type(text)" json:"body"`
	CreatedAt time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`

	User *User `orm:"-" json:"user,omitempty"`
}

func (c *Comment) TableName() string {
	return "comments"
}
