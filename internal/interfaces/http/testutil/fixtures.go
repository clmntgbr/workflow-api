package testutil

import "github.com/google/uuid"

// Stable UUIDs shared across handler tests.
var (
	TestUserID       = uuid.MustParse("01960000-0000-7000-8000-000000000001")
	TestProjectID    = uuid.MustParse("01960000-0000-7000-8000-000000000002")
	TestEndpointID   = uuid.MustParse("01960000-0000-7000-8000-000000000003")
	TestWorkflowID   = uuid.MustParse("01960000-0000-7000-8000-000000000004")
	TestStepID       = uuid.MustParse("01960000-0000-7000-8000-000000000005")
	TestConnectionID = uuid.MustParse("01960000-0000-7000-8000-000000000006")
	TestVariableID   = uuid.MustParse("01960000-0000-7000-8000-000000000007")
	TestAssertionID  = uuid.MustParse("01960000-0000-7000-8000-000000000008")
	TestWorkflowRunID = uuid.MustParse("01960000-0000-7000-8000-000000000009")
	TestPlanID       = uuid.MustParse("01960000-0000-7000-8000-00000000000a")
)
