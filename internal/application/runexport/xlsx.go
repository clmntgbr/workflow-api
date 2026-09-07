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
	sheetStepRuns   = "Step runs"
	sheetAssertions = "Assertions"
	maxFileBytes    = 15 * 1024 * 1024
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
	if _, err := file.NewSheet(sheetStepRuns); err != nil {
		return nil, err
	}
	if _, err := file.NewSheet(sheetAssertions); err != nil {
		return nil, err
	}

	insightsByStepRun := indexInsights(in.Insights)

	if err := writeRunsSheet(file, in.Runs); err != nil {
		return nil, err
	}
	if err := writeStepRunsSheet(file, in.StepRuns, insightsByStepRun, in.WithInsights); err != nil {
		return nil, err
	}
	if err := writeAssertionsSheet(file, in.StepRuns); err != nil {
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

func writeRunsSheet(file *excelize.File, runs []domainworkflowrun.WorkflowRunView) error {
	headers := []string{"id", "status", "triggeredBy", "startedAt", "finishedAt", "error"}
	if err := writeHeader(file, sheetRuns, headers); err != nil {
		return err
	}
	for i, run := range runs {
		row := i + 2
		values := []any{
			run.ID.String(),
			string(run.Status),
			string(run.TriggeredBy),
			formatTime(run.StartedAt),
			formatTime(run.FinishedAt),
			run.Error,
		}
		if err := writeRow(file, sheetRuns, row, values); err != nil {
			return err
		}
	}
	return nil
}

func writeStepRunsSheet(
	file *excelize.File,
	stepRuns []domainsteprun.StepRunView,
	insights map[uuid.UUID]domaininsight.InsightView,
	withInsights bool,
) error {
	headers := []string{
		"runId", "stepName", "type", "status", "attempt",
		"method", "url", "requestHeaders", "requestQuery", "requestBody",
		"responseStatus", "responseHeaders", "responseBody",
		"extractedVariables", "matchedBranch", "error",
	}
	if withInsights {
		headers = append(headers,
			"durationMs", "ttfbMs", "dnsMs", "tcpMs", "tlsMs",
			"requestSize", "responseSize", "insightAttempts", "insightError",
		)
	}
	if err := writeHeader(file, sheetStepRuns, headers); err != nil {
		return err
	}

	for i, stepRun := range stepRuns {
		row := i + 2
		responseStatus := ""
		var responseHeaders any
		var responseBody any
		if stepRun.ResponseSnapshot != nil {
			responseStatus = strconv.Itoa(stepRun.ResponseSnapshot.Status)
			responseHeaders = redactHeaders(stepRun.ResponseSnapshot.Headers)
			responseBody = redactBody(stepRun.ResponseSnapshot.Body)
		}
		values := []any{
			stepRun.WorkflowRunID.String(),
			stepRun.Name,
			string(stepRun.StepType),
			string(stepRun.Status),
			stepRun.Attempt,
			stepRun.Method,
			stepRun.URL,
			toJSONCell(redactHeaders(stepRun.Headers)),
			toJSONCell(redactQuery(stepRun.Query)),
			toJSONCell(redactBody(stepRun.Body)),
			responseStatus,
			toJSONCell(responseHeaders),
			toJSONCell(responseBody),
			toJSONCell(redactBody(stepRun.ExtractedVariables)),
			formatBool(stepRun.MatchedBranch),
			stepRun.Error,
		}
		if withInsights {
			insight, ok := insights[stepRun.ID]
			if ok {
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
		if err := writeRow(file, sheetStepRuns, row, values); err != nil {
			return err
		}
	}
	return nil
}

func writeAssertionsSheet(file *excelize.File, stepRuns []domainsteprun.StepRunView) error {
	headers := []string{
		"runId", "stepName", "description", "source", "operator",
		"expected", "passed", "actual", "message",
	}
	if err := writeHeader(file, sheetAssertions, headers); err != nil {
		return err
	}

	row := 2
	for _, stepRun := range stepRuns {
		snapshots := make(map[string]int, len(stepRun.Assertions))
		for i, snap := range stepRun.Assertions {
			snapshots[snap.AssertionID] = i
		}
		for _, result := range stepRun.AssertionsResult {
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
