package dbtype

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONB []byte

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

func (j *JSONB) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*j = append(JSONB(nil), v...)
		return nil
	case string:
		*j = append(JSONB(nil), v...)
		return nil
	default:
		return fmt.Errorf("unsupported JSONB scan type %T", value)
	}
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("JSONB: UnmarshalJSON on nil pointer")
	}
	*j = append(JSONB(nil), data...)
	return nil
}
