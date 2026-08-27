package assertion

import "fmt"

type AssertionOperator string

const (
	OperatorEquals       AssertionOperator = "equals"
	OperatorNotEquals    AssertionOperator = "not_equals"
	OperatorNotNull      AssertionOperator = "not_null"
	OperatorIsNull       AssertionOperator = "is_null"
	OperatorContains     AssertionOperator = "contains"
	OperatorGreaterThan  AssertionOperator = "greater_than"
	OperatorLessThan     AssertionOperator = "less_than"
	OperatorMatchesRegex AssertionOperator = "matches_regex"
	OperatorIsString     AssertionOperator = "is_string"
	OperatorIsNumber     AssertionOperator = "is_number"
	OperatorIsBoolean    AssertionOperator = "is_boolean"
	OperatorIsArray      AssertionOperator = "is_array"
	OperatorIsObject     AssertionOperator = "is_object"
)

func (o AssertionOperator) Valid() bool {
	switch o {
	case OperatorEquals,
		OperatorNotEquals,
		OperatorNotNull,
		OperatorIsNull,
		OperatorContains,
		OperatorGreaterThan,
		OperatorLessThan,
		OperatorMatchesRegex,
		OperatorIsString,
		OperatorIsNumber,
		OperatorIsBoolean,
		OperatorIsArray,
		OperatorIsObject:
		return true
	default:
		return false
	}
}

func (o AssertionOperator) RequiresExpectedValue() bool {
	switch o {
	case OperatorNotNull,
		OperatorIsNull,
		OperatorIsString,
		OperatorIsNumber,
		OperatorIsBoolean,
		OperatorIsArray,
		OperatorIsObject:
		return false
	default:
		return true
	}
}

func ParseOperator(raw string) (AssertionOperator, error) {
	op := AssertionOperator(raw)
	if !op.Valid() {
		return "", fmt.Errorf("%w: %s", ErrInvalidOperator, raw)
	}
	return op, nil
}
