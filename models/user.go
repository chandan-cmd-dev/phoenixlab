package models

import (
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type Role string

const (
	RoleSuperAdmin Role = "super_admin"
	RoleAdmin      Role = "admin"
	RoleTechnician Role = "technician"
	RoleViewer     Role = "viewer"
)

type User struct {
	Id           int       `orm:"auto;pk" json:"id"`
	Name         string    `orm:"size(100)" json:"name" form:"name" validate:"required,min=2"`
	Email        string    `orm:"unique;size(200)" json:"email" form:"email" validate:"required,email"`
	PasswordHash string    `orm:"size(255)" json:"-"`
	Role         string    `orm:"size(50)" json:"role" form:"role" validate:"required,oneof=super_admin admin technician viewer"`
	BranchId     int       `orm:"column(branch_id)" json:"branch_id" form:"branch_id"`
	IsActive     bool      `orm:"default(true)" json:"is_active"`
	LastLoginAt  time.Time `orm:"null;type(timestamptz)" json:"last_login_at"`
	CreatedBy    int       `orm:"column(created_by)" json:"created_by"`
	CreatedAt    time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
	UpdatedAt    time.Time `orm:"auto_now;type(timestamptz)" json:"updated_at"`

	Branch        *Branch `orm:"-" json:"branch,omitempty"`
	CreatedByUser *User   `orm:"-" json:"created_by_user,omitempty"`
}

func (u *User) TableName() string {
	return "users"
}

func GetAllBranches() ([]*Branch, error) {
	o := orm.NewOrm()
	var branches []*Branch
	_, err := o.QueryTable("branches").Filter("is_active", true).All(&branches)
	return branches, err
}

func (u *User) IsAdmin() bool {
	return u.Role == string(RoleAdmin) || u.Role == string(RoleSuperAdmin)
}

func (u *User) IsSuperAdmin() bool {
	return u.Role == string(RoleSuperAdmin)
}

func (u *User) CanAccessBranch(branchId int) bool {
	if u.IsSuperAdmin() {
		return true
	}
	return u.BranchId == branchId
}
