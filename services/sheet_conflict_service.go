package services

import (
	"PhoenixLab/models"
	"encoding/json"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type SheetConflictService struct{}

func (s *SheetConflictService) ListOpen(connID int) ([]*models.SheetConflict, error) {
	o := orm.NewOrm()
	var conflicts []*models.SheetConflict
	_, err := o.QueryTable("sheet_conflicts").
		Filter("connection_id", connID).
		Filter("status", "open").
		OrderBy("Id").All(&conflicts)
	return conflicts, err
}

func (s *SheetConflictService) CountOpen(connID int) int64 {
	o := orm.NewOrm()
	c, _ := o.QueryTable("sheet_conflicts").
		Filter("connection_id", connID).
		Filter("status", "open").Count()
	return c
}

func (s *SheetConflictService) Resolve(connID int, decisions map[int]string, userID int) error {
	o := orm.NewOrm()
	conn, err := (&SheetConnectionService{}).GetByID(connID)
	if err != nil {
		return err
	}

	syncSvc := SheetSyncService{}
	audit := AuditService{}

	needSheet := false
	for _, d := range decisions {
		if d == "db" {
			needSheet = true
		}
	}

	var ctx *syncContext
	fieldToCol := map[string]int{}
	linkRow := map[int]int{}
	if needSheet {
		ctx, err = syncSvc.loadContext(conn)
		if err != nil {
			return err
		}
		for col, f := range ctx.fieldCols {
			if _, ok := fieldToCol[f]; !ok {
				fieldToCol[f] = col
			}
		}
		for col, k := range ctx.customCols {
			fieldToCol[models.CustomFieldPrefix+k] = col
		}
		for i := range ctx.dataRows {
			if l := syncSvc.resolveLink(o, ctx, ctx.dataRows[i]); l != nil {
				linkRow[l.Id] = ctx.dataStart + i
			}
		}
	}

	var updates []CellUpdate
	for cid, decision := range decisions {
		if decision != "sheet" && decision != "db" {
			continue
		}
		var c models.SheetConflict
		if err := o.QueryTable("sheet_conflicts").
			Filter("id", cid).Filter("connection_id", connID).One(&c); err != nil {
			continue
		}
		if c.Status != "open" {
			continue
		}

		link := &models.SheetRowLink{Id: c.LinkId}
		o.Read(link)
		t := &models.Ticket{Id: c.TicketId}
		if err := o.Read(t); err != nil {
			continue
		}

		baselineVal := c.BaselineValue
		if decision == "sheet" {
			if strings.HasPrefix(c.FieldName, models.CustomFieldPrefix) {
				t.SetCustomField(strings.TrimPrefix(c.FieldName, models.CustomFieldPrefix), c.SheetValue)
			} else {
				setTicketField(t, c.FieldName, c.SheetValue, transformFor(c.FieldName))
			}
			o.Update(t)
			baselineVal = c.SheetValue
			audit.Log("ticket", t.Id, "update", c.FieldName, c.DbValue, c.SheetValue, userID, "")
		} else {
			col, ok := fieldToCol[c.FieldName]
			row, hasRow := linkRow[c.LinkId]
			if ok && hasRow {
				val := c.DbValue
				if strings.HasPrefix(c.FieldName, models.CustomFieldPrefix) {
					val = t.GetCustomFields()[strings.TrimPrefix(c.FieldName, models.CustomFieldPrefix)]
				} else {
					val = formatFieldValue(t, c.FieldName)
				}
				updates = append(updates, CellUpdate{Row: row, Col: col, Value: val})
			}
			baselineVal = c.DbValue
			audit.Log("ticket", t.Id, "update", c.FieldName, c.SheetValue, c.DbValue, userID, "")
		}

		snap := decodeSnapshot(link.BaselineSnapshot)
		snap[c.FieldName] = baselineVal
		if b, err := json.Marshal(snap); err == nil {
			link.BaselineSnapshot = string(b)
			link.ContentHash = hashCanonical(snap)
			o.Update(link)
		}

		c.Status = "resolved"
		c.Resolution = decision
		c.ResolvedBy = userID
		c.ResolvedAt = time.Now()
		o.Update(&c)
	}

	if len(updates) > 0 && ctx != nil {
		if err := ctx.client.WriteCells(conn.SpreadsheetId, conn.TabName, updates); err != nil {
			return err
		}
	}
	return nil
}
