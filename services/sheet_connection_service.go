package services

import (
	"PhoenixLab/models"
	"fmt"

	"github.com/beego/beego/v2/client/orm"
)

type SheetConnectionService struct{}

func (s *SheetConnectionService) Create(spreadsheetInput string, branchID, userID int) (*models.SheetConnection, error) {
	id := ExtractSpreadsheetID(spreadsheetInput)
	if id == "" {
		return nil, fmt.Errorf("invalid spreadsheet URL or ID")
	}

	client, err := NewSheetsClient()
	if err != nil {
		return nil, err
	}
	meta, err := client.GetMeta(id)
	if err != nil {
		return nil, fmt.Errorf("could not open spreadsheet (check sharing/permissions): %w", err)
	}

	conn := &models.SheetConnection{
		SpreadsheetId:       id,
		SpreadsheetTitle:    meta.Title,
		BranchId:            branchID,
		Status:              "draft",
		SyncDirection:       "two_way",
		ConflictPolicy:      "review",
		HeaderRow:           0,
		IdentityKey:         "SerialNumber,IssueDescription",
		SyncIntervalMinutes: 15,
		CreatedBy:           userID,
	}
	if len(meta.Tabs) > 0 {
		conn.TabName = meta.Tabs[0]
		conn.Brand = detectBrand(meta.Tabs[0])
	}

	o := orm.NewOrm()
	if _, err := o.Insert(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

func (s *SheetConnectionService) GetByID(id int) (*models.SheetConnection, error) {
	o := orm.NewOrm()
	conn := &models.SheetConnection{Id: id}
	if err := o.Read(conn); err != nil {
		return nil, err
	}
	if conn.BranchId > 0 {
		b := &models.Branch{Id: conn.BranchId}
		if err := o.Read(b); err == nil {
			conn.Branch = b
		}
	}
	return conn, nil
}

func (s *SheetConnectionService) List(branchScope int) ([]*models.SheetConnection, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("sheet_connections").OrderBy("-UpdatedAt")
	if branchScope > 0 {
		qs = qs.Filter("branch_id", branchScope)
	}
	var conns []*models.SheetConnection
	if _, err := qs.All(&conns); err != nil {
		return nil, err
	}
	branches, _ := models.GetAllBranches()
	byID := make(map[int]*models.Branch)
	for _, b := range branches {
		byID[b.Id] = b
	}
	for _, c := range conns {
		c.Branch = byID[c.BranchId]
	}
	return conns, nil
}

func (s *SheetConnectionService) Save(conn *models.SheetConnection) error {
	o := orm.NewOrm()
	_, err := o.Update(conn)
	return err
}

func (s *SheetConnectionService) Delete(id int) error {
	o := orm.NewOrm()
	_, err := o.Raw("DELETE FROM sheet_connections WHERE id = ?", id).Exec()
	return err
}
