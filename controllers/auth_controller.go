package controllers

import (
	"PhoenixLab/services"
)

type AuthController struct {
	BaseController
}

func (c *AuthController) Prepare() {
	c.BaseController.Prepare()
	c.GetFlashMessages()
}

func (c *AuthController) ShowLogin() {
	if c.safeGetSession("user_id") != nil {
		c.Redirect("/dashboard", 302)
		return
	}

	c.GetFlashMessages()
	c.Data["title"] = "Login - Phoenix Lab"
	c.TplName = "auth/login.html"
}

func (c *AuthController) Login() {
	email := c.GetString("email")
	password := c.GetString("password")

	if email == "" || password == "" {
		c.FlashError("Email and password are required")
		c.Redirect("/auth/login", 302)
		return
	}

	userService := services.UserService{}
	user, err := userService.Authenticate(email, password)
	if err != nil {
		c.FlashError("Invalid credentials")
		c.Redirect("/auth/login", 302)
		return
	}

	userService.UpdateLastLogin(user.Id)

	c.safeSetSession("user_id", user.Id)
	c.safeSetSession("user_name", user.Name)
	c.safeSetSession("user_role", user.Role)

	auditService := services.AuditService{}
	auditService.Log("user", user.Id, "login", "", "", "", user.Id, c.GetClientIP())

	c.FlashSuccess("Welcome back, " + user.Name + "!")
	c.Redirect("/dashboard", 302)
}

func (c *AuthController) Logout() {
	if userID := c.safeGetSession("user_id"); userID != nil {
		auditService := services.AuditService{}
		auditService.Log("user", userID.(int), "logout", "", "", "", userID.(int), c.GetClientIP())
	}

	if c.CruSession != nil {
		c.DestroySession()
	}
	c.FlashSuccess("You have been logged out")
	c.Redirect("/auth/login", 302)
}

func (c *AuthController) Profile() {
	c.RequireRole("technician", "admin", "super_admin", "viewer")

	user := c.GetCurrentUser()
	c.Data["user"] = user
	c.Data["title"] = "My Profile"
	c.SetActivePage("profile")
	c.GetFlashMessages()
	c.TplName = "users/profile.html"
}

func (c *AuthController) UpdateProfile() {
	c.RequireRole("technician", "admin", "super_admin", "viewer")

	user := c.GetCurrentUser()

	name := c.GetString("name")
	email := c.GetString("email")

	if name != "" {
		user.Name = name
	}
	if email != "" && email != user.Email {
		userService := services.UserService{}
		if existingUser, err := userService.GetByEmail(email); err == nil && existingUser.Id != user.Id {
			c.FlashError("Email is already taken by another user")
			c.Redirect("/auth/profile", 302)
			return
		}
		user.Email = email
	}

	newPassword := c.GetString("new_password")
	if newPassword != "" {
		currentPassword := c.GetString("current_password")
		if currentPassword == "" {
			c.FlashError("Current password is required to change password")
			c.Redirect("/auth/profile", 302)
			return
		}

		userService := services.UserService{}
		if _, err := userService.Authenticate(user.Email, currentPassword); err != nil {
			c.FlashError("Current password is incorrect")
			c.Redirect("/auth/profile", 302)
			return
		}

		if err := userService.UpdatePassword(user.Id, newPassword); err != nil {
			c.FlashError("Failed to update password: " + err.Error())
			c.Redirect("/auth/profile", 302)
			return
		}

		c.FlashSuccess("Password updated successfully")
	}

	if lang := c.GetString("language"); lang != "" {
		user.Language = lang
	}

	userService := services.UserService{}
	if err := userService.Update(user, []string{"Name", "Email", "Language"}); err != nil {
		c.FlashError("Failed to update profile: " + err.Error())
		c.Redirect("/auth/profile", 302)
		return
	}

	c.safeSetSession("user_name", user.Name)
	c.safeSetSession("user_lang", user.Language)

	c.FlashSuccess("Profile updated successfully")
	c.Redirect("/auth/profile", 302)
}
