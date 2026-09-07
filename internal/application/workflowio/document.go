package workflowio

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	domainassertion "go-api/internal/domain/assertion"
	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/httpquery"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

const FormatVersion = 1

var (
	ErrUnsupportedFormat = errors.New("unsupported workflow export format")
	ErrInvalidDocument   = errors.New("invalid workflow document")
)

type Document struct {
	FormatVersion int          `json:"formatVersion" validate:"required,eq=1"`
	ExportedAt    time.Time    `json:"exportedAt" validate:"omitempty"`
	Workflow      Workflow     `json:"workflow" validate:"required"`
	Steps         []Step       `json:"steps" validate:"dive"`
	Connections   []Connection `json:"connections" validate:"dive"`
	Variables     []Variable   `json:"variables" validate:"dive"`
	Assertions    []Assertion  `json:"assertions" validate:"dive"`
}

type Workflow struct {
	Name                  string     `json:"name" validate:"required,min=1,max=255"`
	Description           string     `json:"description" validate:"omitempty,max=2000"`
	ScheduleType          string     `json:"scheduleType" validate:"omitempty,schedule_type"`
	ScheduleIntervalValue int        `json:"scheduleIntervalValue" validate:"omitempty,min=0"`
	ScheduleIntervalUnit  string     `json:"scheduleIntervalUnit" validate:"omitempty,schedule_unit"`
	ScheduleAt            *time.Time `json:"scheduleAt" validate:"omitempty"`
	ScheduleTimezone      string     `json:"scheduleTimezone" validate:"omitempty,max=64"`
	Concurrency           int        `json:"concurrency" validate:"omitempty,min=1,max=100"`
	NotificationsEnabled  bool       `json:"notificationsEnabled"`
	NotifyOnSuccess       bool       `json:"notifyOnSuccess"`
	NotifyOnFailure       bool       `json:"notifyOnFailure"`
	NotifyOnCancel        bool       `json:"notifyOnCancel"`
}

type Step struct {
	Ref                  string            `json:"ref" validate:"required,min=1,max=64"`
	Type                 string            `json:"type" validate:"required,oneof=http delay condition"`
	Name                 string            `json:"name" validate:"omitempty,max=255"`
	Description          string            `json:"description,omitempty" validate:"omitempty,max=2000"`
	URL                  string            `json:"url,omitempty" validate:"omitempty,max=2048"`
	Method               string            `json:"method,omitempty" validate:"omitempty,http_method"`
	Headers              map[string]string `json:"headers,omitempty"`
	Query                httpquery.Params  `json:"query,omitempty"`
	Body                 map[string]any    `json:"body,omitempty"`
	Timeout              int               `json:"timeout,omitempty" validate:"omitempty,min=30000,max=300000"`
	RetryOnFailure       bool              `json:"retryOnFailure,omitempty"`
	RetryCount           int               `json:"retryCount,omitempty" validate:"omitempty,min=0,max=10"`
	RetryDelay           int               `json:"retryDelay,omitempty" validate:"omitempty,min=10000,max=60000"`
	Expression           string            `json:"expression,omitempty" validate:"omitempty,min=1"`
	DelayDurationSeconds int               `json:"delayDurationSeconds,omitempty" validate:"omitempty,min=1"`
	Position             Position          `json:"position"`
}

type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Connection struct {
	SourceRef string  `json:"sourceRef" validate:"required,min=1,max=64"`
	TargetRef string  `json:"targetRef" validate:"required,min=1,max=64"`
	Branch    *string `json:"branch" validate:"omitempty,oneof=true false"`
}

type Variable struct {
	Ref         string `json:"ref" validate:"required,min=1,max=64"`
	Kind        string `json:"kind" validate:"required,oneof=extracted static"`
	StepRef     string `json:"stepRef,omitempty" validate:"omitempty,min=1,max=64"`
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Key         string `json:"key" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty" validate:"omitempty,max=255"`
	Path        string `json:"path,omitempty" validate:"omitempty,max=255"`
	Value       any    `json:"value,omitempty"`
}

type Assertion struct {
	StepRef       string `json:"stepRef" validate:"required,min=1,max=64"`
	Description   string `json:"description,omitempty" validate:"omitempty,max=255"`
	Source        string `json:"source" validate:"required,oneof=status header body"`
	Path          string `json:"path,omitempty" validate:"omitempty,max=255"`
	Operator      string `json:"operator" validate:"required,oneof=equals not_equals not_null is_null contains greater_than less_than matches_regex is_string is_number is_boolean is_array is_object"`
	ExpectedValue string `json:"expectedValue,omitempty"`
}

func Invalid(msg string) error {
	return fmt.Errorf("%w: %s", ErrInvalidDocument, msg)
}

