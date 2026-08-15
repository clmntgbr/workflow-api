package variable

import (
	"reflect"
	"testing"
)

func TestExtractResponsePaths(t *testing.T) {
	data := map[string]any{
		"token": "abc",
		"user": map[string]any{
			"id":    "u1",
			"email": "a@b.c",
			"roles": []any{"ROLE_USER", "ROLE_ADMIN"},
			"orders": []any{
				map[string]any{"id": "o1"},
				map[string]any{"id": "o2"},
			},
		},
	}

	paths := ExtractResponsePaths(data, "")
	wantContains := []string{
		"$.token",
		"$.user",
		"$.user.id",
		"$.user.email",
		"$.user.roles[0]",
		"$.user.roles[1]",
		"$.user.orders[0].id",
		"$.user.orders[1].id",
	}
	for _, want := range wantContains {
		if !contains(paths, want) {
			t.Fatalf("expected path %q in %v", want, paths)
		}
	}

	filtered := ExtractResponsePaths(data, "email")
	if !reflect.DeepEqual(filtered, []string{"$.user.email"}) && !contains(filtered, "$.user.email") {
		t.Fatalf("expected email filter to include $.user.email, got %v", filtered)
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
