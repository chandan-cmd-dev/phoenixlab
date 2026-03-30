package services

import (
	"PhoenixLab/models"
	"errors"

	"github.com/beego/beego/v2/client/orm"
)

type CommentService struct{}

func (s *CommentService) Create(ticketID, userID int, body string) (*models.Comment, error) {
	if body == "" {
		return nil, errors.New("comment body is required")
	}

	o := orm.NewOrm()

	ticket := &models.Ticket{Id: ticketID}
	if err := o.Read(ticket); err != nil {
		return nil, errors.New("ticket not found")
	}

	comment := &models.Comment{
		TicketId: ticketID,
		UserId:   userID,
		Body:     body,
	}

	_, err := o.Insert(comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentService) GetByTicket(ticketID int) ([]*models.Comment, error) {
	o := orm.NewOrm()

	var comments []*models.Comment
	_, err := o.QueryTable("comments").Filter("TicketId", ticketID).OrderBy("CreatedAt").All(&comments)
	if err != nil {
		return nil, err
	}

	for _, c := range comments {
		if c.UserId > 0 {
			user := &models.User{Id: c.UserId}
			if err := o.Read(user); err == nil {
				c.User = user
			}
		}
	}

	return comments, nil
}
