package port

import "context"

type OpenAPIOperation struct {
	Summary     string
	OperationID string
	Description string
	Method      string
	Path        string
}

type OpenAPISpecParser interface {
	Parse(ctx context.Context, spec []byte) ([]OpenAPIOperation, error)
}
