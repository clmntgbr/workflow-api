package assertion

import (
	"testing"
)

func TestEvaluateStatusEquals(t *testing.T) {
	result := Evaluate(Snapshot{
		AssertionID:   "1",
		Source:        SourceStatus,
		Operator:      OperatorEquals,
		ExpectedValue: "200",
	}, ResponseInput{Status: 200})
	if !result.Passed {
		t.Fatalf("expected pass, got %#v", result)
	}
}

func TestEvaluateBodyNotNull(t *testing.T) {
	result := Evaluate(Snapshot{
		AssertionID: "2",
		Source:      SourceBody,
		Path:        "$.id",
		Operator:    OperatorNotNull,
	}, ResponseInput{Body: map[string]any{"id": "abc"}})
	if !result.Passed {
		t.Fatalf("expected pass, got %#v", result)
	}

	missing := Evaluate(Snapshot{
		AssertionID: "2",
		Source:      SourceBody,
		Path:        "$.id",
		Operator:    OperatorNotNull,
	}, ResponseInput{Body: map[string]any{}})
	if missing.Passed {
		t.Fatalf("expected fail for missing path")
	}
}

func TestEvaluateHeaderContains(t *testing.T) {
	result := Evaluate(Snapshot{
		AssertionID:   "3",
		Source:        SourceHeader,
		Path:          "Content-Type",
		Operator:      OperatorContains,
		ExpectedValue: "application/json",
	}, ResponseInput{Headers: map[string]string{"content-type": "application/json; charset=utf-8"}})
	if !result.Passed {
		t.Fatalf("expected pass, got %#v", result)
	}
}

func TestValidateShapeRejectsExpectedForNotNull(t *testing.T) {
	err := ValidateShape(SourceBody, "$.id", OperatorNotNull, "x")
	if err != ErrExpectedValueForbidden {
		t.Fatalf("expected ErrExpectedValueForbidden, got %v", err)
	}
}

func TestValidateShapeRejectsInvalidRegex(t *testing.T) {
	err := ValidateShape(SourceBody, "$.name", OperatorMatchesRegex, "[")
	if err != ErrInvalidExpectedRegex {
		t.Fatalf("expected ErrInvalidExpectedRegex, got %v", err)
	}
}
