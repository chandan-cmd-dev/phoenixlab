package controllers

import (
	"PhoenixLab/models"
	"PhoenixLab/services"
	"strconv"
)

type UserController struct {
	BaseController
}

func (c *UserController) List() {
	c.RequireRole("admin", "super_admin")

	userService := services.UserService{}
	users, err := userService.GetAll(c.GetBranchScope(), c.GetCurrentUser().Role)
	if err != nil {
		c.FlashError("Failed to load users: " + err.Error())
		c.Redirect("/dashboard", 302)
		return
	}

	c.Data["users"] = users
	c.Data["title"] = "Manage Users"
	c.SetActivePage("users")
	c.GetFlashMessages()
	c.TplName = "users/list.html"
}

func (c *UserController) New() {
	c.RequireRole("admin", "super_admin")

	branches, _ := models.GetAllBranches()
	c.Data["branches"] = branches
	c.Data["title"] = "Create User"
	c.SetActivePage("users")
	c.GetFlashMessages()
	c.TplName = "users/form.html"
}

func (c *UserController) Create() {
	c.RequireRole("admin", "super_admin")

	u := &models.User{}
	if err := c.ParseForm(u); err != nil {
		c.FlashError("Invalid form data")
		c.Redirect("/users/new", 302)
		return
	}

	password := c.GetString("password")
	if password == "" {
		c.FlashError("Password is required")
		c.Redirect("/users/new", 302)
		return
	}

	u.CreatedBy = c.GetCurrentUser().Id
	u.IsActive = true

	userService := services.UserService{}
	if err := userService.Create(u, password); err != nil {
		c.FlashError("Could not create user: " + err.Error())
		c.Redirect("/users/new", 302)
		return
	}

	auditService := services.AuditService{}
	auditService.Log("user", u.Id, "create", "", "", u.Name, c.GetCurrentUser().Id, c.GetClientIP())

	c.FlashSuccess("User created successfully")
	c.Redirect("/users", 302)
}

func (c *UserController) Edit() {
	c.RequireRole("admin", "super_admin")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.FlashError("Invalid user ID")
		c.Redirect("/users", 302)
		return
	}

	userService := services.UserService{}
	user, err := userService.GetByID(id)
	if err != nil {
		c.FlashError("User not found")
		c.Redirect("/users", 302)
		return
	}

	if !c.GetCurrentUser().IsSuperAdmin() && user.BranchId != c.GetCurrentUser().BranchId {
		c.FlashError("You don't have permission to edit this user")
		c.Redirect("/users", 302)
		return
	}

	branches, _ := models.GetAllBranches()
	c.Data["user"] = user
	c.Data["branches"] = branches
	c.Data["title"] = "Edit User"
	c.SetActivePage("users")
	c.GetFlashMessages()
	c.TplName = "users/form.html"
}

func (c *UserController) Update() {
	c.RequireRole("admin", "super_admin")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.FlashError("Invalid user ID")
		c.Redirect("/users", 302)
		return
	}

	userService := services.UserService{}
	user, err := userService.GetByID(id)
	if err != nil {
		c.FlashError("User not found")
		c.Redirect("/users", 302)
		return
	}

	if !c.GetCurrentUser().IsSuperAdmin() && user.BranchId != c.GetCurrentUser().BranchId {
		c.FlashError("You don't have permission to edit this user")
		c.Redirect("/users", 302)
		return
	}

	oldRole := user.Role
	oldBranchId := user.BranchId
	if err := c.ParseForm(user); err != nil {
		c.FlashError("Invalid form data")
		c.Redirect("/users/edit/"+strconv.Itoa(id), 302)
		return
	}

	fields := []string{"Name", "Email", "Role", "BranchId"}
	if err := userService.Update(user, fields); err != nil {
		c.FlashError("Failed to update user: " + err.Error())
		c.Redirect("/users/edit/"+strconv.Itoa(id), 302)
		return
	}

	auditService := services.AuditService{}
	if oldRole != user.Role {
		auditService.Log("user", user.Id, "update", "role", oldRole, user.Role, c.GetCurrentUser().Id, c.GetClientIP())
	}
	if oldBranchId != user.BranchId {
		auditService.Log("user", user.Id, "update", "branch_id", strconv.Itoa(oldBranchId), strconv.Itoa(user.BranchId), c.GetCurrentUser().Id, c.GetClientIP())
	}

	c.FlashSuccess("User updated successfully")
	c.Redirect("/users", 302)
}

func (c *UserController) ToggleActive() {
	c.RequireRole("admin", "super_admin")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.FlashError("Invalid user ID")
		c.Redirect("/users", 302)
		return
	}

	userService := services.UserService{}
	user, err := userService.GetByID(id)
	if err != nil {
		c.FlashError("User not found")
		c.Redirect("/users", 302)
		return
	}

	if !c.GetCurrentUser().IsSuperAdmin() && user.BranchId != c.GetCurrentUser().BranchId {
		c.FlashError("You don't have permission to modify this user")
		c.Redirect("/users", 302)
		return
	}

	if user.Id == c.GetCurrentUser().Id {
		c.FlashError("You cannot deactivate your own account")
		c.Redirect("/users", 302)
		return
	}

	if err := userService.ToggleActive(id, c.GetCurrentUser().Id); err != nil {
		c.FlashError("Failed to toggle user status: " + err.Error())
		c.Redirect("/users", 302)
		return
	}

	c.FlashSuccess("User status updated successfully")
	c.Redirect("/users", 302)
}
