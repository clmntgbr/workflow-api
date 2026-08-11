package processed

import (
	"context"
	"errors"
	"time"

	"go-api/internal/domain/port"
	"go-api/internal/infrastructure/persistence/write"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) port.ProcessedEventRepository {
	return &repository{db: db}
}

func (r *repository) MarkProcessed(ctx context.Context, eventID, handlerName string) (bool, error) {
	id, err := uuid.Parse(eventID)
	if err != nil {
		return false, err
	}

	row := ProcessedEvent{
		EventID:     id,
		HandlerName: handlerName,
		ProcessedAt: time.Now().UTC(),
	}

	err = write.DBWithContext(ctx, r.db).Create(&row).Error
	if err == nil {
		return true, nil
	}
	if isUniqueViolation(err) {
		return false, nil
	}
	return false, err
}

func (r *repository) UnmarkProcessed(ctx context.Context, eventID, handlerName string) error {
	id, err := uuid.Parse(eventID)
	if err != nil {
		return err
	}
	return write.DBWithContext(ctx, r.db).
		Where("event_id = ? AND handler_name = ?", id, handlerName).
		Delete(&ProcessedEvent{}).Error
}

func isUniqueViolation(err error) bool {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
