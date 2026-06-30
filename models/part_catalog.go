package models

import (
	"time"
)

type PartCatalog struct {
	Id           int       `orm:"auto;pk" json:"id"`
	PartNumber   string    `orm:"unique;size(100)" json:"part_number" form:"part_number" validate:"required"`
	Description  string    `orm:"type(text)" json:"description" form:"description" validate:"required"`
	UnitCost     float64   `orm:"digits(10);decimals(2);default(0)" json:"unit_cost" form:"unit_cost"`
	Supplier     string    `orm:"null;size(200)" json:"supplier" form:"supplier"`
	LeadTimeDays int       `orm:"default(0)" json:"lead_time_days" form:"lead_time_days"`
	IsActive     bool      `orm:"default(true)" json:"is_active" form:"is_active"`
	CreatedAt    time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
}

func (p *PartCatalog) TableName() string {
	return "parts_catalog"
}
