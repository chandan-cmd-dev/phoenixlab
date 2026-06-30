package services

import (
	"PhoenixLab/models"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type NotificationService struct{}

func (s *NotificationService) NotifySuperAdmins(actor *models.User, ticketID int, action, message string) {
	if actor == nil {
		return
	}

	o := orm.NewOrm()
	var recipients []*models.User
	if _, err := o.QueryTable("users").
		Filter("Role", string(models.RoleSuperAdmin)).
		Filter("IsActive", true).
		All(&recipients); err != nil {
		return
	}

	for _, r := range recipients {
		if r.Id == actor.Id {
			continue
		}
		n := &models.Notification{
			RecipientId: r.Id,
			ActorId:     actor.Id,
			ActorName:   actor.Name,
			TicketId:    ticketID,
			Action:      action,
			Message:     message,
		}
		o.Insert(n)
	}
}

func (s *NotificationService) RecentForUser(userID, limit int) ([]*models.Notification, error) {
	o := orm.NewOrm()
	var notifications []*models.Notification
	_, err := o.QueryTable("notifications").
		Filter("RecipientId", userID).
		OrderBy("-CreatedAt").
		Limit(limit).
		All(&notifications)
	return notifications, err
}

func (s *NotificationService) CountUnread(userID int) (int64, error) {
	o := orm.NewOrm()
	return o.QueryTable("notifications").
		Filter("RecipientId", userID).
		Filter("IsRead", false).
		Count()
}

func (s *NotificationService) ListForUser(userID, page, pageSize int) ([]*models.Notification, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("notifications").
		Filter("RecipientId", userID).
		OrderBy("-CreatedAt")

	total, _ := qs.Count()
	var notifications []*models.Notification
	_, err := qs.Limit(pageSize, (page-1)*pageSize).All(&notifications)
	return notifications, total, err
}

func (s *NotificationService) MarkRead(id, userID int) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("notifications").
		Filter("Id", id).
		Filter("RecipientId", userID).
		Update(orm.Params{"IsRead": true, "ReadAt": time.Now()})
	return err
}

func (s *NotificationService) MarkAllRead(userID int) error {
	o := orm.NewOrm()
	_, err := o.QueryTable("notifications").
		Filter("RecipientId", userID).
		Filter("IsRead", false).
		Update(orm.Params{"IsRead": true, "ReadAt": time.Now()})
	return err
}
