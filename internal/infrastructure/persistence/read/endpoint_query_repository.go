package read

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/paginate"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type endpointRow struct {
	ID             uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         string
	Headers        dbtype.JSONB
	QueryParams    dbtype.JSONB
	Body           dbtype.JSONB
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
	Status         string
	OrganizationID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (endpointRow) TableName() string { return "endpoints" }

var endpointSelectColumns = []string{
	"id", "name", "description", "url", "method", "headers", "query_params", "body",
	"timeout_ms", "retry_on_failure", "retry_count", "retry_delay_ms", "status",
	"organization_id", "created_at", "updated_at",
}

type endpointReadRepository struct {
	db *gorm.DB
}

func NewEndpointReadRepository(db *gorm.DB) domainendpoint.EndpointReadRepository {
	return &endpointReadRepository{db: db}
}

func (r *endpointReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainendpoint.EndpointView, error) {
	var row endpointRow
	err := r.db.WithContext(ctx).
		Select(endpointSelectColumns).
		First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toEndpointView(row)
}

func (r *endpointReadRepository) FindByOrganizationID(
	ctx context.Context,
	organizationID uuid.UUID,
	filter domainendpoint.ListEndpointsFilter,
) ([]domainendpoint.EndpointView, int64, error) {
	query := filter.PaginateQuery
	query.Normalize()
	if query.SortBy == "" {
		query.SortBy = "created_at"
	}
	if query.OrderBy != paginate.OrderByAsc {
		query.OrderBy = paginate.OrderByDesc
	}

	db := r.db.WithContext(ctx).
		Model(&endpointRow{}).
		Where("organization_id = ? AND status <> ?", organizationID, domainendpoint.StatusDeleted)

	if query.Search != "" {
		like := "%" + query.Search + "%"
		db = db.Where("name ILIKE ? OR url ILIKE ?", like, like)
	}
	if len(filter.Methods) > 0 {
		methods := make([]string, 0, len(filter.Methods))
		for _, method := range filter.Methods {
			methods = append(methods, string(method))
		}
		db = db.Where("method IN ?", methods)
	}

	db, total, err := Paginate(db, query)
	if err != nil {
		return nil, 0, err
	}

	var rows []endpointRow
	err = db.Select(endpointSelectColumns).Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	views := make([]domainendpoint.EndpointView, 0, len(rows))
	for _, row := range rows {
		view, err := toEndpointView(row)
		if err != nil {
			return nil, 0, err
		}
		views = append(views, *view)
	}
	return views, total, nil
}

func toEndpointView(row endpointRow) (*domainendpoint.EndpointView, error) {
	headers := map[string]string{}
	if len(row.Headers) > 0 {
		if err := json.Unmarshal(row.Headers, &headers); err != nil {
			return nil, err
		}
	}

	query := map[string]string{}
	if len(row.QueryParams) > 0 {
		if err := json.Unmarshal(row.QueryParams, &query); err != nil {
			return nil, err
		}
	}
	body := map[string]any{}
	if len(row.Body) > 0 {
		if err := json.Unmarshal(row.Body, &body); err != nil {
			return nil, err
		}
	}

	return &domainendpoint.EndpointView{
		ID:             row.ID,
		Name:           row.Name,
		Description:    row.Description,
		URL:            row.URL,
		Method:         domainendpoint.Method(row.Method),
		Headers:        headers,
		Query:          query,
		Body:           body,
		Timeout:        row.Timeout,
		RetryOnFailure: row.RetryOnFailure,
		RetryCount:     row.RetryCount,
		RetryDelay:     row.RetryDelay,
		Status:         domainendpoint.Status(row.Status),
		OrganizationID: row.OrganizationID,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}, nil
}
