package runexport

import (
	"strconv"
	"time"

	domaininsight "go-api/internal/domain/insight"
	domainrunexport "go-api/internal/domain/runexport"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

const (
	sheetRuns       = "Runs"
	sheetAssertions = "Assertions"
	maxFileBytes    = 15 * 1024 * 1024
	kindRun         = "run"
	kindStep        = "step"
)

type BuildInput struct {
	Runs         []domainworkflowrun.WorkflowRunView
	StepRuns     []domainsteprun.StepRunView
	Insights     []domaininsight.InsightView
	WithInsights bool
}

func BuildXLSX(in BuildInput) ([]byte, error) {
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()

	if err := file.SetSheetName("Sheet1", sheetRuns); err != nil {
		return nil, err
	}
	if _, err := file.NewSheet(sheetAssertions); err != nil {
		return nil, err
	}

	insightsByStepRun := indexInsights(in.Insights)
	stepRunsByRun := groupStepRuns(in.StepRuns)

	if err := writeRunsSheet(file, in.Runs, stepRunsByRun, insightsByStepRun, in.WithInsights); err != nil {
		return nil, err
	}
	if err := writeAssertionsSheet(file, in.Runs, stepRunsByRun); err != nil {
		return nil, err
	}

	buf, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	if buf.Len() > maxFileBytes {
		return nil, domainrunexport.ErrFileTooLarge
	}
	return buf.Bytes(), nil
}

func writeRunsSheet(
	file *excelize.File,
	runs []domainworkflowrun.WorkflowRunView,
	stepRunsByRun map[uuid.UUID][]domainsteprun.StepRunView,
	insights map[uuid.UUID]domaininsight.InsightView,
	withInsights bool,
) error {
	headers := []string{
		"kind",
		"runId", "runStatus", "runStartedAt", "runFinishedAt", "runDurationMs",
		"runCreatedAt", "runError",
		"stepId", "stepName", "stepDescription", "stepType", "stepStatus",
		"executionOrder", "attempt", "stepStartedAt", "stepFinishedAt", "stepDurationMs",
		"method", "url", "timeout", "retryOnFailure", "retryCount", "retryDelay",
		"requestHeaders", "requestQuery", "requestBody",
		"responseStatus",
		"extractedVariables", "matchedBranch", "resumeAt", "delaySeconds", "stepError",
	}
	if withInsights {
		headers = append(headers,
			"durationMs", "ttfbMs", "dnsMs", "tcpMs", "tlsMs",
			"requestSize", "responseSize", "insightAttempts", "insightError",
		)
	}
	if err := writeHeader(file, sheetRuns, headers); err != nil {
		return err
	}

	row := 2
	for i, run := range runs {
		if err := writeRow(file, sheetRuns, row, runRow(run, withInsights)); err != nil {
			return err
		}
		row++
		for _, stepRun := range stepRunsByRun[run.ID] {
			if err := writeRow(file, sheetRuns, row, stepRow(run.ID, stepRun, insights[stepRun.ID], withInsights)); err != nil {
				return err
			}
			row++
		}
		if i < len(runs)-1 {
			row++
		}
	}
	return nil
}

func runRow(run domainworkflowrun.WorkflowRunView, withInsights bool) []any {
	values := []any{
		kindRun,
		run.ID.String(),
		string(run.Status),
		formatTime(run.StartedAt),
		formatTime(run.FinishedAt),
		durationBetween(run.StartedAt, run.FinishedAt),
		run.CreatedAt.UTC().Format(time.RFC3339),
		run.Error,
		"", "", "", "", "",
		"", "", "", "", "",
		"", "", "", "", "", "",
		"", "", "",
		"",
		"", "", "", "", "",
	}
	if withInsights {
		values = append(values, "", "", "", "", "", "", "", "", "")
	}
	return values
}

func stepRow(
	runID uuid.UUID,
	stepRun domainsteprun.StepRunView,
	insight domaininsight.InsightView,
	withInsights bool,
) []any {
	responseStatus := ""
	if stepRun.ResponseSnapshot != nil {
		responseStatus = strconv.Itoa(stepRun.ResponseSnapshot.Status)
	}

	values := []any{
		kindStep,
		runID.String(),
		"", "", "", "", "", "",
		stepRun.ID.String(),
		stepRun.Name,
		stepRun.Description,
		string(stepRun.StepType),
		string(stepRun.Status),
		stepRun.ExecutionOrder,
		stepRun.Attempt,
		formatTime(stepRun.StartedAt),
		formatTime(stepRun.FinishedAt),
		durationBetween(stepRun.StartedAt, stepRun.FinishedAt),
		stepRun.Method,
		stepRun.URL,
		stepRun.Timeout,
		stepRun.RetryOnFailure,
		stepRun.RetryCount,
		stepRun.RetryDelay,
		toJSONCell(redactHeaders(stepRun.Headers)),
		toJSONCell(redactQuery(stepRun.Query)),
		toJSONCell(redactBody(stepRun.Body)),
		responseStatus,
		toJSONCell(redactBody(stepRun.ExtractedVariables)),
		formatBool(stepRun.MatchedBranch),
		formatTime(stepRun.ResumeAt),
		delaySeconds(stepRun),
		stepRun.Error,
	}
	if withInsights {
		if insight.ID != uuid.Nil {
			values = append(values,
				durationMS(insight.Duration),
				durationMS(insight.TTFB),
				durationMS(insight.DNSLookupDuration),
				durationMS(insight.TCPConnectionTime),
				durationMS(insight.TLSHandshakeTime),
				int64Cell(insight.RequestSize),
				int64Cell(insight.ResponseSize),
				insight.TotalAttempts,
				insight.ErrorMessage,
			)
		} else {
			values = append(values, "", "", "", "", "", "", "", "", "")
		}
	}
	return values
}

func writeAssertionsSheet(
	file *excelize.File,
	runs []domainworkflowrun.WorkflowRunView,
	stepRunsByRun map[uuid.UUID][]domainsteprun.StepRunView,
) error {
	headers := []string{
		"runId", "stepName", "description", "source", "operator",
		"expected", "passed", "actual", "message",
	}
	if err := writeHeader(file, sheetAssertions, headers); err != nil {
		return err
	}

	row := 2
	wroteGroup := false
	for _, run := range runs {
		groupStarted := false
		for _, stepRun := range stepRunsByRun[run.ID] {
			snapshots := make(map[string]int, len(stepRun.Assertions))
			for idx, snap := range stepRun.Assertions {
				snapshots[snap.AssertionID] = idx
			}
			for _, result := range stepRun.AssertionsResult {
				if wroteGroup && !groupStarted {
					row++
				}
				groupStarted = true
				wroteGroup = true

				description, source, operator, expected := "", "", "", ""
				if idx, ok := snapshots[result.AssertionID]; ok {
					snap := stepRun.Assertions[idx]
					description = snap.Description
					source = string(snap.Source)
					operator = string(snap.Operator)
					expected = snap.ExpectedValue
				}
				values := []any{
					stepRun.WorkflowRunID.String(),
					stepRun.Name,
					description,
					source,
					operator,
					expected,
					result.Passed,
					toJSONCell(result.ActualValue),
					result.Message,
				}
				if err := writeRow(file, sheetAssertions, row, values); err != nil {
					return err
				}
				row++
			}
		}
	}
	return nil
}

func writeHeader(file *excelize.File, sheet string, headers []string) error {
	values := make([]any, len(headers))
	for i, header := range headers {
		values[i] = header
	}
	return writeRow(file, sheet, 1, values)
}

func writeRow(file *excelize.File, sheet string, row int, values []any) error {
	for i, value := range values {
		cell, err := excelize.CoordinatesToCellName(i+1, row)
		if err != nil {
			return err
		}
		if err := file.SetCellValue(sheet, cell, value); err != nil {
			return err
		}
	}
	return nil
}

func groupStepRuns(stepRuns []domainsteprun.StepRunView) map[uuid.UUID][]domainsteprun.StepRunView {
	out := make(map[uuid.UUID][]domainsteprun.StepRunView)
	for _, stepRun := range stepRuns {
		out[stepRun.WorkflowRunID] = append(out[stepRun.WorkflowRunID], stepRun)
	}
	return out
}

func indexInsights(views []domaininsight.InsightView) map[uuid.UUID]domaininsight.InsightView {
	out := make(map[uuid.UUID]domaininsight.InsightView, len(views))
	for _, view := range views {
		if _, exists := out[view.StepRunID]; !exists {
			out[view.StepRunID] = view
		}
	}
	return out
}

func formatTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func formatBool(value *bool) string {
	if value == nil {
		return ""
	}
	if *value {
		return "true"
	}
	return "false"
}

func durationBetween(start, end *time.Time) any {
	if start == nil || end == nil {
		return ""
	}
	return end.Sub(*start).Milliseconds()
}

func durationMS(d *time.Duration) any {
	if d == nil {
		return ""
	}
	return d.Milliseconds()
}

func int64Cell(v *int64) any {
	if v == nil {
		return ""
	}
	return *v
}

func delaySeconds(stepRun domainsteprun.StepRunView) any {
	if stepRun.DelayDurationSeconds <= 0 {
		return ""
	}
	return stepRun.DelayDurationSeconds
}
