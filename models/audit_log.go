package models

import (
	"time"
)

type AuditLog struct {
	Id            int       `orm:"auto;pk" json:"id"`
	EntityType    string    `orm:"size(50)" json:"entity_type"`
	EntityId      int       `orm:"column(entity_id)" json:"entity_id"`
	Action        string    `orm:"size(50)" json:"action"`
	FieldName     string    `orm:"null;size(100)" json:"field_name"`
	OldValue      string    `orm:"null;type(text)" json:"old_value"`
	NewValue      string    `orm:"null;type(text)" json:"new_value"`
	ChangedBy     int       `orm:"column(changed_by)" json:"changed_by"`
	ChangedByName string    `orm:"null;size(100)" json:"changed_by_name"`
	IpAddress     string    `orm:"null;size(45)" json:"ip_address"`
	UserAgent     string    `orm:"null;type(text)" json:"user_agent"`
	ChangedAt     time.Time `orm:"auto_now_add;type(timestamptz)" json:"changed_at"`

	ChangedByUser *User `orm:"-" json:"changed_by_user,omitempty"`
}

func (a *AuditLog) TableName() string {
	return "audit_log"
}
