package assertion

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	domainvariable "go-api/internal/domain/variable"
)

// Snapshot is the frozen assertion config copied onto a StepRun at creation.
type Snapshot struct {
	AssertionID   string            `json:"assertionId"`
	Description   string            `json:"description,omitempty"`
	Source        AssertionSource   `json:"source"`
	Path          string            `json:"path,omitempty"`
	Operator      AssertionOperator `json:"operator"`
	ExpectedValue string            `json:"expectedValue,omitempty"`
}

// Result is one evaluated assertion outcome stored on StepRun.assertionsResult.
type Result struct {
	AssertionID string `json:"assertionId"`
	Passed      bool   `json:"passed"`
	ActualValue any    `json:"actualValue"`
	Message     string `json:"message,omitempty"`
}

// ResponseInput is the HTTP response surface used for evaluation.
type ResponseInput struct {
	Status  int
	Headers map[string]string
	Body    any
}

func EvaluateAll(snapshots []Snapshot, response ResponseInput) []Result {
	results := make([]Result, 0, len(snapshots))
	for _, snap := range snapshots {
		results = append(results, Evaluate(snap, response))
	}
	return results
}

func AnyFailed(results []Result) bool {
	for _, r := range results {
		if !r.Passed {
			return true
		}
	}
	return false
}

func FailureSummary(results []Result) string {
	failed := make([]string, 0)
	for _, r := range results {
		if r.Passed {
			continue
		}
		if r.Message != "" {
			failed = append(failed, fmt.Sprintf("%s: %s", r.AssertionID, r.Message))
			continue
		}
		failed = append(failed, r.AssertionID)
	}
	if len(failed) == 0 {
		return "assertion failed"
	}
	return "assertion failed: " + strings.Join(failed, "; ")
}

func Evaluate(snap Snapshot, response ResponseInput) Result {
	actual, resolveErr := resolveActual(snap, response)
	passed, message := applyOperator(snap.Operator, snap.ExpectedValue, actual, snap.Path, resolveErr)
	return Result{
		AssertionID: snap.AssertionID,
		Passed:      passed,
		ActualValue: actual,
		Message:     message,
	}
}

func resolveActual(snap Snapshot, response ResponseInput) (any, error) {
	switch snap.Source {
	case SourceStatus:
		return response.Status, nil
	case SourceHeader:
		value, ok := lookupHeader(response.Headers, snap.Path)
		if !ok {
			return nil, fmt.Errorf("header %q not found", snap.Path)
		}
		return value, nil
	case SourceBody:
		value, err := domainvariable.ExtractByPath(response.Body, snap.Path)
		if err != nil {
			return nil, fmt.Errorf("path %s not found", snap.Path)
		}
		return value, nil
	default:
		return nil, fmt.Errorf("invalid source %q", snap.Source)
	}
}

func lookupHeader(headers map[string]string, name string) (string, bool) {
	if headers == nil {
		return "", false
	}
	if v, ok := headers[name]; ok {
		return v, true
	}
	target := strings.ToLower(name)
	for k, v := range headers {
		if strings.ToLower(k) == target {
			return v, true
		}
	}
	return "", false
}

