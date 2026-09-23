package service

import (
	"encoding/json"
	"testing"
)

func TestNormalizeGrokToolSchemasFiltersNonObjectRootBranch(t *testing.T) {
	body := []byte(`{"model":"grok-4.7","tools":[{"type":"function","function":{"name":"automation_update","parameters":{"anyOf":[{"type":"object","properties":{"mode":{"type":"string"}}},{"type":"null"}]}}}]}`)
	got := normalizeGrokToolSchemas(body)
	var req map[string]any
	if err := json.Unmarshal(got, &req); err != nil {
		t.Fatal(err)
	}
	params := req["tools"].([]any)[0].(map[string]any)["function"].(map[string]any)["parameters"].(map[string]any)
	if params["type"] != "object" {
		t.Fatalf("root type = %#v", params["type"])
	}
	branches := params["anyOf"].([]any)
	if len(branches) != 1 || branches[0].(map[string]any)["type"] != "object" {
		t.Fatalf("branches = %#v", branches)
	}
}

func TestNormalizeGrokToolSchemasPreservesNestedUnionAndUnchangedBody(t *testing.T) {
	body := []byte(`{"model":"gpt-5","tools":[{"type":"function","function":{"name":"x","parameters":{"type":"object","properties":{"value":{"anyOf":[{"type":"string"},{"type":"null"}]}}}}}]}`)
	got := normalizeGrokToolSchemas(body)
	if string(got) != string(body) {
		t.Fatalf("non-Grok body changed: %s", got)
	}
}
