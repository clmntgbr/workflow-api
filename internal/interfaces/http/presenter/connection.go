package presenter

import domainconnection "go-api/internal/domain/connection"

type ConnectionDetailResponse struct {
	ID           string `json:"id"`
	WorkflowID   string `json:"workflowId"`
	SourceStepID string `json:"sourceStepId"`
	TargetStepID string `json:"targetStepId"`
}

func NewConnectionDetailResponseFromEntity(c domainconnection.Connection) ConnectionDetailResponse {
	return ConnectionDetailResponse{
		ID:           c.ID.String(),
		WorkflowID:   c.WorkflowID.String(),
		SourceStepID: c.SourceStepID.String(),
		TargetStepID: c.TargetStepID.String(),
	}
}

func NewConnectionDetailResponseFromView(v domainconnection.ConnectionView) ConnectionDetailResponse {
	return ConnectionDetailResponse{
		ID:           v.ID.String(),
		WorkflowID:   v.WorkflowID.String(),
		SourceStepID: v.SourceStepID.String(),
		TargetStepID: v.TargetStepID.String(),
	}
}

func NewConnectionListResponseFromViews(views []domainconnection.ConnectionView) []ConnectionDetailResponse {
	items := make([]ConnectionDetailResponse, 0, len(views))
	for _, v := range views {
		items = append(items, NewConnectionDetailResponseFromView(v))
	}
	return items
}
