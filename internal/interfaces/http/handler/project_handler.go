package handler

import (
	"strings"

	projectcmd "go-api/internal/application/command/project"
	usercmd "go-api/internal/application/command/user"
	queryproject "go-api/internal/application/query/project"
	"go-api/internal/domain/paginate"
	httpctx "go-api/internal/interfaces/http/context"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/validation"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	createHandler           projectCreateHandler
	updateHandler           projectUpdateHandler
	deleteHandler           projectDeleteHandler
	removeMemberHandler     projectRemoveMemberHandler
	getByIDHandler          projectGetByIDHandler
	listByUserHandler       projectListByUserHandler
	setActiveProjectHandler projectSetActiveHandler
}

func NewProjectHandler(
	createHandler projectCreateHandler,
	updateHandler projectUpdateHandler,
	deleteHandler projectDeleteHandler,
	removeMemberHandler projectRemoveMemberHandler,
	getByIDHandler projectGetByIDHandler,
	listByUserHandler projectListByUserHandler,
	setActiveProjectHandler projectSetActiveHandler,
) *ProjectHandler {
	return &ProjectHandler{
		createHandler:           createHandler,
		updateHandler:           updateHandler,
		deleteHandler:           deleteHandler,
		removeMemberHandler:     removeMemberHandler,
		getByIDHandler:          getByIDHandler,
		listByUserHandler:       listByUserHandler,
		setActiveProjectHandler: setActiveProjectHandler,
	}
}

func (h *ProjectHandler) List(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var query paginate.PaginateQuery
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid query parameters",
		})
	}

	sortBy, orderBy := query.SortBy, query.OrderBy
	query.Normalize()
	if sortBy == "" {
		query.SortBy = "created_at"
	}
	if orderBy == "" {
		query.OrderBy = paginate.OrderByAsc
	}

	views, total, err := h.listByUserHandler.Handle(c.Context(), queryproject.ListProjectsByUserQuery{
		UserID: user.ID,
		Query:  query,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to list projects"})
	}

	return c.Status(fiber.StatusOK).JSON(paginate.NewPaginateResponse(
		presenter.NewProjectListResponseFromViews(views, user.ActiveProjectID),
		int(total),
		query,
	))
}

func (h *ProjectHandler) Create(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	var req dto.CreateProjectRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	project, err := h.createHandler.Handle(c.Context(), projectcmd.CreateProjectCommand{
		Name:          req.Name,
		CreatorUserID: user.ID,
	})
	if err != nil {
		if handled, resp := respondQuotaError(c, err); handled {
			return resp
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to create project"})
	}

	activeID := project.ID
	return c.Status(fiber.StatusCreated).JSON(
		presenter.NewProjectDetailResponseFromEntity(*project, &activeID),
	)
}

func (h *ProjectHandler) GetByID(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid project id"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryproject.GetProjectByIDQuery{ID: id})
	if err != nil {
		if err.Error() == "project not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Project not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get project"})
	}

	return c.Status(fiber.StatusOK).JSON(
		presenter.NewProjectDetailResponseFromView(*view, user.ActiveProjectID),
	)
}

func (h *ProjectHandler) Update(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid project id"})
	}

	var req dto.UpdateProjectRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}
	req.Name = strings.TrimSpace(req.Name)
	if err := validation.Struct(c, &req); err != nil {
		return err
	}

	err = h.updateHandler.Handle(c.Context(), projectcmd.UpdateProjectCommand{
		ID:   id,
		Name: req.Name,
	})
	if err != nil {
		if err.Error() == "project not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Project not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to update project"})
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryproject.GetProjectByIDQuery{ID: id})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get project"})
	}

	return c.Status(fiber.StatusOK).JSON(
		presenter.NewProjectDetailResponseFromView(*view, user.ActiveProjectID),
	)
}

func (h *ProjectHandler) Activate(c fiber.Ctx) error {
	user, err := httpctx.GetUser(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Unauthorized"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid project id"})
	}

	err = h.setActiveProjectHandler.Handle(c.Context(), usercmd.SetActiveProjectCommand{
		UserID:    user.ID,
		ProjectID: id,
	})
	if err != nil {
		switch err.Error() {
		case "project not found":
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Project not found"})
		case "user is not a member of the project":
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "User is not a member of the project"})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to activate project"})
		}
	}

	view, err := h.getByIDHandler.Handle(c.Context(), queryproject.GetProjectByIDQuery{ID: id})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to get project"})
	}

	activeID := id
	return c.Status(fiber.StatusOK).JSON(
		presenter.NewProjectDetailResponseFromView(*view, &activeID),
	)
}

func (h *ProjectHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid project id"})
	}

	if err := h.deleteHandler.Handle(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to delete project"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *ProjectHandler) RemoveMember(c fiber.Ctx) error {
	projectID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid project id"})
	}
	userID, err := uuid.Parse(c.Params("userId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid user id"})
	}

	err = h.removeMemberHandler.Handle(c.Context(), projectcmd.RemoveProjectMemberCommand{
		ProjectID: projectID,
		UserID:    userID,
	})
	if err != nil {
		if err.Error() == "project not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Project not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Failed to remove member"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
