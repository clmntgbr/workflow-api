package handler

import (
	"errors"

	runexportcmd "go-api/internal/application/command/runexport"
	queryrunexport "go-api/internal/application/query/runexport"
	queryworkflow "go-api/internal/application/query/workflow"
	domainrunexport "go-api/internal/domain/runexport"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type RunExportHandler struct {
	requestHandler        runExportRequestHandler
	getByIDHandler        runExportGetByIDHandler
	getWorkflowHandler    runExportGetWorkflowHandler
}

func NewRunExportHandler(
	requestHandler runExportRequestHandler,
	getByIDHandler runExportGetByIDHandler,
	getWorkflowHandler runExportGetWorkflowHandler,
) *RunExportHandler {
	return &RunExportHandler{
		requestHandler:     requestHandler,
		getByIDHandler:     getByIDHandler,
		getWorkflowHandler: getWorkflowHandler,
	}
}

func (h *RunExportHandler) Create(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	projectID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	workflow, err := h.getWorkflowHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: workflowID})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}
	if workflow.ProjectID != projectID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	var req dto.RequestRunExportRequest
	if len(c.Body()) > 0 {
		if err := validation.BindBody(c, &req); err != nil {
			return err
		}
	}

	job, err := h.requestHandler.Handle(c.Context(), runexportcmd.RequestRunExportCommand{
		UserID:     user.ID,
		ProjectID:  projectID,
		WorkflowID: workflowID,
		From:       req.From,
		To:         req.To,
	})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		if errors.Is(err, domainrunexport.ErrInvalidDateRange) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": domainrunexport.ErrInvalidDateRange.Error()})
		}
		if handled, resp := respondQuotaError(c, err); handled {
			return resp
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to request run export"})
	}

	return c.Status(fiber.StatusAccepted).JSON(presenter.NewRunExportAcceptedResponse(*job))
}

func (h *RunExportHandler) GetByID(c fiber.Ctx) error {
	_, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	projectID, err := httpctx.GetActiveProjectID(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Active project is required"})
	}

	workflowID, err := uuid.Parse(c.Params("workflowId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid workflow id"})
	}

	jobID, err := uuid.Parse(c.Params("jobId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid job id"})
	}

	workflow, err := h.getWorkflowHandler.Handle(c.Context(), queryworkflow.GetWorkflowByIDQuery{ID: workflowID})
	if err != nil {
		if err.Error() == "workflow not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get workflow"})
	}
	if workflow.ProjectID != projectID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Workflow not found"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryrunexport.GetRunExportByIDQuery{ID: jobID})
	if err != nil {
		if err.Error() == "run export not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Run export not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get run export"})
	}
	if view.WorkflowID != workflowID || view.ProjectID != projectID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Run export not found"})
	}

	return c.Status(fiber.StatusOK).JSON(presenter.NewRunExportResponseFromView(*view))
}
