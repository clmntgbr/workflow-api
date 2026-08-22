package write

import (
	"encoding/json"
	"time"

	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/httpquery"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

type EndpointModel struct {
	ID             uuid.UUID    `gorm:"column:id;primaryKey"`
	Name           string       `gorm:"column:name"`
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
	Status         string       `gorm:"column:status"`
	ProjectID uuid.UUID    `gorm:"column:project_id"`
	CreatedAt      time.Time    `gorm:"column:created_at"`
	UpdatedAt      time.Time    `gorm:"column:updated_at"`
}

func (EndpointModel) TableName() string {
	return "endpoints"
}

func endpointModelFromDomain(e *domainendpoint.Endpoint) (*EndpointModel, error) {
	headers := e.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	headersRaw, err := json.Marshal(headers)
	if err != nil {
		return nil, err
	}

	query := e.Query
	if query == nil {
		query = httpquery.Empty()
	}
	queryRaw, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}
	body := e.Body
	if body == nil {
		body = map[string]any{}
	}
	bodyRaw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return &EndpointModel{
		ID:             e.ID,
		Name:           e.Name,
		Description:    e.Description,
		URL:            e.URL,
		Method:         string(e.Method),
		Headers:        dbtype.JSONB(headersRaw),
		QueryParams:    dbtype.JSONB(queryRaw),
		Body:           dbtype.JSONB(bodyRaw),
		Timeout:        e.Timeout,
		RetryOnFailure: e.RetryOnFailure,
		RetryCount:     e.RetryCount,
		RetryDelay:     e.RetryDelay,
		Status:         string(e.Status),
		ProjectID: e.ProjectID,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}, nil
}

func endpointDomainFromModel(m *EndpointModel) (*domainendpoint.Endpoint, error) {
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

	return &domainendpoint.Endpoint{
		ID:             m.ID,
		Name:           m.Name,
		Description:    m.Description,
		URL:            m.URL,
		Method:         domainendpoint.Method(m.Method),
		Headers:        headers,
		Query:          query,
		Body:           body,
		Timeout:        m.Timeout,
		RetryOnFailure: m.RetryOnFailure,
		RetryCount:     m.RetryCount,
		RetryDelay:     m.RetryDelay,
		Status:         domainendpoint.Status(m.Status),
		ProjectID: m.ProjectID,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}, nil
}
