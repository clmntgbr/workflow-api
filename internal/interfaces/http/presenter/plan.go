package presenter

import (
	"time"

	domainplan "go-api/internal/domain/plan"
	domainquota "go-api/internal/domain/quota"
)

type QuotaResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	MaxProjectMembers       int `json:"maxProjectMembers"`
	MaxProjects             int `json:"maxProjects"`
	MaxWorkflows            int `json:"maxWorkflows"`
	MaxStepsPerWorkflow     int `json:"maxStepsPerWorkflow"`
	MaxEndpoints            int `json:"maxEndpoints"`
	MaxVariablesPerWorkflow int `json:"maxVariablesPerWorkflow"`

	MaxWorkflowRunsPerMonth    int `json:"maxWorkflowRunsPerMonth"`
	MaxConcurrentRuns          int `json:"maxConcurrentRuns"`
	MinScheduleIntervalMinutes int `json:"minScheduleIntervalMinutes"`

	RunHistoryRetentionDays int `json:"runHistoryRetentionDays"`

	MaxStepTimeoutSeconds int `json:"maxStepTimeoutSeconds"`
	MaxRetryCountPerStep  int `json:"maxRetryCountPerStep"`
	MaxRequestBodySizeKB  int `json:"maxRequestBodySizeKb"`
	MaxResponseBodySizeKB int `json:"maxResponseBodySizeKb"`

	AllowsOpenAPIImport bool `json:"allowsOpenApiImport"`
	AllowsInsights      bool `json:"allowsInsights"`
	AllowsDataExport    bool `json:"allowsDataExport"`
	ExecutorPriority    int  `json:"executorPriority"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PlanResponse struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Description     *string       `json:"description"`
	Slug            string        `json:"slug"`
	StripePriceID   *string       `json:"stripePriceId"`
	IsActive        bool          `json:"isActive"`
	BillingInterval string        `json:"billingInterval"`
	Price           float64       `json:"price"`
	Currency        string        `json:"currency"`
	QuotaID         string        `json:"quotaId"`
	Quota           QuotaResponse `json:"quota"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

func NewQuotaResponseFromView(view domainquota.QuotaView) QuotaResponse {
	return QuotaResponse{
		ID:                         view.ID.String(),
		Name:                       view.Name,
		MaxProjectMembers:          view.MaxProjectMembers,
		MaxProjects:                view.MaxProjects,
		MaxWorkflows:               view.MaxWorkflows,
		MaxStepsPerWorkflow:        view.MaxStepsPerWorkflow,
		MaxEndpoints:               view.MaxEndpoints,
		MaxVariablesPerWorkflow:    view.MaxVariablesPerWorkflow,
		MaxWorkflowRunsPerMonth:    view.MaxWorkflowRunsPerMonth,
		MaxConcurrentRuns:          view.MaxConcurrentRuns,
		MinScheduleIntervalMinutes: view.MinScheduleIntervalMinutes,
		RunHistoryRetentionDays:    view.RunHistoryRetentionDays,
		MaxStepTimeoutSeconds:      view.MaxStepTimeoutSeconds,
		MaxRetryCountPerStep:       view.MaxRetryCountPerStep,
		MaxRequestBodySizeKB:       view.MaxRequestBodySizeKB,
		MaxResponseBodySizeKB:      view.MaxResponseBodySizeKB,
		AllowsOpenAPIImport:        view.AllowsOpenAPIImport,
		AllowsInsights:             view.AllowsInsights,
		AllowsDataExport:           view.AllowsDataExport,
		ExecutorPriority:           view.ExecutorPriority,
		CreatedAt:                  view.CreatedAt,
		UpdatedAt:                  view.UpdatedAt,
	}
}

func NewPlanResponseFromView(view domainplan.PlanView) PlanResponse {
	resp := PlanResponse{
		ID:              view.ID.String(),
		Name:            view.Name,
		Description:     optionalNonEmptyString(view.Description),
		Slug:            view.Slug,
		StripePriceID:   optionalNonEmptyString(view.StripePriceID),
		IsActive:        view.IsActive,
		BillingInterval: string(view.BillingInterval),
		Price:           view.Price,
		Currency:        string(view.Currency),
		QuotaID:         view.QuotaID.String(),
		CreatedAt:       view.CreatedAt,
		UpdatedAt:       view.UpdatedAt,
	}
	if view.Quota != nil {
		resp.Quota = NewQuotaResponseFromView(*view.Quota)
	}
	return resp
}

func NewPlanListResponseFromViews(views []domainplan.PlanView) []PlanResponse {
	items := make([]PlanResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewPlanResponseFromView(view))
	}
	return items
}
