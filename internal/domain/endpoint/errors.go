package endpoint

import "errors"

var (
	ErrInvalidOpenAPI     = errors.New("invalid OpenAPI document")
	ErrNoOperations       = errors.New("OpenAPI document has no HTTP operations")
	ErrTooManyOperations  = errors.New("OpenAPI document has too many HTTP operations")
	ErrInvalidEndpointURL = errors.New("invalid endpoint url")
)
