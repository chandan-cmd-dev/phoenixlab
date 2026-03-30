package controllers

import (
	"PhoenixLab/models"
	"PhoenixLab/services"
	"strconv"
	"time"
)

type TicketController struct {
	BaseController
}

func (c *TicketController) List() {
	c.RequireRole("technician", "admin", "super_admin", "viewer")

	filters := make(map[string]string)
	if status := c.GetString("status"); status != "" {
		filters["status"] = status
	}
	if warranty := c.GetString("warranty"); warranty != "" {
		filters["warranty"] = warranty
	}
	if q := c.GetString("q"); q != "" {
		filters["q"] = q
	}
	if assigned := c.GetString("assigned"); assigned != "" {
		filters["assigned"] = assigned
	}
	if brand := c.GetString("brand"); brand != "" {
		filters["brand"] = brand
	}

	ticketService := services.TicketService{}
	tickets, err := ticketService.GetByBranch(c.GetBranchScope(), filters)
	if err != nil {
		c.FlashError("Failed to load tickets: " + err.Error())
		c.Redirect("/dashboard", 302)
		return
	}

	userService := services.UserService{}
	technicians, _ := userService.GetAll(c.GetBranchScope(), c.GetCurrentUser().Role)

	c.Data["tickets"] = tickets
	c.Data["technicians"] = technicians
	c.Data["filters"] = filters
	c.Data["title"] = "Tickets"
	c.SetActivePage("tickets")
	c.GetFlashMessages()
	c.TplName = "tickets/list.html"
}

func (c *TicketController) New() {
	c.RequireRole("technician", "admin", "super_admin")

	userService := services.UserService{}
	technicians, _ := userService.GetAll(c.GetBranchScope(), c.GetCurrentUser().Role)
	c.Data["technicians"] = technicians
	c.Data["title"] = "Create Ticket"
	c.SetActivePage("tickets")
	c.GetFlashMessages()
	c.TplName = "tickets/form.html"
}

func (c *TicketController) Create() {
	c.RequireRole("technician", "admin", "super_admin")

	t := &models.Ticket{}
	if err := c.ParseForm(t); err != nil {
		c.FlashError("Invalid form data")
		c.Redirect("/tickets/new", 302)
		return
	}
	if !c.GetCurrentUser().IsSuperAdmin() {
		t.BranchId = c.GetCurrentUser().BranchId
	} else if t.BranchId == 0 {
		if branches, err := models.GetAllBranches(); err == nil && len(branches) > 0 {
			t.BranchId = branches[0].Id
		}
	}

	if t.SerialNumber == "" {
		c.FlashError("Serial number is required")
		c.Redirect("/tickets/new", 302)
		return
	}
	if t.BranchId == 0 {
		c.FlashError("Branch is not configured; please contact administrator")
		c.Redirect("/tickets/new", 302)
		return
	}

	ticketService := services.TicketService{}
	if err := ticketService.Create(t, c.GetCurrentUser().Id); err != nil {
		c.FlashError("Could not create ticket: " + err.Error())
		c.Redirect("/tickets/new", 302)
		return
	}

	c.FlashSuccess("Ticket created successfully")
	c.Redirect("/tickets", 302)
}

func (c *TicketController) Detail() {
	c.RequireRole("technician", "admin", "super_admin", "viewer")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.FlashError("Invalid ticket ID")
		c.Redirect("/tickets", 302)
		return
	}

	ticketService := services.TicketService{}
	ticket, err := ticketService.GetByID(id)
	if err != nil {
		c.FlashError("Ticket not found")
		c.Redirect("/tickets", 302)
		return
	}

	if !c.GetCurrentUser().CanAccessBranch(ticket.BranchId) {
		c.FlashError("You don't have permission to view this ticket")
		c.Redirect("/tickets", 302)
		return
	}

	auditService := services.AuditService{}
	auditLogs, _ := auditService.GetForTicket(id)

	commentService := services.CommentService{}
	comments, _ := commentService.GetByTicket(id)

	c.Data["ticket"] = ticket
	c.Data["auditLogs"] = auditLogs
	c.Data["comments"] = comments
	c.Data["title"] = "Ticket #" + strconv.Itoa(id)
	c.SetActivePage("tickets")
	c.GetFlashMessages()
	c.TplName = "tickets/detail.html"
}

