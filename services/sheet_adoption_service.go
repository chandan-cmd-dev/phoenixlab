package services

import (
	"PhoenixLab/models"
	"encoding/json"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type SheetAdoptionService struct{}

func (s *SheetAdoptionService) ListOpenReviews(connID int) ([]*models.SheetAdoption, error) {
	o := orm.NewOrm()
	var adoptions []*models.SheetAdoption
	_, err := o.QueryTable("sheet_adoptions").
		Filter("connection_id", connID).
		Filter("status", "open").
		OrderBy("Id").All(&adoptions)
	if err != nil {
		return nil, err
	}
	for _, a := range adoptions {
		for _, id := range splitInts(a.CandidateIds) {
			t := &models.Ticket{Id: id}
			if o.Read(t) == nil {
				a.Candidates = append(a.Candidates, t)
			}
		}
	}
	return adoptions, nil
}

func (s *SheetAdoptionService) CountOpen(connID int) int64 {
	o := orm.NewOrm()
	c, _ := o.QueryTable("sheet_adoptions").
		Filter("connection_id", connID).
		Filter("status", "open").Count()
	return c
}

func (s *SheetAdoptionService) Resolve(connID int, decisions map[int]string, userID int) error {
	o := orm.NewOrm()
	conn, err := (&SheetConnectionService{}).GetByID(connID)
	if err != nil {
		return err
	}
	audit := AuditService{}

	for aid, decision := range decisions {
		var a models.SheetAdoption
		if err := o.QueryTable("sheet_adoptions").
			Filter("id", aid).Filter("connection_id", connID).One(&a); err != nil {
			continue
		}
		if a.Status != "open" {
			continue
		}
		data := decodeSnapshot(a.RowDataJson)

		switch {
		case decision == "dismiss":
			a.Status = "dismissed"
			a.Resolution = "dismiss"

		case decision == "new":
			t := newTicketFromData(conn, data, userID)
			if err := ensureTicketOwnership(o, t); err != nil {
				continue
			}
			if _, err := o.Insert(t); err != nil {
				continue
			}
			audit.Log("ticket", t.Id, "create", "", "", "Created via adoption review", userID, "")
			s.upsertLink(o, conn.Id, a.SheetRowUid, t, data)
			a.Status = "resolved"
			a.Resolution = "new"
			a.ResultTicketId = t.Id

		default:
			ticketID, convErr := strconv.Atoi(decision)
			if convErr != nil || !containsInt(splitInts(a.CandidateIds), ticketID) {
				continue
			}
			t := &models.Ticket{Id: ticketID}
			if o.Read(t) != nil {
				continue
			}
			applyStashedData(t, data)
			o.Update(t)
			audit.Log("ticket", t.Id, "update", "", "", "Adopted via adoption review", userID, "")
			s.upsertLink(o, conn.Id, a.SheetRowUid, t, data)
			a.Status = "resolved"
			a.Resolution = "adopt"
			a.ResultTicketId = t.Id
		}

		a.ResolvedBy = userID
		a.ResolvedAt = time.Now()
		o.Update(&a)
	}
	return nil
}

func newTicketFromData(conn *models.SheetConnection, data map[string]string, userID int) *models.Ticket {
	t := &models.Ticket{
		BranchId:       conn.BranchId,
		Brand:          conn.Brand,
		CreatedBy:      userID,
		WarrantyStatus: "in_warranty",
		Priority:       "normal",
		Status:         "open",
		ReceivedAt:     time.Now(),
	}
	applyStashedData(t, data)
	return t
}

func (s *SheetAdoptionService) upsertLink(o orm.Ormer, connID int, uid string, t *models.Ticket, data map[string]string) {
	snap := snapshotFromData(t, data)
	snapJSON, _ := json.Marshal(snap)

	var link models.SheetRowLink
	if err := o.QueryTable("sheet_row_links").
		Filter("connection_id", connID).
		Filter("sheet_row_uid", uid).One(&link); err == nil {
		link.TicketId = t.Id
		link.BaselineSnapshot = string(snapJSON)
		link.ContentHash = hashCanonical(snap)
		link.LastPulledAt = time.Now()
		o.Update(&link)
		return
	}
	o.Insert(&models.SheetRowLink{
		ConnectionId:     connID,
		SheetRowUid:      uid,
		TicketId:         t.Id,
		BaselineSnapshot: string(snapJSON),
		ContentHash:      hashCanonical(snap),
		LastPulledAt:     time.Now(),
	})
}

func snapshotFromData(t *models.Ticket, data map[string]string) map[string]string {
	snap := map[string]string{}
	custom := t.GetCustomFields()
	for k := range data {
		if len(k) > len(models.CustomFieldPrefix) && k[:len(models.CustomFieldPrefix)] == models.CustomFieldPrefix {
			snap[k] = canonicalValue(custom[k[len(models.CustomFieldPrefix):]], "text")
		} else {
			snap[k] = fieldCanonical(t, k)
		}
	}
	return snap
}

func containsInt(haystack []int, needle int) bool {
	for _, n := range haystack {
		if n == needle {
			return true
		}
	}
	return false
}
