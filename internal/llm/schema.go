package llm

import (
	"encoding/json"

	"maps"

	"github.com/google/generative-ai-go/genai"
)

// TranslateGenaiSchemaToJSONSchema translates a genai.Schema to
// a JSON-Schema string.
func TranslateGenaiSchemaToJSONSchema(s *genai.Schema) string {
	// Start with top-level fields.
	out := map[string]any{
		"$schema": "http://json-schema.org/draft-07/schema#",
	}

	// Convert the provided schema and merge into the output.
	converted := convertSchema(s)
	maps.Copy(out, converted)

	b, _ := json.MarshalIndent(out, "", "  ")
	return string(b)
}

// convertSchema recursively converts a genai.Schema into a JSON-Schema
// represented as a map[string]any.
func convertSchema(s *genai.Schema) map[string]any {
	var res map[string]any

	switch s.Type {
	case genai.TypeString:
		res = map[string]any{
			"type": "string",
		}
		if s.Format != "" {
			res["format"] = s.Format
		}
		if len(s.Enum) > 0 {
			res["enum"] = s.Enum
		}
	case genai.TypeNumber:
		res = map[string]any{
			"type": "number",
		}
		if s.Format != "" {
			res["format"] = s.Format
		}
	case genai.TypeInteger:
		res = map[string]any{
			"type": "integer",
		}
		if s.Format != "" {
			res["format"] = s.Format
		}
	case genai.TypeBoolean:
		res = map[string]any{
			"type": "boolean",
		}
	case genai.TypeArray:
		res = map[string]any{
			"type": "array",
		}
		if s.Items != nil {
			res["items"] = convertSchema(s.Items)
		}
	case genai.TypeObject:
		res = map[string]any{
			"type": "object",
		}
		// Map properties if any.
		if s.Properties != nil && len(s.Properties) > 0 {
			props := map[string]any{}
			for propName, propSchema := range s.Properties {
				props[propName] = convertSchema(propSchema)
			}
			res["properties"] = props
		}
		// Filter out any required fields whose schema is nullable.
		if len(s.Required) > 0 {
			req := []string{}
			for _, key := range s.Required {
				if prop, ok := s.Properties[key]; ok {
					if !prop.Nullable {
						req = append(req, key)
					}
				} else {
					// If not defined in properties, add it.
					req = append(req, key)
				}
			}
			if len(req) > 0 {
				res["required"] = req
			}
		}
		// Disallow extra properties.
		res["additionalProperties"] = false
	default:
		res = map[string]any{}
	}

	if s.Description != "" {
		res["description"] = s.Description
	}

	// If the schema is nullable, wrap it using anyOf.
	if s.Nullable {
		res = map[string]any{
			"anyOf": []any{
				res,
				map[string]any{"type": "null"},
			},
		}
	}

	return res
}
