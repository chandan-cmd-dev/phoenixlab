package services

import (
	"PhoenixLab/models"

	"github.com/beego/beego/v2/client/orm"
)

type AuditService struct{}

func (a *AuditService) Log(
	entityType string,
	entityID int,
	action string,
	field string,
	oldVal string,
	newVal string,
	userID int,
	ipAddress string,
) {
	o := orm.NewOrm()
	entry := &models.AuditLog{
		EntityType: entityType,
		EntityId:   entityID,
		Action:     action,
		FieldName:  field,
		OldValue:   oldVal,
		NewValue:   newVal,
		ChangedBy:  userID,
		IpAddress:  ipAddress,
	}

	u := &models.User{Id: userID}
	if err := o.Read(u); err == nil {
		entry.ChangedByName = u.Name
	}

	o.Insert(entry)
}

func (a *AuditService) GetForTicket(ticketID int) ([]*models.AuditLog, error) {
	o := orm.NewOrm()
	var logs []*models.AuditLog
	_, err := o.QueryTable("audit_log").
		Filter("EntityType", "ticket").
		Filter("EntityId", ticketID).
		OrderBy("-ChangedAt").
		All(&logs)
	return logs, err
}

func (a *AuditService) GetAll(branchID int, role string, page, pageSize int) ([]*models.AuditLog, int64, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("audit_log").OrderBy("-ChangedAt")

	if role != string(models.RoleSuperAdmin) {
		qs = qs.Filter("changed_by__in", getUserIDsForBranch(branchID))
	}

	total, _ := qs.Count()
	var logs []*models.AuditLog
	_, err := qs.Limit(pageSize, (page-1)*pageSize).All(&logs)
	return logs, total, err
}

func getUserIDsForBranch(branchID int) []int {
	o := orm.NewOrm()
	var users []*models.User
	o.QueryTable("users").Filter("branch_id", branchID).All(&users)

	userIDs := make([]int, len(users))
	for i, u := range users {
		userIDs[i] = u.Id
	}
	return userIDs
}
