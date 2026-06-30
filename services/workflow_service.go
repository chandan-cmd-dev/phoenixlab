package services

import (
	"PhoenixLab/models"
	"errors"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type WorkflowService struct{}

func (s *WorkflowService) Start(ticketID int, workflowType string, userID int) (*models.TicketWorkflow, error) {
	steps, ok := models.WorkflowDefinitions[workflowType]
	if !ok || len(steps) == 0 {
		return nil, errors.New("unknown workflow type")
	}

	o := orm.NewOrm()

	existing := &models.TicketWorkflow{}
	err := o.QueryTable("ticket_workflows").
		Filter("TicketId", ticketID).
		Filter("WorkflowType", workflowType).
		Filter("CompletedAt__isnull", true).
		One(existing)
	if err == nil {
		return nil, errors.New("an active workflow of this type already exists for this ticket")
	}

	wf := &models.TicketWorkflow{
		TicketId:     ticketID,
		WorkflowType: workflowType,
		CurrentStep:  steps[0],
		StepData:     "{}",
		StartedAt:    time.Now(),
		StartedBy:    userID,
	}

	_, err = o.Insert(wf)
	if err != nil {
		return nil, err
	}

	audit := AuditService{}
	audit.Log("ticket", ticketID, "create", "workflow", "", workflowType, userID, "")

	return wf, nil
}

func (s *WorkflowService) Advance(workflowID int, userID int) error {
	o := orm.NewOrm()

	wf := &models.TicketWorkflow{Id: workflowID}
	if err := o.Read(wf); err != nil {
		return errors.New("workflow not found")
	}

	if wf.IsComplete() {
		return errors.New("workflow is already complete")
	}

	if !wf.CanAdvance() {
		return errors.New("workflow cannot advance further")
	}

	oldStep := wf.CurrentStep
	newStep := wf.NextStep()
	wf.CurrentStep = newStep

	steps := wf.GetSteps()
	if newStep == steps[len(steps)-1] {
		wf.CompletedAt = time.Now()
		_, err := o.Update(wf, "CurrentStep", "CompletedAt", "UpdatedAt")
		if err != nil {
			return err
		}
	} else {
		_, err := o.Update(wf, "CurrentStep", "UpdatedAt")
		if err != nil {
			return err
		}
	}

	audit := AuditService{}
	audit.Log("ticket", wf.TicketId, "update", "workflow_step", oldStep, newStep, userID, "")

	return nil
}

func (s *WorkflowService) GetByTicket(ticketID int) ([]*models.TicketWorkflow, error) {
	o := orm.NewOrm()
	var workflows []*models.TicketWorkflow
	_, err := o.QueryTable("ticket_workflows").
		Filter("TicketId", ticketID).
		OrderBy("-StartedAt").
		All(&workflows)

	for _, wf := range workflows {
		if wf.StartedBy > 0 {
			user := &models.User{Id: wf.StartedBy}
			if err := o.Read(user); err == nil {
				wf.Starter = user
			}
		}
	}

	return workflows, err
}

func (s *WorkflowService) GetByID(id int) (*models.TicketWorkflow, error) {
	o := orm.NewOrm()
	wf := &models.TicketWorkflow{Id: id}
	if err := o.Read(wf); err != nil {
		return nil, err
	}
	return wf, nil
}
