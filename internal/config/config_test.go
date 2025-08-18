package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCookiecutterJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantSpecs map[string]VarSpec
		wantOrder []string
		wantErr   bool
	}{
		// String variables
		{
			name:  "simple string variable",
			input: `{"project_name": "my-project"}`,
			wantSpecs: map[string]VarSpec{
				"project_name": {
					Name: "project_name", Prompt: "project_name", Default: "my-project", Kind: "string",
				},
			},
			wantOrder: []string{"project_name"},
		},
		{
			name:  "multiple strings sorted alphabetically",
			input: `{"z_last": "last", "a_first": "first", "m_middle": "middle"}`,
			wantSpecs: map[string]VarSpec{
				"a_first":  {Name: "a_first", Prompt: "a_first", Default: "first", Kind: "string"},
				"m_middle": {Name: "m_middle", Prompt: "m_middle", Default: "middle", Kind: "string"},
				"z_last":   {Name: "z_last", Prompt: "z_last", Default: "last", Kind: "string"},
			},
			wantOrder: []string{"a_first", "m_middle", "z_last"},
		},
		{
			name:  "regular strings not boolean-like",
			input: `{"name": "john", "description": "some text"}`,
			wantSpecs: map[string]VarSpec{
				"description": {Name: "description", Prompt: "description", Default: "some text", Kind: "string"},
				"name":        {Name: "name", Prompt: "name", Default: "john", Kind: "string"},
			},
			wantOrder: []string{"description", "name"},
		},

		// Boolean-like string variables
		{
			name:  "boolean-like string y",
			input: `{"use_feature": "y"}`,
			wantSpecs: map[string]VarSpec{
				"use_feature": {Name: "use_feature", Prompt: "use_feature", Default: "y", Kind: "bool"},
			},
			wantOrder: []string{"use_feature"},
		},
		{
			name:  "boolean-like string yes",
			input: `{"enable_logs": "yes"}`,
			wantSpecs: map[string]VarSpec{
				"enable_logs": {Name: "enable_logs", Prompt: "enable_logs", Default: "yes", Kind: "bool"},
			},
			wantOrder: []string{"enable_logs"},
		},
		{
			name:  "boolean-like string true",
			input: `{"debug_mode": "true"}`,
			wantSpecs: map[string]VarSpec{
				"debug_mode": {Name: "debug_mode", Prompt: "debug_mode", Default: "true", Kind: "bool"},
			},
			wantOrder: []string{"debug_mode"},
		},
		{
			name:  "boolean-like string n",
			input: `{"disable_cache": "n"}`,
			wantSpecs: map[string]VarSpec{
				"disable_cache": {Name: "disable_cache", Prompt: "disable_cache", Default: "n", Kind: "bool"},
			},
			wantOrder: []string{"disable_cache"},
		},
		{
			name:  "boolean-like mixed case",
			input: `{"production": "False"}`,
			wantSpecs: map[string]VarSpec{
				"production": {Name: "production", Prompt: "production", Default: "False", Kind: "bool"},
			},
			wantOrder: []string{"production"},
		},
		{
			name:  "boolean-like numeric strings",
			input: `{"enabled": "1", "disabled": "0"}`,
			wantSpecs: map[string]VarSpec{
				"disabled": {Name: "disabled", Prompt: "disabled", Default: "0", Kind: "bool"},
				"enabled":  {Name: "enabled", Prompt: "enabled", Default: "1", Kind: "bool"},
			},
			wantOrder: []string{"disabled", "enabled"},
		},

		// Actual boolean variables
		{
			name:  "JSON boolean values",
			input: `{"enabled": true, "disabled": false}`,
			wantSpecs: map[string]VarSpec{
				"disabled": {Name: "disabled", Prompt: "disabled", Default: false, Kind: "bool"},
				"enabled":  {Name: "enabled", Prompt: "enabled", Default: true, Kind: "bool"},
			},
			wantOrder: []string{"disabled", "enabled"},
		},

		// Number variables
		{
			name:  "integer and float numbers",
			input: `{"port": 8080, "timeout": 30.5}`,
			wantSpecs: map[string]VarSpec{
				"port":    {Name: "port", Prompt: "port", Default: float64(8080), Kind: "number"},
				"timeout": {Name: "timeout", Prompt: "timeout", Default: 30.5, Kind: "number"},
			},
			wantOrder: []string{"port", "timeout"},
		},

		// Choice variables
		{
			name:  "choice array with default",
			input: `{"database": ["mysql", "postgres", "sqlite"]}`,
			wantSpecs: map[string]VarSpec{
				"database": {
					Name: "database", Prompt: "database", Default: "mysql", Kind: "choice",
					Choices: []string{"mysql", "postgres", "sqlite"},
				},
			},
			wantOrder: []string{"database"},
		},
		{
			name:  "empty choice array",
			input: `{"empty_choices": []}`,
			wantSpecs: map[string]VarSpec{
				"empty_choices": {
					Name: "empty_choices", Prompt: "empty_choices", Kind: "choice",
					Choices: []string{},
				},
			},
			wantOrder: []string{"empty_choices"},
		},

		// Complex object variables
		{
			name:  "object with all fields",
			input: `{"version": {"default": "1.0.0", "prompt": "Application version", "choices": ["1.0.0", "2.0.0"]}}`,
			wantSpecs: map[string]VarSpec{
				"version": {
					Name: "version", Prompt: "Application version", Default: "1.0.0", Kind: "choice",
					Choices: []string{"1.0.0", "2.0.0"},
				},
			},
			wantOrder: []string{"version"},
		},
		{
			name:  "object with custom prompt only",
			input: `{"description": {"default": "My app", "prompt": "Enter app description"}}`,
			wantSpecs: map[string]VarSpec{
				"description": {Name: "description", Prompt: "Enter app description", Default: "My app", Kind: "string"},
			},
			wantOrder: []string{"description"},
		},
		{
			name:  "object with boolean-like default",
			input: `{"feature": {"default": "y", "prompt": "Enable feature?"}}`,
			wantSpecs: map[string]VarSpec{
				"feature": {Name: "feature", Prompt: "Enable feature?", Default: "y", Kind: "bool"},
			},
			wantOrder: []string{"feature"},
		},
		{
			name:  "object with choices but no default",
			input: `{"env": {"choices": ["dev", "prod"], "prompt": "Environment"}}`,
			wantSpecs: map[string]VarSpec{
				"env": {
					Name: "env", Prompt: "Environment", Kind: "choice",
					Choices: []string{"dev", "prod"},
				},
			},
			wantOrder: []string{"env"},
		},

		// Edge cases
		{
			name:  "empty object defaults to any",
			input: `{"unknown": {}}`,
			wantSpecs: map[string]VarSpec{
				"unknown": {Name: "unknown", Prompt: "unknown", Kind: "any"},
			},
			wantOrder: []string{"unknown"},
		},
		{
			name:  "null value defaults to any",
			input: `{"complex": null}`,
			wantSpecs: map[string]VarSpec{
				"complex": {Name: "complex", Prompt: "complex", Kind: "any"},
			},
			wantOrder: []string{"complex"},
		},
		{
			name:      "empty JSON object",
			input:     `{}`,
			wantSpecs: map[string]VarSpec{},
			wantOrder: []string{},
		},

		// Error cases
		{
			name:    "invalid JSON",
			input:   `{"invalid": json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specs, order, err := ParseCookiecutterJSON([]byte(tt.input))

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid character")
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSpecs, specs)
			assert.Equal(t, tt.wantOrder, order)
		})
	}
}
