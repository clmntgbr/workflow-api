package condition

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"

	domainvariable "go-api/internal/domain/variable"

	"github.com/expr-lang/expr"
)

var placeholderPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_-]+)\s*\}\}`)

var (
	ErrInvalidExpression = errors.New("invalid condition expression")
	ErrNotBooleanResult  = errors.New("condition expression must evaluate to a boolean")
)

func ValidateSyntax(expression string) error {
	if expression == "" {
		return ErrInvalidExpression
	}
	resolved, err := substitutePlaceholders(expression, dummyContext(expression))
	if err != nil {
		return err
	}
	program, err := expr.Compile(resolved, expr.AsBool())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidExpression, err)
	}
	_ = program
	return nil
}

func EvaluateBoolean(expression string, context map[string]any) (bool, error) {
	if expression == "" {
		return false, ErrInvalidExpression
	}
	resolved, err := substitutePlaceholders(expression, context)
	if err != nil {
		return false, err
	}
	program, err := expr.Compile(resolved, expr.AsBool())
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrInvalidExpression, err)
	}
	result, err := expr.Run(program, nil)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrInvalidExpression, err)
	}
	value, ok := result.(bool)
	if !ok {
		return false, ErrNotBooleanResult
	}
	return value, nil
}

func substitutePlaceholders(expression string, context map[string]any) (string, error) {
	var missing *domainvariable.MissingVariableError
	out := placeholderPattern.ReplaceAllStringFunc(expression, func(match string) string {
		sub := placeholderPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		key := sub[1]
		value, ok := context[key]
		if !ok {
			missing = &domainvariable.MissingVariableError{Key: key}
			return match
		}
		return literalForExpr(value)
	})
	if missing != nil {
		return "", missing
	}
	return out, nil
}

func dummyContext(expression string) map[string]any {
	context := map[string]any{}
	for _, match := range placeholderPattern.FindAllStringSubmatch(expression, -1) {
		if len(match) < 2 {
			continue
		}
		context[match[1]] = ""
	}
	return context
}

func literalForExpr(value any) string {
	switch typed := value.(type) {
	case nil:
		return "nil"
	case string:
		return strconv.Quote(typed)
	case bool:
		return strconv.FormatBool(typed)
	case json.Number:
		return typed.String()
	case float64, float32, int, int64, int32:
		return fmt.Sprint(typed)
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return strconv.Quote(fmt.Sprint(typed))
		}
		return string(encoded)
	}
}
