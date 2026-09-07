package write

import (
	"context"
	"errors"
	"strings"
	"time"

	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type stepRunWriteRepository struct {
	db *gorm.DB
}

func NewStepRunWriteRepository(db *gorm.DB) domainsteprun.StepRunWriteRepository {
	return &stepRunWriteRepository{db: db}
}

func (r *stepRunWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *stepRunWriteRepository) Save(ctx context.Context, run *domainsteprun.StepRun) error {
	model, err := stepRunModelFromDomain(run)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Create(model).Error
}

func (r *stepRunWriteRepository) Update(ctx context.Context, run *domainsteprun.StepRun) error {
	model, err := stepRunModelFromDomain(run)
	if err != nil {
		return err
	}
	return DBWithContext(ctx, r.db).Save(model).Error
}

func (r *stepRunWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainsteprun.StepRun, error) {
	var model StepRunModel
	err := DBWithContext(ctx, r.db).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return stepRunDomainFromModel(&model)
}

func (r *stepRunWriteRepository) FindByWorkflowRunID(
	ctx context.Context,
	workflowRunID uuid.UUID,
) ([]*domainsteprun.StepRun, error) {
	var models []StepRunModel
	err := DBWithContext(ctx, r.db).
		Where("workflow_run_id = ?", workflowRunID).
		Order("execution_order ASC, created_at ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	runs := make([]*domainsteprun.StepRun, 0, len(models))
	for i := range models {
		run, err := stepRunDomainFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

func (r *stepRunWriteRepository) ClaimDueWaiting(
	ctx context.Context,
	now time.Time,
	limit int,
) ([]*domainsteprun.StepRun, error) {
	if limit <= 0 {
		limit = 100
	}

	var claimed []*domainsteprun.StepRun
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var models []StepRunModel
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ?", domainsteprun.StatusWaiting).
			Where("resume_at IS NOT NULL AND resume_at <= ?", now.UTC()).
			Order("resume_at ASC").
			Limit(limit).
			Find(&models).Error
		if err != nil {
			return err
		}
		if len(models) == 0 {
			return nil
		}

		claimed = make([]*domainsteprun.StepRun, 0, len(models))
		for i := range models {
			run, err := stepRunDomainFromModel(&models[i])
			if err != nil {
				return err
			}
			claimed = append(claimed, run)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return claimed, nil
}

func (r *stepRunWriteRepository) FindActiveNonDelay(
	ctx context.Context,
	now time.Time,
	pendingMaxAge, grace time.Duration,
	limit int,
) ([]*domainsteprun.StepRun, error) {
	if limit <= 0 {
		limit = 100
	}
	if grace < 0 {
		grace = 0
	}

	nowUTC := now.UTC()
	args := map[string]any{
		"delay_type":         domainstep.TypeDelay,
		"running_status":     domainsteprun.StatusRunning,
		"default_timeout_ms": domainsteprun.DefaultHTTPTimeoutMS,
		"grace_ms":           grace.Milliseconds(),
		"as_of":              nowUTC,
	}

	filters := []string{`(
		status = @running_status
		AND (
			COALESCE(started_at, created_at)
			+ (
				(CASE WHEN timeout_ms <= 0 THEN @default_timeout_ms ELSE timeout_ms END)
				* (CASE WHEN retry_on_failure THEN GREATEST(retry_count, 1) ELSE 1 END)
			) * interval '1 millisecond'
			+ (
				CASE
					WHEN retry_on_failure AND GREATEST(retry_count, 1) > 1 AND retry_delay_ms > 0
					THEN (GREATEST(retry_count, 1) - 1) * retry_delay_ms
					ELSE 0
				END
			) * interval '1 millisecond'
			+ (@grace_ms)::bigint * interval '1 millisecond'
		) <= @as_of
	)`}
	if pendingMaxAge > 0 {
		args["pending_status"] = domainsteprun.StatusPending
		args["pending_cutoff"] = nowUTC.Add(-pendingMaxAge)
		filters = append(filters, `(status = @pending_status AND created_at <= @pending_cutoff)`)
	}

	var models []StepRunModel
	err := DBWithContext(ctx, r.db).
		Where(
			"step_type <> @delay_type AND ("+strings.Join(filters, " OR ")+")",
			args,
		).
		Order("COALESCE(started_at, created_at) ASC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	runs := make([]*domainsteprun.StepRun, 0, len(models))
	for i := range models {
		run, err := stepRunDomainFromModel(&models[i])
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}
