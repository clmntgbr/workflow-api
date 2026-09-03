package read

import (
	"context"
	"fmt"

	domainheader "go-api/internal/domain/header"

	"gorm.io/gorm"
)

type headerReadRepository struct {
	db *gorm.DB
}

func NewHeaderReadRepository(db *gorm.DB) domainheader.ReadRepository {
	return &headerReadRepository{db: db}
}

func (r *headerReadRepository) FindSuggestions(
	ctx context.Context,
	filter domainheader.HeaderSuggestionFilter,
) ([]domainheader.HeaderSuggestion, int64, error) {
	query := filter.Paginate
	query.Normalize()

	// CTE that combines headers from endpoints and steps
	unionQuery := `
	WITH all_headers AS (
		SELECT jsonb_object_keys(headers) AS header_key
		FROM endpoints
		WHERE project_id = ?
		UNION ALL
		SELECT jsonb_object_keys(headers) AS header_key
		FROM steps
		WHERE project_id = ?
	)
	SELECT header_key, COUNT(*) as count
	FROM all_headers
	`

	args := []interface{}{filter.ProjectID, filter.ProjectID}

	// Add search filter if provided
	if filter.Search != "" {
		unionQuery += " WHERE header_key ILIKE ?"
		args = append(args, "%"+filter.Search+"%")
	}

	unionQuery += " GROUP BY header_key ORDER BY count DESC, header_key ASC"

	// Count total before pagination
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS counted", unionQuery)
	var total int64
	if err := r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Add pagination
	unionQuery += " LIMIT ? OFFSET ?"
	args = append(args, query.Limit, query.Offset())

	// Execute query
	type result struct {
		HeaderKey string `gorm:"column:header_key"`
		Count     int64  `gorm:"column:count"`
	}

	var results []result
	if err := r.db.WithContext(ctx).Raw(unionQuery, args...).Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	// Map to domain
	suggestions := make([]domainheader.HeaderSuggestion, len(results))
	for i, r := range results {
		suggestions[i] = domainheader.HeaderSuggestion{
			Key:   r.HeaderKey,
			Count: r.Count,
		}
	}

	return suggestions, total, nil
}

func (r *headerReadRepository) FindValuesByKey(
	ctx context.Context,
	filter domainheader.HeaderValueSuggestionFilter,
) ([]domainheader.HeaderValueSuggestion, int64, error) {
	query := filter.Paginate
	query.Normalize()

	var unionQuery string
	var args []interface{}

	if filter.Key != "" {
		// Search for a specific key
		unionQuery = `
		WITH all_values AS (
			SELECT ? AS header_key, headers->>? AS header_value
			FROM endpoints
			WHERE project_id = ? AND headers ? ?
			UNION ALL
			SELECT ? AS header_key, headers->>? AS header_value
			FROM steps
			WHERE project_id = ? AND headers ? ?
		)
		SELECT header_key, header_value, COUNT(*) as count
		FROM all_values
		WHERE header_value IS NOT NULL
		`
		args = []interface{}{
			filter.Key, filter.Key, filter.ProjectID, filter.Key,
			filter.Key, filter.Key, filter.ProjectID, filter.Key,
		}
	} else {
		// Search all keys
		unionQuery = `
		WITH all_values AS (
			SELECT 
				jsonb_object_keys(headers) AS header_key,
				headers->>jsonb_object_keys(headers) AS header_value
			FROM endpoints
			WHERE project_id = ?
			UNION ALL
			SELECT 
				jsonb_object_keys(headers) AS header_key,
				headers->>jsonb_object_keys(headers) AS header_value
			FROM steps
			WHERE project_id = ?
		)
		SELECT header_key, header_value, COUNT(*) as count
		FROM all_values
		WHERE header_value IS NOT NULL
		`
		args = []interface{}{filter.ProjectID, filter.ProjectID}
	}

	// Add search filter if provided
	if filter.Search != "" {
		unionQuery += " AND header_value ILIKE ?"
		args = append(args, "%"+filter.Search+"%")
	}

	unionQuery += " GROUP BY header_key, header_value ORDER BY count DESC, header_key ASC, header_value ASC"

	// Count total before pagination
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS counted", unionQuery)
	var total int64
	if err := r.db.WithContext(ctx).Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	// Add pagination
	unionQuery += " LIMIT ? OFFSET ?"
	args = append(args, query.Limit, query.Offset())

	// Execute query
	type result struct {
		HeaderKey   string `gorm:"column:header_key"`
		HeaderValue string `gorm:"column:header_value"`
		Count       int64  `gorm:"column:count"`
	}

	var results []result
	if err := r.db.WithContext(ctx).Raw(unionQuery, args...).Scan(&results).Error; err != nil {
		return nil, 0, err
	}

	// Map to domain
	suggestions := make([]domainheader.HeaderValueSuggestion, len(results))
	for i, r := range results {
		suggestions[i] = domainheader.HeaderValueSuggestion{
			Key:   r.HeaderKey,
			Value: r.HeaderValue,
			Count: r.Count,
		}
	}

	return suggestions, total, nil
}
