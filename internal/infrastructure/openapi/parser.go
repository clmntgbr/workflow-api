package openapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/port"
)

var importMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(_ context.Context, spec []byte) ([]port.OpenAPIOperation, error) {
	root, err := parseOpenAPIRoot(spec)
	if err != nil {
		return nil, err
	}

	rawPaths, ok := root["paths"].(map[string]any)
	if !ok || len(rawPaths) == 0 {
		return nil, domainendpoint.ErrNoOperations
	}

	paths := make([]string, 0, len(rawPaths))
	for path := range rawPaths {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	operations := make([]port.OpenAPIOperation, 0)
	for _, path := range paths {
		if !strings.HasPrefix(path, "/") {
			return nil, fmt.Errorf("%w: path %q must start with /", domainendpoint.ErrInvalidOpenAPI, path)
		}
		item, ok := rawPaths[path].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: path %q must be an object", domainendpoint.ErrInvalidOpenAPI, path)
		}
		for _, method := range importMethods {
			rawOperation, exists := item[strings.ToLower(method)]
			if !exists {
				continue
			}
			operation, ok := rawOperation.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%w: operation %s %s must be an object", domainendpoint.ErrInvalidOpenAPI, method, path)
			}
			operations = append(operations, port.OpenAPIOperation{
				Summary:     stringField(operation, "summary"),
				OperationID: stringField(operation, "operationId"),
				Description: stringField(operation, "description"),
				Method:      method,
				Path:        path,
			})
		}
	}
	if len(operations) == 0 {
		return nil, domainendpoint.ErrNoOperations
	}
	return operations, nil
}

func parseOpenAPIRoot(spec []byte) (map[string]any, error) {
	trimmed := bytes.TrimSpace(spec)
	if len(trimmed) == 0 || !json.Valid(trimmed) {
		return nil, fmt.Errorf("%w: file must be valid JSON", domainendpoint.ErrInvalidOpenAPI)
	}

	var root map[string]any
	if err := json.Unmarshal(trimmed, &root); err != nil {
		return nil, fmt.Errorf("%w: file must be a JSON object", domainendpoint.ErrInvalidOpenAPI)
	}
	if swagger, _ := root["swagger"].(string); strings.HasPrefix(swagger, "2.") {
		return nil, fmt.Errorf("%w: OpenAPI 3.x is required", domainendpoint.ErrInvalidOpenAPI)
	}

	version, _ := root["openapi"].(string)
	if version == "" || !strings.HasPrefix(version, "3.") {
		return nil, fmt.Errorf("%w: missing or unsupported openapi version", domainendpoint.ErrInvalidOpenAPI)
	}

	info, ok := root["info"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: info is required", domainendpoint.ErrInvalidOpenAPI)
	}
	if stringField(info, "title") == "" || stringField(info, "version") == "" {
		return nil, fmt.Errorf("%w: info.title and info.version are required", domainendpoint.ErrInvalidOpenAPI)
	}

	if _, ok := root["paths"]; !ok {
		return nil, fmt.Errorf("%w: paths is required", domainendpoint.ErrInvalidOpenAPI)
	}
	if _, ok := root["paths"].(map[string]any); !ok {
		return nil, fmt.Errorf("%w: paths must be an object", domainendpoint.ErrInvalidOpenAPI)
	}

	return root, nil
}

func stringField(object map[string]any, key string) string {
	value, _ := object[key].(string)
	return strings.TrimSpace(value)
}
