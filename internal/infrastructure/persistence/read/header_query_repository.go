package read

import (
	"context"

	domainheader "go-api/internal/domain/header"

	"gorm.io/gorm"
)

// Named parameters (@name) are mandatory here: the Postgres jsonb key-exists
// operator is also `?`, which GORM would rewrite as a bind placeholder.
const headerKeysSQL = `
WITH all_headers AS (
	SELECT k AS header_key
	FROM endpoints e, LATERAL jsonb_object_keys(e.headers) AS k
	WHERE e.project_id = @projectId
	UNION ALL
	SELECT k
	FROM steps s, LATERAL jsonb_object_keys(s.headers) AS k
	WHERE s.project_id = @projectId
)
SELECT header_key, COUNT(*) AS count
FROM all_headers
WHERE (@search = '' OR header_key ILIKE @like)
GROUP BY header_key`

const headerValuesSQL = `
WITH all_headers AS (
	SELECT h.key AS header_key, h.value AS header_value
	FROM endpoints e, LATERAL jsonb_each_text(e.headers) AS h
	WHERE e.project_id = @projectId
	UNION ALL
	SELECT h.key, h.value
	FROM steps s, LATERAL jsonb_each_text(s.headers) AS h
	WHERE s.project_id = @projectId
)
SELECT header_key, header_value, COUNT(*) AS count
FROM all_headers
WHERE (@key = '' OR header_key = @key)
  AND (@search = '' OR header_value ILIKE @like)
GROUP BY header_key, header_value`

type headerKeyRow struct {
	HeaderKey string `gorm:"column:header_key"`
	Count     int64  `gorm:"column:count"`
}

type headerValueRow struct {
	HeaderKey   string `gorm:"column:header_key"`
	HeaderValue string `gorm:"column:header_value"`
	Count       int64  `gorm:"column:count"`
}

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

	args := map[string]any{
		"projectId": filter.ProjectID,
		"search":    filter.Search,
		"like":      "%" + filter.Search + "%",
		"limit":     query.Limit,
		"offset":    query.Offset(),
	}

	total, err := r.countGroups(ctx, headerKeysSQL, args)
	if err != nil {
		return nil, 0, err
	}

	var rows []headerKeyRow
	pageSQL := headerKeysSQL + `
ORDER BY count DESC, header_key ASC
LIMIT @limit OFFSET @offset`
	if err := r.db.WithContext(ctx).Raw(pageSQL, args).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	suggestions := make([]domainheader.HeaderSuggestion, 0, len(rows))
	for _, row := range rows {
		suggestions = append(suggestions, domainheader.HeaderSuggestion{
			Key:   row.HeaderKey,
			Count: row.Count,
		})
	}
	return suggestions, total, nil
}

func (r *headerReadRepository) FindValuesByKey(
	ctx context.Context,
	filter domainheader.HeaderValueSuggestionFilter,
) ([]domainheader.HeaderValueSuggestion, int64, error) {
	query := filter.Paginate
	query.Normalize()

	args := map[string]any{
		"projectId": filter.ProjectID,
		"key":       filter.Key,
		"search":    filter.Search,
		"like":      "%" + filter.Search + "%",
		"limit":     query.Limit,
		"offset":    query.Offset(),
	}

	total, err := r.countGroups(ctx, headerValuesSQL, args)
	if err != nil {
		return nil, 0, err
	}

	var rows []headerValueRow
	pageSQL := headerValuesSQL + `
ORDER BY count DESC, header_key ASC, header_value ASC
LIMIT @limit OFFSET @offset`
	if err := r.db.WithContext(ctx).Raw(pageSQL, args).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	suggestions := make([]domainheader.HeaderValueSuggestion, 0, len(rows))
	for _, row := range rows {
		suggestions = append(suggestions, domainheader.HeaderValueSuggestion{
			Key:   row.HeaderKey,
			Value: row.HeaderValue,
			Count: row.Count,
		})
	}
	return suggestions, total, nil
}

func (r *headerReadRepository) countGroups(
	ctx context.Context,
	baseSQL string,
	args map[string]any,
) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Raw("SELECT COUNT(*) FROM ("+baseSQL+") AS counted", args).
		Scan(&total).Error
	return total, err
}
