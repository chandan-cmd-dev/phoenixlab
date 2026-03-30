package services

import (
	"PhoenixLab/models"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type TicketService struct{}

func (s *TicketService) Create(t *models.Ticket, userID int) error {
	t.CreatedBy = userID
	if t.BranchId == 0 {
		t.BranchId = getUserBranch(userID)
	}
	if t.ReceivedAt.IsZero() {
		t.ReceivedAt = time.Now()
	}
	if t.Status == "" {
		t.Status = string(models.StatusOpen)
	}
	o := orm.NewOrm()

	wasUnassigned := t.AssignedTo == 0
	if wasUnassigned {
		t.AssignedTo = t.CreatedBy
	}

	_, err := o.Insert(t)
	if err != nil {
		return fmt.Errorf("failed to create ticket: %w", err)
	}

	if wasUnassigned {
		o.Raw("UPDATE tickets SET assigned_to = NULL WHERE id = ?", t.Id).Exec()
		t.AssignedTo = 0
	}

	auditService := AuditService{}
	auditService.Log("ticket", t.Id, "create", "", "", t.SerialNumber, userID, "")
	return nil
}

func (s *TicketService) GetByID(id int) (*models.Ticket, error) {
	o := orm.NewOrm()
	t := &models.Ticket{Id: id}
	if err := o.Read(t); err != nil {
		return nil, err
	}

	if t.BranchId > 0 {
		branch := &models.Branch{Id: t.BranchId}
		if err := o.Read(branch); err == nil {
			t.Branch = branch
		}
	}

	if t.AssignedTo > 0 {
		user := &models.User{Id: t.AssignedTo}
		if err := o.Read(user); err == nil {
			t.Assigned = user
		}
	}

	if t.CreatedBy > 0 {
		user := &models.User{Id: t.CreatedBy}
		if err := o.Read(user); err == nil {
			t.Creator = user
		}
	}

	return t, nil
}

func (s *TicketService) Update(t *models.Ticket, changedFields []string, userID int) error {
	o := orm.NewOrm()

	existing := &models.Ticket{Id: t.Id}
	if err := o.Read(existing); err != nil {
		return errors.New("ticket not found")
	}
	if existing.Version != t.Version {
		return errors.New("conflict: this record was modified by someone else — please refresh and try again")
	}

	audit := AuditService{}
	for _, field := range changedFields {
		oldVal := getField(existing, field)
		newVal := getField(t, field)
		if oldVal != newVal {
			audit.Log("ticket", t.Id, "update", field, oldVal, newVal, userID, "")
		}
	}

	t.Version = existing.Version + 1
	_, err := o.Update(t, append(changedFields, "Version", "UpdatedAt")...)
	return err
}

func (s *TicketService) GetByBranch(branchID int, filters map[string]string) ([]*models.Ticket, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("tickets")

	if branchID > 0 {
		qs = qs.Filter("BranchId", branchID)
	}

	if v, ok := filters["status"]; ok && v != "" {
		qs = qs.Filter("Status", v)
	}
	if v, ok := filters["warranty"]; ok && v != "" {
		qs = qs.Filter("WarrantyStatus", v)
	}
	if v, ok := filters["q"]; ok && v != "" {
		qs = qs.Filter("SerialNumber__icontains", v)
	}
	if v, ok := filters["assigned"]; ok && v != "" {
		if assignedID, err := strconv.Atoi(v); err == nil {
			qs = qs.Filter("AssignedTo", assignedID)
		}
	}
	if v, ok := filters["brand"]; ok && v != "" {
		qs = qs.Filter("Brand", v)
	}

	var tickets []*models.Ticket
	_, err := qs.OrderBy("-CreatedAt").All(&tickets)
	return tickets, err
}

func (s *TicketService) UpdateStatus(ticketID int, newStatus string, userID int) error {
	o := orm.NewOrm()

	t := &models.Ticket{Id: ticketID}
	if err := o.Read(t); err != nil {
		return errors.New("ticket not found")
	}

	if !t.CanTransitionTo(newStatus) {
		return errors.New("invalid status transition")
	}

	oldStatus := t.Status
	t.Status = newStatus

	if newStatus == string(models.StatusResolved) {
		t.ResolvedAt = time.Now()
	}

	_, err := o.Update(t, "Status", "ResolvedAt", "UpdatedAt")
	if err != nil {
		return err
	}

	audit := AuditService{}
	audit.Log("ticket", ticketID, "update", "status", oldStatus, newStatus, userID, "")

	return nil
}

func (s *TicketService) Assign(ticketID int, assignedTo int, userID int) error {
	o := orm.NewOrm()

	t := &models.Ticket{Id: ticketID}
	if err := o.Read(t); err != nil {
		return errors.New("ticket not found")
	}

	oldAssigned := t.AssignedTo
	t.AssignedTo = assignedTo

	_, err := o.Update(t, "AssignedTo", "UpdatedAt")
	if err != nil {
		return err
	}

	audit := AuditService{}
	audit.Log("ticket", ticketID, "update", "assigned_to",
		strconv.Itoa(oldAssigned), strconv.Itoa(assignedTo), userID, "")

	return nil
}

func (s *TicketService) Delete(ticketID int, userID int) error {
	o := orm.NewOrm()

	t := &models.Ticket{Id: ticketID}
	if err := o.Read(t); err != nil {
		return errors.New("ticket not found")
	}

	_, err := o.Delete(t)
	if err != nil {
		return err
	}

	audit := AuditService{}
	audit.Log("ticket", ticketID, "delete", "", "", t.SerialNumber, userID, "")

	return nil
}

func (s *TicketService) GetStats(branchID int, role string) (map[string]int, error) {
	o := orm.NewOrm()
	stats := make(map[string]int)

	qs := o.QueryTable("tickets")
	if role != string(models.RoleSuperAdmin) && branchID > 0 {
		qs = qs.Filter("BranchId", branchID)
	}

	total, _ := qs.Count()
	stats["total"] = int(total)

	statuses := []string{"open", "diagnosing", "parts_ordered", "part_applied", "in_repair", "qc_check", "resolved", "closed", "on_hold", "cancelled"}
	for _, status := range statuses {
		count, _ := qs.Filter("Status", status).Count()
		stats[status] = int(count)
	}

	overdue, _ := qs.Filter("DueDate__lt", time.Now()).Exclude("Status__in", []string{"resolved", "closed", "cancelled"}).Count()
	stats["overdue"] = int(overdue)

	return stats, nil
}

func getUserBranch(userID int) int {
	o := orm.NewOrm()
	user := &models.User{Id: userID}
	if err := o.Read(user); err == nil {
		return user.BranchId
	}
	return 0
}

func getField(obj interface{}, field string) string {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return ""
	}

	f := v.FieldByName(field)
	if !f.IsValid() {
		return ""
	}

	switch f.Kind() {
	case reflect.String:
		return f.String()
	case reflect.Int, reflect.Int64:
		return strconv.FormatInt(f.Int(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(f.Float(), 'f', 2, 64)
	case reflect.Bool:
		return strconv.FormatBool(f.Bool())
	default:
		return fmt.Sprintf("%v", f.Interface())
	}
}
