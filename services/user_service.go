package services

import (
	"PhoenixLab/models"
	"errors"

	"github.com/beego/beego/v2/client/orm"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct{}

func (s *UserService) Create(u *models.User, plainPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), 12)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	o := orm.NewOrm()
	_, err = o.Insert(u)
	return err
}

func (s *UserService) Authenticate(email, password string) (*models.User, error) {
	o := orm.NewOrm()
	u := &models.User{Email: email}
	if err := o.Read(u, "Email"); err != nil {
		return nil, errors.New("invalid credentials")
	}
	if !u.IsActive {
		return nil, errors.New("account disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

func (s *UserService) GetByID(id int) (*models.User, error) {
	o := orm.NewOrm()
	u := &models.User{Id: id}
	if err := o.Read(u); err != nil {
		return nil, err
	}

	if u.BranchId > 0 {
		branch := &models.Branch{Id: u.BranchId}
		if err := o.Read(branch); err == nil {
			u.Branch = branch
		}
	}

	return u, nil
}

func (s *UserService) GetAll(branchId int, role string) ([]*models.User, error) {
	o := orm.NewOrm()
	qs := o.QueryTable("users")

	if role != string(models.RoleSuperAdmin) && branchId > 0 {
		qs = qs.Filter("branch_id", branchId)
	}

	var users []*models.User
	_, err := qs.OrderBy("-created_at").All(&users)
	return users, err
}

func (s *UserService) Update(u *models.User, fields []string) error {
	o := orm.NewOrm()
	_, err := o.Update(u, fields...)
	return err
}

func (s *UserService) UpdateRole(userID int, newRole models.Role, actorID int) error {
	o := orm.NewOrm()
	u := &models.User{Id: userID}
	if err := o.Read(u); err != nil {
		return err
	}
	oldRole := u.Role
	u.Role = string(newRole)
	if _, err := o.Update(u, "Role"); err != nil {
		return err
	}

	auditService := AuditService{}
	auditService.Log("user", userID, "update", "role", oldRole, string(newRole), actorID, "")
	return nil
}

func (s *UserService) ToggleActive(userID int, actorID int) error {
	o := orm.NewOrm()
	u := &models.User{Id: userID}
	if err := o.Read(u); err != nil {
		return err
	}

	u.IsActive = !u.IsActive
	if _, err := o.Update(u, "IsActive"); err != nil {
		return err
	}

	auditService := AuditService{}
	auditService.Log("user", userID, "update", "is_active",
		map[bool]string{true: "active", false: "inactive"}[!u.IsActive],
		map[bool]string{true: "active", false: "inactive"}[u.IsActive],
		actorID, "")
	return nil
}

func (s *UserService) GetByEmail(email string) (*models.User, error) {
	o := orm.NewOrm()
	u := &models.User{Email: email}
	if err := o.Read(u, "Email"); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) UpdatePassword(userID int, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	o := orm.NewOrm()
	u := &models.User{Id: userID}
	u.PasswordHash = string(hash)
	_, err = o.Update(u, "PasswordHash")
	return err
}

func (s *UserService) UpdateLastLogin(userID int) error {
	o := orm.NewOrm()
	u := &models.User{Id: userID}
	_, err := o.Update(u, "LastLoginAt")
	return err
}