func (c *TicketController) Edit() {
	c.RequireRole("technician", "admin", "super_admin")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.FlashError("Invalid ticket ID")
		c.Redirect("/tickets", 302)
		return
	}

	ticketService := services.TicketService{}
	ticket, err := ticketService.GetByID(id)
	if err != nil {
		c.FlashError("Ticket not found")
		c.Redirect("/tickets", 302)
		return
	}

	if !c.getCurrentUserCanEditTicket(ticket) {
		c.FlashError("You don't have permission to edit this ticket")
		c.Redirect("/tickets/"+strconv.Itoa(id), 302)
		return
	}

	userService := services.UserService{}
	technicians, _ := userService.GetAll(c.GetBranchScope(), c.GetCurrentUser().Role)
	c.Data["ticket"] = ticket
	c.Data["technicians"] = technicians
	c.Data["title"] = "Edit Ticket #" + strconv.Itoa(id)
	c.SetActivePage("tickets")
	c.GetFlashMessages()
	c.TplName = "tickets/form.html"
}

func (c *TicketController) Update() {
	c.RequireRole("technician", "admin", "super_admin")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.FlashError("Invalid ticket ID")
		c.Redirect("/tickets", 302)
		return
	}

	ticketService := services.TicketService{}
	ticket, err := ticketService.GetByID(id)
	if err != nil {
		c.FlashError("Ticket not found")
		c.Redirect("/tickets", 302)
		return
	}

	if !c.getCurrentUserCanEditTicket(ticket) {
		c.FlashError("You don't have permission to edit this ticket")
		c.Redirect("/tickets/"+strconv.Itoa(id), 302)
		return
	}

	originalTicket := *ticket
	if err := c.ParseForm(ticket); err != nil {
		c.FlashError("Invalid form data")
		c.Redirect("/tickets/"+strconv.Itoa(id)+"/edit", 302)
		return
	}

	if !c.GetCurrentUser().IsSuperAdmin() {
		ticket.BranchId = c.GetCurrentUser().BranchId
	}

	if dueDate := c.GetString("due_date"); dueDate != "" {
		if parsed, err := time.Parse("2006-01-02", dueDate); err == nil {
			ticket.DueDate = parsed
		}
	}

	changedFields := c.getChangedFields(&originalTicket, ticket)

	if err := ticketService.Update(ticket, changedFields, c.GetCurrentUser().Id); err != nil {
		c.FlashError("Failed to update ticket: " + err.Error())
		c.Redirect("/tickets/"+strconv.Itoa(id)+"/edit", 302)
		return
	}

	c.FlashSuccess("Ticket updated successfully")
	c.Redirect("/tickets/"+strconv.Itoa(id), 302)
}

func (c *TicketController) UpdateStatus() {
	c.RequireRole("technician", "admin", "super_admin")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.RenderError(400, "Invalid ticket ID")
		return
	}

	newStatus := c.GetString("status")
	if newStatus == "" {
		c.RenderError(400, "Status is required")
		return
	}

	ticketService := services.TicketService{}
	ticket, err := ticketService.GetByID(id)
	if err != nil {
		c.RenderError(404, "Ticket not found")
		return
	}

	if !c.getCurrentUserCanEditTicket(ticket) {
		c.RenderError(403, "Permission denied")
		return
	}

	if err := ticketService.UpdateStatus(id, newStatus, c.GetCurrentUser().Id); err != nil {
		c.RenderError(400, "Failed to update status: "+err.Error())
		return
	}

	c.RenderJSON(map[string]interface{}{
		"success": true,
		"message": "Status updated successfully",
		"status":  newStatus,
	})
}

func (c *TicketController) Assign() {
	c.RequireRole("admin", "super_admin")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.RenderError(400, "Invalid ticket ID")
		return
	}

	assignedTo, err := c.GetInt("assigned_to")
	if err != nil {
		c.RenderError(400, "Invalid technician ID")
		return
	}

	ticketService := services.TicketService{}
	ticket, err := ticketService.GetByID(id)
	if err != nil {
		c.RenderError(404, "Ticket not found")
		return
	}

	if !c.GetCurrentUser().CanAccessBranch(ticket.BranchId) {
		c.RenderError(403, "Permission denied")
		return
	}

	if err := ticketService.Assign(id, assignedTo, c.GetCurrentUser().Id); err != nil {
		c.RenderError(400, "Failed to assign ticket: "+err.Error())
		return
	}

	c.RenderJSON(map[string]interface{}{
		"success": true,
		"message": "Ticket assigned successfully",
	})
}

