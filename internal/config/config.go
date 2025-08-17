package config

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const CookiecutterJSON = "cookiecutter.json"

// VarSpec represents a variable specification from cookiecutter.json
type VarSpec struct {
	Name    string
	Prompt  string
	Default any
	Choices []string
	Kind    string // "string","bool","number","choice","any"
}

// ParseCookiecutterJSON parses cookiecutter.json and returns variable specs
func ParseCookiecutterJSON(b []byte) (map[string]VarSpec, []string, error) {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, err
	}

	specs := make(map[string]VarSpec, len(raw))
	order := make([]string, 0, len(raw))
	for k := range raw {
		order = append(order, k)
	}
	sort.Strings(order)

	for _, k := range order {
		v := raw[k]
		spec := VarSpec{Name: k, Prompt: k}
		switch t := v.(type) {
		case string, float64, bool:
			spec.Default = t
			spec.Kind = kindOf(t)
		case []any:
			spec.Kind = "choice"
			spec.Choices = toStringSlice(t)
			if len(spec.Choices) > 0 {
				spec.Default = spec.Choices[0]
			}
		case map[string]any:
			// Support {"default": "...", "choices": [...], "prompt": "..."}
			if p, ok := t["prompt"].(string); ok && p != "" {
				spec.Prompt = p
			}
			if ch, ok := t["choices"].([]any); ok {
				spec.Choices = toStringSlice(ch)
				spec.Kind = "choice"
			}
			if d, ok := t["default"]; ok {
				spec.Default = d
				if spec.Kind == "" {
					spec.Kind = kindOf(d)
				}
			}
			if spec.Kind == "" {
				spec.Kind = "any"
			}
		default:
			spec.Kind = "any"
		}
		specs[k] = spec
	}
	return specs, order, nil
}

func kindOf(v any) string {
	switch t := v.(type) {
	case string:
		// Check if string looks like a boolean
		if IsBooleanLikeString(t) {
			return "bool"
		}
		return "string"
	case bool:
		return "bool"
	case float64:
		return "number"
	default:
		return "any"
	}
}

// IsBooleanLikeString checks if a string represents a boolean value
func IsBooleanLikeString(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	switch lower {
	case "y", "n", "yes", "no", "true", "false", "1", "0":
		return true
	default:
		return false
	}
}

// AsBool converts various types to boolean
func AsBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		switch strings.ToLower(strings.TrimSpace(t)) {
		case "y", "yes", "true", "1":
			return true
		default:
			return false
		}
	case float64:
		return t != 0
	default:
		return false
	}
}

func toStringSlice(vs []any) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, fmt.Sprint(v))
	}
	return out
}
