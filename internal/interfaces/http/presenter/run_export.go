package presenter

import (
	"time"

	domainrunexport "go-api/internal/domain/runexport"
)

type RunExportAcceptedResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type RunExportResponse struct {
	ID        string     `json:"id"`
	Status    string     `json:"status"`
	From      time.Time  `json:"from"`
	To        time.Time  `json:"to"`
	Error     *string    `json:"error"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func NewRunExportAcceptedResponse(job domainrunexport.RunExport) RunExportAcceptedResponse {
	return RunExportAcceptedResponse{
		ID:     job.ID.String(),
		Status: string(job.Status),
	}
}

func NewRunExportResponseFromView(view domainrunexport.View) RunExportResponse {
	return RunExportResponse{
		ID:        view.ID.String(),
		Status:    string(view.Status),
		From:      view.DateRangeFrom,
		To:        view.DateRangeTo,
		Error:     optionalNonEmptyString(view.Error),
		CreatedAt: view.CreatedAt,
		UpdatedAt: view.UpdatedAt,
	}
}
