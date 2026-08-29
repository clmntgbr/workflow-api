package write

import (
	"encoding/json"
	"time"

	"go-api/internal/domain/httpquery"
	domainstep "go-api/internal/domain/step"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

type StepModel struct {
	ID                   uuid.UUID  `gorm:"column:id;primaryKey"`
	WorkflowID           uuid.UUID  `gorm:"column:workflow_id"`
	EndpointID           *uuid.UUID `gorm:"column:endpoint_id"`
	ProjectID            uuid.UUID  `gorm:"column:project_id"`
	Type                 string     `gorm:"column:type"`
	DelayDurationSeconds *int       `gorm:"column:delay_duration_seconds"`
	Expression           *string    `gorm:"column:expression"`
	Name                 string     `gorm:"column:name"`
	Description    string       `gorm:"column:description"`
	URL            string       `gorm:"column:url"`
	Method         string       `gorm:"column:method"`
	Headers        dbtype.JSONB `gorm:"column:headers"`
	QueryParams    dbtype.JSONB `gorm:"column:query_params"`
	Body           dbtype.JSONB `gorm:"column:body"`
	Timeout        int          `gorm:"column:timeout_ms"`
	RetryOnFailure bool         `gorm:"column:retry_on_failure"`
	RetryCount     int          `gorm:"column:retry_count"`
	RetryDelay     int          `gorm:"column:retry_delay_ms"`
	StepIndex      string       `gorm:"column:step_index"`
	ExecutionOrder int          `gorm:"column:execution_order"`
	TreeIndex      int          `gorm:"column:tree_index"`
	PositionX      float64      `gorm:"column:position_x"`
	PositionY      float64      `gorm:"column:position_y"`
	Status         string       `gorm:"column:status"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
}

func (StepModel) TableName() string {
	return "steps"
}

func stepModelFromDomain(s *domainstep.Step) (*StepModel, error) {
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
		query = httpquery.Empty()
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

	stepType := string(s.Type)
	if stepType == "" {
		stepType = string(domainstep.TypeHTTP)
	}

	return &StepModel{
		ID:                   s.ID,
		WorkflowID:           s.WorkflowID,
		EndpointID:           s.EndpointID,
		ProjectID:            s.ProjectID,
		Type:                 stepType,
		DelayDurationSeconds: intPtrOrNil(s.DelayDurationSeconds),
		Expression:           s.Expression,
		Name:                 s.Name,
		Description:    s.Description,
		URL:            s.URL,
		Method:         s.Method,
		Headers:        dbtype.JSONB(headersRaw),
		QueryParams:    dbtype.JSONB(queryRaw),
		Body:           dbtype.JSONB(bodyRaw),
		Timeout:        s.Timeout,
		RetryOnFailure: s.RetryOnFailure,
		RetryCount:     s.RetryCount,
		RetryDelay:     s.RetryDelay,
		StepIndex:      s.Index,
		ExecutionOrder: s.ExecutionOrder,
		TreeIndex:      s.TreeIndex,
		PositionX:      s.Position.X,
		PositionY:      s.Position.Y,
		Status:         string(s.Status),
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}, nil
}

func stepDomainFromModel(m *StepModel) (*domainstep.Step, error) {
	headers := map[string]string{}
	if len(m.Headers) > 0 {
		if err := json.Unmarshal(m.Headers, &headers); err != nil {
			return nil, err
		}
	}

	query := httpquery.Empty()
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

	stepType := domainstep.Type(m.Type)
	if stepType == "" {
		stepType = domainstep.TypeHTTP
	}

	return &domainstep.Step{
		ID:                   m.ID,
		WorkflowID:           m.WorkflowID,
		EndpointID:           m.EndpointID,
		ProjectID:            m.ProjectID,
		Type:                 stepType,
		DelayDurationSeconds: intValueOrZero(m.DelayDurationSeconds),
		Expression:           m.Expression,
		Name:                 m.Name,
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
		Status:    domainstep.Status(m.Status),
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}, nil
}

func intPtrOrNil(value int) *int {
	if value == 0 {
		return nil
	}
	return &value
}

func intValueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
