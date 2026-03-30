package models

import (
	"time"
	"github.com/beego/beego/v2/client/orm"
)

type Branch struct {
	Id        int       `orm:"auto;pk" json:"id"`
	Name      string    `orm:"size(100)" json:"name" form:"name" validate:"required"`
	Code      string    `orm:"unique;size(20)" json:"code" form:"code" validate:"required"`
	Address   string    `orm:"null;type(text)" json:"address" form:"address"`
	Phone     string    `orm:"null;size(50)" json:"phone" form:"phone"`
	IsActive  bool      `orm:"default(true)" json:"is_active"`
	CreatedAt time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
}

func (b *Branch) TableName() string {
	return "branches"
}

func GetOrm() orm.Ormer {
	return orm.NewOrm()
}