func BuildDocument(
	workflow domainworkflow.WorkflowView,
	steps []domainstep.StepView,
	connections []domainconnection.ConnectionView,
	variables []domainvariable.VariableView,
	assertions []domainassertion.AssertionView,
	exportedAt time.Time,
) (Document, error) {
	sortedSteps := append([]domainstep.StepView(nil), steps...)
	sort.SliceStable(sortedSteps, func(i, j int) bool {
		if sortedSteps[i].ExecutionOrder != sortedSteps[j].ExecutionOrder {
			return sortedSteps[i].ExecutionOrder < sortedSteps[j].ExecutionOrder
		}
		if !sortedSteps[i].CreatedAt.Equal(sortedSteps[j].CreatedAt) {
			return sortedSteps[i].CreatedAt.Before(sortedSteps[j].CreatedAt)
		}
		return sortedSteps[i].ID.String() < sortedSteps[j].ID.String()
	})

	stepRefByID := make(map[uuid.UUID]string, len(sortedSteps))
	outSteps := make([]Step, 0, len(sortedSteps))
	for i, step := range sortedSteps {
		ref := fmt.Sprintf("step-%d", i+1)
		stepRefByID[step.ID] = ref
		outSteps = append(outSteps, exportStep(ref, step))
	}

	outConnections := make([]Connection, 0, len(connections))
	for _, conn := range connections {
		sourceRef, ok := stepRefByID[conn.SourceStepID]
		if !ok {
			return Document{}, Invalid("connection source step is missing from export")
		}
		targetRef, ok := stepRefByID[conn.TargetStepID]
		if !ok {
			return Document{}, Invalid("connection target step is missing from export")
		}
		var branch *string
		if conn.Branch != nil {
			value := string(*conn.Branch)
			branch = &value
		}
		outConnections = append(outConnections, Connection{
			SourceRef: sourceRef,
			TargetRef: targetRef,
			Branch:    branch,
		})
	}

	sortedVars := append([]domainvariable.VariableView(nil), variables...)
	sort.SliceStable(sortedVars, func(i, j int) bool {
		if !sortedVars[i].CreatedAt.Equal(sortedVars[j].CreatedAt) {
			return sortedVars[i].CreatedAt.Before(sortedVars[j].CreatedAt)
		}
		return sortedVars[i].Key < sortedVars[j].Key
	})
	outVars := make([]Variable, 0, len(sortedVars))
	for i, variable := range sortedVars {
		item, err := exportVariable(fmt.Sprintf("var-%d", i+1), variable, stepRefByID)
		if err != nil {
			return Document{}, err
		}
		outVars = append(outVars, item)
	}

	sortedAssertions := append([]domainassertion.AssertionView(nil), assertions...)
	sort.SliceStable(sortedAssertions, func(i, j int) bool {
		if !sortedAssertions[i].CreatedAt.Equal(sortedAssertions[j].CreatedAt) {
			return sortedAssertions[i].CreatedAt.Before(sortedAssertions[j].CreatedAt)
		}
		return sortedAssertions[i].ID.String() < sortedAssertions[j].ID.String()
	})
	outAssertions := make([]Assertion, 0, len(sortedAssertions))
	for _, assertion := range sortedAssertions {
		stepRef, ok := stepRefByID[assertion.StepID]
		if !ok {
			return Document{}, Invalid("assertion step is missing from export")
		}
		outAssertions = append(outAssertions, Assertion{
			StepRef:       stepRef,
			Description:   assertion.Description,
			Source:        string(assertion.Source),
			Path:          assertion.Path,
			Operator:      string(assertion.Operator),
			ExpectedValue: assertion.ExpectedValue,
		})
	}

	return Document{
		FormatVersion: FormatVersion,
		ExportedAt:    exportedAt.UTC(),
		Workflow: Workflow{
			Name:                  workflow.Name,
			Description:           workflow.Description,
			ScheduleType:          string(workflow.ScheduleType),
			ScheduleIntervalValue: workflow.ScheduleIntervalValue,
			ScheduleIntervalUnit:  string(workflow.ScheduleIntervalUnit),
			ScheduleAt:            workflow.ScheduleAt,
			ScheduleTimezone:      workflow.ScheduleTimezone,
			Concurrency:           workflow.Concurrency,
			NotificationsEnabled:  workflow.NotificationsEnabled,
			NotifyOnSuccess:       workflow.NotifyOnSuccess,
			NotifyOnFailure:       workflow.NotifyOnFailure,
			NotifyOnCancel:        workflow.NotifyOnCancel,
		},
		Steps:       outSteps,
		Connections: outConnections,
		Variables:   outVars,
		Assertions:  outAssertions,
	}, nil
}

func exportStep(ref string, step domainstep.StepView) Step {
	item := Step{
		Ref:         ref,
		Type:        string(step.Type),
		Name:        step.Name,
		Description: step.Description,
		Position: Position{
			X: step.Position.X,
			Y: step.Position.Y,
		},
	}
	switch step.Type {
	case domainstep.TypeHTTP:
		item.URL = step.URL
		item.Method = step.Method
		item.Headers = step.Headers
		item.Query = step.Query
		item.Body = step.Body
		item.Timeout = step.Timeout
		item.RetryOnFailure = step.RetryOnFailure
		item.RetryCount = step.RetryCount
		item.RetryDelay = step.RetryDelay
	case domainstep.TypeDelay:
		item.DelayDurationSeconds = step.DelayDurationSeconds
	case domainstep.TypeCondition:
		if step.Expression != nil {
			item.Expression = *step.Expression
		}
	}
	return item
}

