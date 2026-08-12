package dbtype

import (
	"testing"
)

func TestJSONBValueReturnsString(t *testing.T) {
	payload := JSONB(`{"firstName":"Clément"}`)
	v, err := payload.Value()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("expected string driver.Value, got %T", v)
	}
	if s != `{"firstName":"Clément"}` {
		t.Fatalf("unexpected value: %s", s)
	}
}
