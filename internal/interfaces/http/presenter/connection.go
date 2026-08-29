package presenter

import domainconnection "go-api/internal/domain/connection"

type ConnectionDetailResponse struct {
	ID           string  `json:"id"`
	SourceStepID string  `json:"sourceStepId"`
	TargetStepID string  `json:"targetStepId"`
	Branch       *string `json:"branch,omitempty"`
}

func NewConnectionDetailResponseFromEntity(c domainconnection.Connection) ConnectionDetailResponse {
	return ConnectionDetailResponse{
		ID:           c.ID.String(),
		SourceStepID: c.SourceStepID.String(),
		TargetStepID: c.TargetStepID.String(),
		Branch:       optionalBranchString(c.Branch),
	}
}

func NewConnectionDetailResponseFromView(v domainconnection.ConnectionView) ConnectionDetailResponse {
	return ConnectionDetailResponse{
		ID:           v.ID.String(),
		SourceStepID: v.SourceStepID.String(),
		TargetStepID: v.TargetStepID.String(),
		Branch:       optionalBranchString(v.Branch),
	}
}

func NewConnectionListResponseFromViews(views []domainconnection.ConnectionView) []ConnectionDetailResponse {
	items := make([]ConnectionDetailResponse, 0, len(views))
	for _, v := range views {
		items = append(items, NewConnectionDetailResponseFromView(v))
	}
	return items
}

func optionalBranchString(branch *domainconnection.ConditionBranch) *string {
	if branch == nil {
		return nil
	}
	value := string(*branch)
	return &value
}
