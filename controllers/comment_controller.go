package controllers

import (
	"PhoenixLab/services"
	"strconv"
	"strings"
)

type CommentController struct {
	BaseController
}

func (c *CommentController) Create() {
	c.RequireRole("technician", "admin", "super_admin", "viewer")

	ticketID, err := c.GetIntParam(":id")
	if err != nil {
		c.FlashError("Invalid ticket ID")
		c.Redirect("/tickets", 302)
		return
	}

	body := strings.TrimSpace(c.GetString("body"))
	if body == "" {
		c.FlashError("Comment cannot be empty")
		c.Redirect("/tickets/"+strconv.Itoa(ticketID), 302)
		return
	}

	commentService := services.CommentService{}
	_, err = commentService.Create(ticketID, c.GetCurrentUser().Id, body)
	if err != nil {
		c.FlashError("Failed to add comment: " + err.Error())
		c.Redirect("/tickets/"+strconv.Itoa(ticketID), 302)
		return
	}

	c.FlashSuccess("Comment added")
	c.Redirect("/tickets/"+strconv.Itoa(ticketID), 302)
}
