package write

import (
	"encoding/json"
	"time"

	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

type StepRunModel struct {
	ID               uuid.UUID    `gorm:"column:id;primaryKey"`
	WorkflowRunID    uuid.UUID    `gorm:"column:workflow_run_id"`
	StepID           uuid.UUID    `gorm:"column:step_id"`
	WorkflowID       uuid.UUID    `gorm:"column:workflow_id"`
	EndpointID       uuid.UUID    `gorm:"column:endpoint_id"`
	OrganizationID   uuid.UUID    `gorm:"column:organization_id"`
	Name             string       `gorm:"column:name"`
	Description      string       `gorm:"column:description"`
	URL              string       `gorm:"column:url"`
	Method           string       `gorm:"column:method"`
	Headers          dbtype.JSONB `gorm:"column:headers"`
	QueryParams      dbtype.JSONB `gorm:"column:query_params"`
	Body             dbtype.JSONB `gorm:"column:body"`
	Timeout          int          `gorm:"column:timeout_ms"`
	RetryOnFailure   bool         `gorm:"column:retry_on_failure"`
	RetryCount       int          `gorm:"column:retry_count"`
	RetryDelay       int          `gorm:"column:retry_delay_ms"`
	StepIndex        string       `gorm:"column:step_index"`
	ExecutionOrder   int          `gorm:"column:execution_order"`
	TreeIndex        int          `gorm:"column:tree_index"`
	PositionX        float64      `gorm:"column:position_x"`
	PositionY        float64      `gorm:"column:position_y"`
	Status           string       `gorm:"column:status"`
	Attempt          int          `gorm:"column:attempt"`
	ResponseSnapshot dbtype.JSONB `gorm:"column:response_snapshot"`
	StartedAt        *time.Time   `gorm:"column:started_at"`
	FinishedAt       *time.Time   `gorm:"column:finished_at"`
	Error            string       `gorm:"column:error"`
	CreatedAt        time.Time    `gorm:"column:created_at"`
	UpdatedAt        time.Time    `gorm:"column:updated_at"`
}

func (StepRunModel) TableName() string {
	return "step_runs"
}

func stepRunModelFromDomain(s *domainsteprun.StepRun) (*StepRunModel, error) {
	headers := s.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	headersRaw, err := json.Marshal(headers)
	if err != nil {
		return nil, err
	}

	query := s.Query
	if query == nil {
		query = map[string]string{}
	}
	queryRaw, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	body := s.Body
	if body == nil {
		body = map[string]any{}
	}
	bodyRaw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	var responseRaw dbtype.JSONB
	if s.ResponseSnapshot != nil {
		encoded, err := json.Marshal(s.ResponseSnapshot)
		if err != nil {
			return nil, err
		}
		responseRaw = dbtype.JSONB(encoded)
	}

	return &StepRunModel{
		ID:               s.ID,
		WorkflowRunID:    s.WorkflowRunID,
		StepID:           s.StepID,
		WorkflowID:       s.WorkflowID,
		EndpointID:       s.EndpointID,
		OrganizationID:   s.OrganizationID,
		Name:             s.Name,
		Description:      s.Description,
		URL:              s.URL,
		Method:           s.Method,
		Headers:          dbtype.JSONB(headersRaw),
		QueryParams:      dbtype.JSONB(queryRaw),
		Body:             dbtype.JSONB(bodyRaw),
		Timeout:          s.Timeout,
		RetryOnFailure:   s.RetryOnFailure,
		RetryCount:       s.RetryCount,
		RetryDelay:       s.RetryDelay,
		StepIndex:        s.Index,
		ExecutionOrder:   s.ExecutionOrder,
		TreeIndex:        s.TreeIndex,
		PositionX:        s.Position.X,
		PositionY:        s.Position.Y,
		Status:           string(s.Status),
		Attempt:          s.Attempt,
		ResponseSnapshot: responseRaw,
		StartedAt:        s.StartedAt,
		FinishedAt:       s.FinishedAt,
		Error:            s.Error,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}, nil
}

func stepRunDomainFromModel(m *StepRunModel) (*domainsteprun.StepRun, error) {
	headers := map[string]string{}
	if len(m.Headers) > 0 {
		if err := json.Unmarshal(m.Headers, &headers); err != nil {
			return nil, err
		}
	}

	query := map[string]string{}
	if len(m.QueryParams) > 0 {
		if err := json.Unmarshal(m.QueryParams, &query); err != nil {
			return nil, err
		}
	}

	body := map[string]any{}
	if len(m.Body) > 0 {
		if err := json.Unmarshal(m.Body, &body); err != nil {
			return nil, err
		}
	}

	var response *domainsteprun.ResponseSnapshot
	if len(m.ResponseSnapshot) > 0 {
		var snapshot domainsteprun.ResponseSnapshot
		if err := json.Unmarshal(m.ResponseSnapshot, &snapshot); err != nil {
			return nil, err
		}
		response = &snapshot
	}

	return &domainsteprun.StepRun{
		ID:             m.ID,
		WorkflowRunID:  m.WorkflowRunID,
		StepID:         m.StepID,
		WorkflowID:     m.WorkflowID,
		EndpointID:     m.EndpointID,
		OrganizationID: m.OrganizationID,
		Name:           m.Name,
		Description:    m.Description,
		URL:            m.URL,
		Method:         m.Method,
		Headers:        headers,
		Query:          query,
		Body:           body,
		Timeout:        m.Timeout,
		RetryOnFailure: m.RetryOnFailure,
		RetryCount:     m.RetryCount,
		RetryDelay:     m.RetryDelay,
		Index:          m.StepIndex,
		ExecutionOrder: m.ExecutionOrder,
		TreeIndex:      m.TreeIndex,
		Position: domainstep.Position{
			X: m.PositionX,
			Y: m.PositionY,
		},
		Status:           domainsteprun.Status(m.Status),
		Attempt:          m.Attempt,
		ResponseSnapshot: response,
		StartedAt:        m.StartedAt,
		FinishedAt:       m.FinishedAt,
		Error:            m.Error,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}, nil
}
