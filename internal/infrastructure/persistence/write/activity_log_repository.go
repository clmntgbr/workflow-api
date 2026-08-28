package write

import (
	"context"
	"errors"
	"time"

	domainactivitylog "go-api/internal/domain/activitylog"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type activityLogWriteRepository struct {
	db *gorm.DB
}

func NewActivityLogWriteRepository(db *gorm.DB) domainactivitylog.WriteRepository {
	return &activityLogWriteRepository{db: db}
}

func (r *activityLogWriteRepository) Save(ctx context.Context, entry *domainactivitylog.Entry) error {
	if entry == nil {
		return errors.New("activity log entry is required")
	}
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	if entry.Level == "" {
		entry.Level = domainactivitylog.LevelInfo
	}

	model, err := activityLogModelFromDomain(entry)
	if err != nil {
		return err
	}

	return DBWithContext(ctx, r.db).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "source_event_id"}},
			DoNothing: true,
		}).
		Create(model).Error
}