func exportVariable(
	ref string,
	variable domainvariable.VariableView,
	stepRefByID map[uuid.UUID]string,
) (Variable, error) {
	item := Variable{
		Ref:         ref,
		Kind:        string(variable.Kind),
		Name:        variable.Name,
		Key:         variable.Key,
		Description: variable.Description,
	}
	switch variable.Kind {
	case domainvariable.KindExtracted:
		if variable.StepID == nil {
			return Variable{}, Invalid("extracted variable is missing a step")
		}
		stepRef, ok := stepRefByID[*variable.StepID]
		if !ok {
			return Variable{}, Invalid("extracted variable step is missing from export")
		}
		item.StepRef = stepRef
		item.Path = variable.Path
	case domainvariable.KindStatic:
		item.Value = variable.Value
	}
	return item, nil
}

func (d Document) EnsureFormatVersion() error {
	if d.FormatVersion != FormatVersion {
		return fmt.Errorf("%w: got %d want %d", ErrUnsupportedFormat, d.FormatVersion, FormatVersion)
	}
	return nil
}

func (d Document) ValidateShape() error {
	if err := d.EnsureFormatVersion(); err != nil {
		return err
	}
	if strings.TrimSpace(d.Workflow.Name) == "" {
		return Invalid("workflow name is required")
	}

	stepRefs := make(map[string]string, len(d.Steps))
	for _, step := range d.Steps {
		ref := strings.TrimSpace(step.Ref)
		if ref == "" {
			return Invalid("step ref is required")
		}
		if _, exists := stepRefs[ref]; exists {
			return Invalid("duplicate step ref " + ref)
		}
		stepType, err := domainstep.ParseType(step.Type)
		if err != nil {
			return Invalid("invalid step type")
		}
		stepRefs[ref] = step.Type
		switch stepType {
		case domainstep.TypeHTTP:
			if strings.TrimSpace(step.URL) == "" {
				return Invalid("http step url is required")
			}
			if !domainstep.ValidHTTPMethod(step.Method) {
				return Invalid("http step method is invalid")
			}
			if step.Timeout < 30000 || step.Timeout > 300000 {
				return Invalid("http step timeout is invalid")
			}
			if step.RetryCount < 0 || step.RetryCount > 10 {
				return Invalid("http step retryCount is invalid")
			}
			if step.RetryDelay < 10000 || step.RetryDelay > 60000 {
				return Invalid("http step retryDelay is invalid")
			}
			if len(step.URL) > 2048 {
				return Invalid("http step url is too long")
			}
		case domainstep.TypeDelay:
			if step.DelayDurationSeconds < 1 {
				return Invalid("delay step delayDurationSeconds is required")
			}
		case domainstep.TypeCondition:
			if strings.TrimSpace(step.Expression) == "" {
				return Invalid("condition step expression is required")
			}
		}
	}

	for _, conn := range d.Connections {
		if _, ok := stepRefs[strings.TrimSpace(conn.SourceRef)]; !ok {
			return Invalid("unknown connection sourceRef")
		}
		if _, ok := stepRefs[strings.TrimSpace(conn.TargetRef)]; !ok {
			return Invalid("unknown connection targetRef")
		}
		if strings.TrimSpace(conn.SourceRef) == strings.TrimSpace(conn.TargetRef) {
			return Invalid("connection source and target must be different")
		}
	}

	varRefs := make(map[string]struct{}, len(d.Variables))
	keys := make(map[string]struct{}, len(d.Variables))
	for _, variable := range d.Variables {
		ref := strings.TrimSpace(variable.Ref)
		if ref == "" {
			return Invalid("variable ref is required")
		}
		if _, exists := varRefs[ref]; exists {
			return Invalid("duplicate variable ref " + ref)
		}
		varRefs[ref] = struct{}{}
		key := strings.TrimSpace(variable.Key)
		if key == "" || strings.TrimSpace(variable.Name) == "" {
			return Invalid("variable name and key are required")
		}
		if _, exists := keys[key]; exists {
			return Invalid("duplicate variable key")
		}
		keys[key] = struct{}{}
		kind, err := domainvariable.ParseKind(variable.Kind)
		if err != nil {
			return Invalid("invalid variable kind")
		}
		switch kind {
		case domainvariable.KindExtracted:
			stepType, ok := stepRefs[strings.TrimSpace(variable.StepRef)]
			if !ok {
				return Invalid("unknown variable stepRef")
			}
			if stepType != string(domainstep.TypeHTTP) {
				return Invalid("extracted variables require an HTTP step")
			}
		case domainvariable.KindStatic:
			if strings.TrimSpace(variable.StepRef) != "" {
				return Invalid("static variables cannot reference a step")
			}
		}
	}

	for _, assertion := range d.Assertions {
		stepType, ok := stepRefs[strings.TrimSpace(assertion.StepRef)]
		if !ok {
			return Invalid("unknown assertion stepRef")
		}
		if stepType != string(domainstep.TypeHTTP) {
			return Invalid("assertions require an HTTP step")
		}
	}
	return nil
}
