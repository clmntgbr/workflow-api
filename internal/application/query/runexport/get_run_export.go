package runexport

import (
	"context"
	"errors"

	domainrunexport "go-api/internal/domain/runexport"

	"github.com/google/uuid"
)

type GetRunExportByIDQuery struct {
	ID uuid.UUID
}

type GetRunExportByIDHandler struct {
	readRepo domainrunexport.ReadRepository
}

func NewGetRunExportByIDHandler(readRepo domainrunexport.ReadRepository) *GetRunExportByIDHandler {
	return &GetRunExportByIDHandler{readRepo: readRepo}
}

func (h *GetRunExportByIDHandler) Handle(
	ctx context.Context,
	q GetRunExportByIDQuery,
) (*domainrunexport.View, error) {
	if q.ID == uuid.Nil {
		return nil, errors.New("id is required")
	}
	view, err := h.readRepo.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get run export")
	}
	if view == nil {
		return nil, errors.New("run export not found")
	}
	return view, nil
}
