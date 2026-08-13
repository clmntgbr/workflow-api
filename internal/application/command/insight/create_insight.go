package insight

import (
	"context"
	"errors"
	"time"

	domaininsight "go-api/internal/domain/insight"

	"github.com/google/uuid"
)

type CreateInsightCommand struct {
	StepRunID         uuid.UUID
	StartTime         time.Time
	EndTime           time.Time
	QueueTime         time.Duration
	DNSLookupDuration time.Duration
	TCPConnectionTime time.Duration
	TLSHandshakeTime  time.Duration
	TTFB              time.Duration
	Duration          time.Duration
	StatusCode        int
	HasStatusCode     bool
	ResponseSize      int64
	RequestSize       int64
	AttemptNumber     int
	TotalAttempts     int
	ErrorMessage      string
	ErrorType         string
}

type CreateInsightHandler struct {
	repo domaininsight.InsightWriteRepository
}

func NewCreateInsightHandler(repo domaininsight.InsightWriteRepository) *CreateInsightHandler {
	return &CreateInsightHandler{repo: repo}
}

func (h *CreateInsightHandler) Handle(ctx context.Context, cmd CreateInsightCommand) (*domaininsight.Insight, error) {
	if cmd.StepRunID == uuid.Nil {
		return nil, errors.New("stepRunId is required")
	}
	if cmd.AttemptNumber <= 0 {
		return nil, errors.New("attemptNumber must be positive")
	}

	insight := domaininsight.NewInsight(domaininsight.NewInsightParams{
		StepRunID:         cmd.StepRunID,
		StartTime:         cmd.StartTime,
		EndTime:           cmd.EndTime,
		QueueTime:         cmd.QueueTime,
		DNSLookupDuration: cmd.DNSLookupDuration,
		TCPConnectionTime: cmd.TCPConnectionTime,
		TLSHandshakeTime:  cmd.TLSHandshakeTime,
		TTFB:              cmd.TTFB,
		Duration:          cmd.Duration,
		StatusCode:        cmd.StatusCode,
		HasStatusCode:     cmd.HasStatusCode,
		ResponseSize:      cmd.ResponseSize,
		RequestSize:       cmd.RequestSize,
		AttemptNumber:     cmd.AttemptNumber,
		TotalAttempts:     cmd.TotalAttempts,
		ErrorMessage:      cmd.ErrorMessage,
		ErrorType:         cmd.ErrorType,
	})

	if err := h.repo.Save(ctx, insight); err != nil {
		return nil, err
	}
	return insight, nil
}
