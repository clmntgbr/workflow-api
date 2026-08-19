package variable

import (
	"encoding/json"
	"fmt"

	"github.com/PaesslerAG/jsonpath"
)

func ExtractByPath(body any, path string) (any, error) {
	if path == "" {
		return nil, ErrInvalidPath
	}
	path = NormalizeJSONPath(path)
	result, err := jsonpath.Get(path, body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPath, err)
	}
	return result, nil
}

func ToRawMessage(value any) (json.RawMessage, error) {
	if value == nil {
		return json.RawMessage("null"), nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}
