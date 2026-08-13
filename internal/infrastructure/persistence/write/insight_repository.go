package write

import (
	"context"

	domaininsight "go-api/internal/domain/insight"

	"gorm.io/gorm"
)

type insightWriteRepository struct {
	db *gorm.DB
}

func NewInsightWriteRepository(db *gorm.DB) domaininsight.InsightWriteRepository {
	return &insightWriteRepository{db: db}
}

func (r *insightWriteRepository) Save(ctx context.Context, insight *domaininsight.Insight) error {
	return DBWithContext(ctx, r.db).Create(insightModelFromDomain(insight)).Error
}
