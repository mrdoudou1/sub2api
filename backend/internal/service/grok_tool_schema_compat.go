package service

import "encoding/json"

// normalizeGrokToolSchemas fixes Grok's root-level function parameter schema
// restrictions without changing nested property unions.
func normalizeGrokToolSchemas(body []byte) []byte {
	var request map[string]any
	if json.Unmarshal(body, &request) != nil {
		return body
	}
	tools, ok := request["tools"].([]any)
	if !ok {
		return body
	}
	changed := false
	for _, raw := range tools {
		tool, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if fn, ok := tool["function"].(map[string]any); ok {
			if schema, ok := fn["parameters"].(map[string]any); ok && normalizeGrokRootSchema(schema) {
				changed = true
			}
		}
		if schema, ok := tool["parameters"].(map[string]any); ok && normalizeGrokRootSchema(schema) {
			changed = true
		}
	}
	if !changed {
		return body
	}
	encoded, err := json.Marshal(request)
	if err != nil {
		return body
	}
	return encoded
}

func normalizeGrokRootSchema(schema map[string]any) bool {
	changed := false
	defs, _ := schema["$defs"].(map[string]any)
	if ref, ok := schema["$ref"].(string); ok {
		if target, ok := resolveGrokSchemaRef(defs, ref); ok {
			for key, value := range target {
				if key != "$ref" {
					if _, exists := schema[key]; !exists {
						schema[key] = value
					}
				}
			}
			delete(schema, "$ref")
			changed = true
		}
	}
	for _, keyword := range []string{"anyOf", "oneOf"} {
		branches, ok := schema[keyword].([]any)
		if !ok {
			continue
		}
		objects := make([]any, 0, len(branches))
		for _, raw := range branches {
			branch, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if ref, ok := branch["$ref"].(string); ok {
				if target, ok := resolveGrokSchemaRef(defs, ref); ok {
					branch = target
				}
			}
			if branch["type"] == "object" {
				objects = append(objects, branch)
			}
		}
		if len(objects) == 0 || len(objects) == len(branches) && schema["type"] == "object" {
			continue
		}
		schema["type"] = "object"
		schema[keyword] = objects
		changed = true
	}
	return changed
}

func resolveGrokSchemaRef(defs map[string]any, ref string) (map[string]any, bool) {
	const prefix = "#/$defs/"
	if len(ref) <= len(prefix) || ref[:len(prefix)] != prefix || defs == nil {
		return nil, false
	}
	target, ok := defs[ref[len(prefix):]].(map[string]any)
	return target, ok
}
