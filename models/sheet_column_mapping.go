package models

import (
	"strings"
	"time"
)

// CustomFieldPrefix marks a mapping target that should be stored in the ticket's
// JSON CustomFields bag rather than a real column.
const CustomFieldPrefix = "custom:"

// SheetColumnMapping maps one sheet column (by 0-indexed position) to a ticket
// field. TargetField is either a real ticket field name, "custom:<slug>", or
// the literal "ignore".
type SheetColumnMapping struct {
	Id           int       `orm:"auto;pk" json:"id"`
	ConnectionId int       `orm:"column(connection_id)" json:"connection_id"`
	ColumnIndex  int       `orm:"column(column_index)" json:"column_index"`
	Header       string    `orm:"null;size(300)" json:"header"`
	TargetField  string    `orm:"size(200)" json:"target_field"`
	Transform    string    `orm:"size(20);default(text)" json:"transform"`
	CreatedAt    time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
	UpdatedAt    time.Time `orm:"auto_now;type(timestamptz)" json:"updated_at"`
}

func (m *SheetColumnMapping) TableName() string {
	return "sheet_column_mappings"
}

func (m *SheetColumnMapping) IsIgnored() bool {
	return m.TargetField == "" || m.TargetField == "ignore"
}

func (m *SheetColumnMapping) IsCustom() bool {
	return strings.HasPrefix(m.TargetField, CustomFieldPrefix)
}

// CustomKey returns the custom field slug (without the prefix).
func (m *SheetColumnMapping) CustomKey() string {
	return strings.TrimPrefix(m.TargetField, CustomFieldPrefix)
}