func applyOperator(
	operator AssertionOperator,
	expectedRaw string,
	actual any,
	path string,
	resolveErr error,
) (bool, string) {
	switch operator {
	case OperatorNotNull:
		if resolveErr != nil || actual == nil {
			return false, fmt.Sprintf("expected non-null value at %s", displayPath(path))
		}
		return true, ""
	case OperatorIsNull:
		if resolveErr != nil || actual == nil {
			return true, ""
		}
		return false, fmt.Sprintf("expected null value at %s", displayPath(path))
	case OperatorIsString:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		if _, ok := actual.(string); ok {
			return true, ""
		}
		return false, fmt.Sprintf("expected string at %s", displayPath(path))
	case OperatorIsNumber:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		if isNumber(actual) {
			return true, ""
		}
		return false, fmt.Sprintf("expected number at %s", displayPath(path))
	case OperatorIsBoolean:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		if _, ok := actual.(bool); ok {
			return true, ""
		}
		return false, fmt.Sprintf("expected boolean at %s", displayPath(path))
	case OperatorIsArray:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		if isArray(actual) {
			return true, ""
		}
		return false, fmt.Sprintf("expected array at %s", displayPath(path))
	case OperatorIsObject:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		if isObject(actual) {
			return true, ""
		}
		return false, fmt.Sprintf("expected object at %s", displayPath(path))
	case OperatorEquals:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		expected, err := coerceExpected(expectedRaw, actual)
		if err != nil {
			return false, err.Error()
		}
		if valuesEqual(actual, expected) {
			return true, ""
		}
		return false, fmt.Sprintf("expected %v, got %v", expected, actual)
	case OperatorNotEquals:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		expected, err := coerceExpected(expectedRaw, actual)
		if err != nil {
			return false, err.Error()
		}
		if !valuesEqual(actual, expected) {
			return true, ""
		}
		return false, fmt.Sprintf("expected value different from %v", expected)
	case OperatorContains:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		return containsValue(actual, expectedRaw)
	case OperatorGreaterThan:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		return compareNumeric(actual, expectedRaw, true)
	case OperatorLessThan:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		return compareNumeric(actual, expectedRaw, false)
	case OperatorMatchesRegex:
		if resolveErr != nil {
			return false, resolveErr.Error()
		}
		s, ok := actual.(string)
		if !ok {
			return false, "actual value is not a string"
		}
		re, err := regexp.Compile(expectedRaw)
		if err != nil {
			return false, "invalid regular expression"
		}
		if re.MatchString(s) {
			return true, ""
		}
		return false, fmt.Sprintf("value %q does not match %q", s, expectedRaw)
	default:
		return false, fmt.Sprintf("unsupported operator %q", operator)
	}
}

func displayPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return "status"
	}
	return path
}

func coerceExpected(expectedRaw string, actual any) (any, error) {
	switch actual.(type) {
	case nil:
		return expectedRaw, nil
	case string:
		return expectedRaw, nil
	case bool:
		switch strings.ToLower(expectedRaw) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return nil, fmt.Errorf("cannot coerce %q to boolean", expectedRaw)
		}
	case float64, float32, int, int32, int64, uint, uint32, uint64:
		n, err := strconv.ParseFloat(expectedRaw, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot coerce %q to number", expectedRaw)
		}
		return n, nil
	default:
		return expectedRaw, nil
	}
}

func valuesEqual(actual, expected any) bool {
	if actualNum, ok := asFloat(actual); ok {
		if expectedNum, ok := asFloat(expected); ok {
			return actualNum == expectedNum
		}
	}
	return reflect.DeepEqual(actual, expected)
}

func compareNumeric(actual any, expectedRaw string, greater bool) (bool, string) {
	actualNum, ok := asFloat(actual)
	if !ok {
		return false, "actual value is not numeric"
	}
	expectedNum, err := strconv.ParseFloat(strings.TrimSpace(expectedRaw), 64)
	if err != nil {
		return false, "expected value is not numeric"
	}
	if greater {
		if actualNum > expectedNum {
			return true, ""
		}
		return false, fmt.Sprintf("expected > %v, got %v", expectedNum, actualNum)
	}
	if actualNum < expectedNum {
		return true, ""
	}
	return false, fmt.Sprintf("expected < %v, got %v", expectedNum, actualNum)
}

func containsValue(actual any, expectedRaw string) (bool, string) {
	switch v := actual.(type) {
	case string:
		if strings.Contains(v, expectedRaw) {
			return true, ""
		}
		return false, fmt.Sprintf("expected substring %q in %q", expectedRaw, v)
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == expectedRaw {
				return true, ""
			}
			if num, ok := asFloat(item); ok {
				if expectedNum, err := strconv.ParseFloat(expectedRaw, 64); err == nil && num == expectedNum {
					return true, ""
				}
			}
			if valuesEqual(item, expectedRaw) {
				return true, ""
			}
		}
		return false, fmt.Sprintf("expected element %q in array", expectedRaw)
	default:
		return false, "contains is only supported for string or array values"
	}
}

func isNumber(v any) bool {
	_, ok := asFloat(v)
	return ok
}

func asFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	default:
		return 0, false
	}
}

func isArray(v any) bool {
	if v == nil {
		return false
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array
}

func isObject(v any) bool {
	if v == nil {
		return false
	}
	switch v.(type) {
	case map[string]any:
		return true
	default:
		return reflect.ValueOf(v).Kind() == reflect.Map
	}
}
