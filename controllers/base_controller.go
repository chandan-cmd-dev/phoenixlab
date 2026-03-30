package controllers

import (
	"PhoenixLab/models"
	"PhoenixLab/services"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/beego/beego/v2/server/web"
)

type BaseController struct {
	web.Controller
	currentUser *models.User
}

func (c *BaseController) Prepare() {
	c.loadCurrentUser()
}

func (c *BaseController) loadCurrentUser() {
	if c.CruSession == nil {
		c.StartSession()
	}
	if c.CruSession == nil {
		return
	}
	sessionUser := c.CruSession.Get(context.Background(), "user_id")
	if sessionUser == nil {
		return
	}

	userID := sessionUser.(int)
	userService := services.UserService{}
	user, err := userService.GetByID(userID)
	if err != nil || !user.IsActive {
		c.DestroySession()
		return
	}

	c.currentUser = user
	c.Data["currentUser"] = user

	if user.BranchId > 0 {
		branch, _ := models.GetAllBranches()
		for _, b := range branch {
			if b.Id == user.BranchId {
				c.Data["branchName"] = b.Name
				break
			}
		}
	}
}

func (c *BaseController) RequireRole(roles ...string) {
	if c.currentUser == nil {
		c.Redirect("/auth/login", 302)
		c.StopRun()
		return
	}

	for _, r := range roles {
		if c.currentUser.Role == r {
			return
		}
	}

	c.Abort("403")
}

func (c *BaseController) RequireBranchAccess(branchId int) {
	if c.currentUser == nil {
		c.Abort("403")
		return
	}

	if !c.currentUser.CanAccessBranch(branchId) {
		c.Abort("403")
		return
	}
}

func (c *BaseController) safeGetSession(key string) interface{} {
	if c.CruSession == nil {
		return nil
	}
	return c.CruSession.Get(context.Background(), key)
}

func (c *BaseController) safeSetSession(key string, val interface{}) {
	if c.CruSession == nil {
		c.StartSession()
	}
	if c.CruSession == nil {
		return
	}
	c.CruSession.Set(context.Background(), key, val)
}

func (c *BaseController) safeDelSession(key string) {
	if c.CruSession == nil {
		return
	}
	c.CruSession.Delete(context.Background(), key)
}

func (c *BaseController) FlashSuccess(msg string) {
	c.safeSetSession("flash_success", msg)
	c.Data["flash_success"] = msg
}

func (c *BaseController) FlashError(msg string) {
	c.safeSetSession("flash_error", msg)
	c.Data["flash_error"] = msg
}

func (c *BaseController) GetFlashMessages() {
	if success := c.safeGetSession("flash_success"); success != nil {
		c.Data["flash_success"] = success.(string)
		c.safeDelSession("flash_success")
	}
	if errMsg := c.safeGetSession("flash_error"); errMsg != nil {
		c.Data["flash_error"] = errMsg.(string)
		c.safeDelSession("flash_error")
	}
}

func (c *BaseController) GetIntParam(key string) (int, error) {
	value := c.Ctx.Input.Param(key)
	if value == "" {
		return 0, fmt.Errorf("parameter %s not found", key)
	}
	return strconv.Atoi(value)
}

func (c *BaseController) GetCurrentUser() *models.User {
	return c.currentUser
}

func (c *BaseController) GetBranchScope() int {
	if c.currentUser.IsSuperAdmin() {
		return 0
	}
	return c.currentUser.BranchId
}

func (c *BaseController) SetActivePage(page string) {
	c.Data["activePage"] = page
}

func (c *BaseController) GetClientIP() string {
	if ip := c.Ctx.Request.Header.Get("X-Forwarded-For"); ip != "" {
		return strings.Split(ip, ",")[0]
	}
	if ip := c.Ctx.Request.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return c.Ctx.Request.RemoteAddr
}

func (c *BaseController) GetUserAgent() string {
	return c.Ctx.Request.UserAgent()
}

func (c *BaseController) RenderJSON(data interface{}) {
	c.Data["json"] = data
	c.ServeJSON()
}

func (c *BaseController) RenderError(status int, message string) {
	c.Ctx.ResponseWriter.WriteHeader(status)
	c.Data["error"] = message
	c.Data["json"] = map[string]interface{}{
		"error":  message,
		"status": status,
	}
	c.ServeJSON()
}
