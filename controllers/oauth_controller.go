package controllers

import (
	"PhoenixLab/services"
	"crypto/rand"
	"encoding/hex"
)

type OAuthController struct {
	BaseController
}

func (c *OAuthController) GoogleConnect() {
	c.RequireRole("admin", "super_admin")
	oauth := services.OAuthService{}
	if !oauth.Configured() {
		c.FlashError("Google OAuth is not configured (missing client credentials)")
		c.Redirect("/sheets", 302)
		return
	}
	state := randomState()
	c.safeSetSession("oauth_state", state)
	c.Redirect(oauth.AuthURL(state), 302)
}

func (c *OAuthController) GoogleCallback() {
	c.RequireRole("admin", "super_admin")
	oauth := services.OAuthService{}

	state := c.GetString("state")
	saved := c.safeGetSession("oauth_state")
	if saved == nil || state == "" || state != saved.(string) {
		c.FlashError("OAuth state mismatch — please try connecting again")
		c.Redirect("/sheets", 302)
		return
	}
	c.safeDelSession("oauth_state")

	if errParam := c.GetString("error"); errParam != "" {
		c.FlashError("Google authorization declined: " + errParam)
		c.Redirect("/sheets", 302)
		return
	}

	code := c.GetString("code")
	if code == "" {
		c.FlashError("No authorization code returned by Google")
		c.Redirect("/sheets", 302)
		return
	}

	tok, err := oauth.Exchange(code)
	if err != nil {
		c.FlashError("Token exchange failed: " + err.Error())
		c.Redirect("/sheets", 302)
		return
	}
	if err := oauth.SaveToken(tok, ""); err != nil {
		c.FlashError("Could not save token: " + err.Error())
		c.Redirect("/sheets", 302)
		return
	}
	c.FlashSuccess("Google account connected")
	c.Redirect("/sheets", 302)
}

func (c *OAuthController) GoogleDisconnect() {
	c.RequireRole("admin", "super_admin")
	(&services.OAuthService{}).Disconnect()
	c.FlashSuccess("Google account disconnected")
	c.Redirect("/sheets", 302)
}

func randomState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