func (c *TicketController) Delete() {
	c.RequireRole("super_admin")

	id, err := c.GetIntParam(":id")
	if err != nil {
		c.FlashError("Invalid ticket ID")
		c.Redirect("/tickets", 302)
		return
	}

	ticketService := services.TicketService{}
	if err := ticketService.Delete(id, c.GetCurrentUser().Id); err != nil {
		c.FlashError("Failed to delete ticket: " + err.Error())
		c.Redirect("/tickets", 302)
		return
	}

	c.FlashSuccess("Ticket deleted successfully")
	c.Redirect("/tickets", 302)
}

func (c *TicketController) getCurrentUserCanEditTicket(ticket *models.Ticket) bool {
	user := c.GetCurrentUser()

	if user.IsSuperAdmin() {
		return true
	}

	if user.IsAdmin() && user.BranchId == ticket.BranchId {
		return true
	}

	if user.Role == string(models.RoleTechnician) && user.BranchId == ticket.BranchId {
		return ticket.AssignedTo == user.Id || ticket.AssignedTo == 0
	}

	return false
}

func (c *TicketController) getChangedFields(original, updated *models.Ticket) []string {
	var fields []string

	if original.SerialNumber != updated.SerialNumber {
		fields = append(fields, "SerialNumber")
	}
	if original.IrNumber != updated.IrNumber {
		fields = append(fields, "IrNumber")
	}
	if original.Upc != updated.Upc {
		fields = append(fields, "Upc")
	}
	if original.Model != updated.Model {
		fields = append(fields, "Model")
	}
	if original.Brand != updated.Brand {
		fields = append(fields, "Brand")
	}
	if original.WarrantyStatus != updated.WarrantyStatus {
		fields = append(fields, "WarrantyStatus")
	}
	if original.IssueDescription != updated.IssueDescription {
		fields = append(fields, "IssueDescription")
	}
	if original.IssueCategory != updated.IssueCategory {
		fields = append(fields, "IssueCategory")
	}
	if original.AssignedTo != updated.AssignedTo {
		fields = append(fields, "AssignedTo")
	}
	if original.Priority != updated.Priority {
		fields = append(fields, "Priority")
	}
	if original.DiagnosticCode != updated.DiagnosticCode {
		fields = append(fields, "DiagnosticCode")
	}
	if original.NeededPart != updated.NeededPart {
		fields = append(fields, "NeededPart")
	}
	if original.ProblemDescription != updated.ProblemDescription {
		fields = append(fields, "ProblemDescription")
	}
	if original.MachinePurchasePrice != updated.MachinePurchasePrice {
		fields = append(fields, "MachinePurchasePrice")
	}
	if original.PartNumber != updated.PartNumber {
		fields = append(fields, "PartNumber")
	}
	if original.PartsCost != updated.PartsCost {
		fields = append(fields, "PartsCost")
	}
	if original.LabourCost != updated.LabourCost {
		fields = append(fields, "LabourCost")
	}
	if original.PoNumber != updated.PoNumber {
		fields = append(fields, "PoNumber")
	}
	if original.CaseNumber != updated.CaseNumber {
		fields = append(fields, "CaseNumber")
	}
	if original.WorkOrderNumber != updated.WorkOrderNumber {
		fields = append(fields, "WorkOrderNumber")
	}
	if original.CourierName != updated.CourierName {
		fields = append(fields, "CourierName")
	}
	if original.CourierTracking != updated.CourierTracking {
		fields = append(fields, "CourierTracking")
	}
	if original.TrackingLink != updated.TrackingLink {
		fields = append(fields, "TrackingLink")
	}
	if original.ReturnPart != updated.ReturnPart {
		fields = append(fields, "ReturnPart")
	}
	if original.ReturnTracking != updated.ReturnTracking {
		fields = append(fields, "ReturnTracking")
	}
	if original.PartArrivedFixed != updated.PartArrivedFixed {
		fields = append(fields, "PartArrivedFixed")
	}
	if original.DefectivePartShipped != updated.DefectivePartShipped {
		fields = append(fields, "DefectivePartShipped")
	}
	if original.CaseFinished != updated.CaseFinished {
		fields = append(fields, "CaseFinished")
	}
	if original.CustomerName != updated.CustomerName {
		fields = append(fields, "CustomerName")
	}
	if original.CustomerEmail != updated.CustomerEmail {
		fields = append(fields, "CustomerEmail")
	}
	if original.CustomerPhone != updated.CustomerPhone {
		fields = append(fields, "CustomerPhone")
	}
	if original.DueDate != updated.DueDate {
		fields = append(fields, "DueDate")
	}
	if original.Notes != updated.Notes {
		fields = append(fields, "Notes")
	}

	return fields
}
