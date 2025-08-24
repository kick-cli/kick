package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKickYAML(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantConfig    Config
		wantErr       bool
		errorContains string
	}{
		// Basic template configuration
		{
			name: "minimal template config",
			input: `name: "simple-template"
description: "A simple template"
version: "1.0.0"

variables:
  project_name:
    type: string
    prompt: "Project name"
    default: "my-project"`,
			wantConfig: Config{
				Name:        "simple-template",
				Description: "A simple template",
				Version:     "1.0.0",
				Variables: map[string]Variable{
					"project_name": {
						Type:    "string",
						Prompt:  "Project name",
						Default: "my-project",
					},
				},
			},
		},
		{
			name: "complete template metadata",
			input: `name: "go-service-template"
description: "Production-ready Go microservice template"
version: "1.2.0"
author: "John Doe <john@example.com>"
repository: "https://github.com/john/template"

variables:
  project_name:
    type: string
    prompt: "Project name"
    default: "my-service"`,
			wantConfig: Config{
				Name:        "go-service-template",
				Description: "Production-ready Go microservice template",
				Version:     "1.2.0",
				Author:      "John Doe <john@example.com>",
				Repository:  "https://github.com/john/template",
				Variables: map[string]Variable{
					"project_name": {
						Type:    "string",
						Prompt:  "Project name",
						Default: "my-service",
					},
				},
			},
		},

		// String variables
		{
			name: "string variable with validation",
			input: `name: "test"
variables:
  project_name:
    type: string
    prompt: "Project name"
    default: "my-project"
    pattern: "^[a-z][a-z0-9-]*$"
    help: "Must be lowercase"`,
			wantConfig: Config{
				Name: "test",
				Variables: map[string]Variable{
					"project_name": {
						Type:    "string",
						Prompt:  "Project name",
						Default: "my-project",
						Pattern: "^[a-z][a-z0-9-]*$",
						Help:    "Must be lowercase",
					},
				},
			},
		},

		// Choice variables
		{
			name: "choice variable",
			input: `name: "test"
variables:
  database:
    type: choice
    prompt: "Database type"
    choices: ["postgres", "mysql", "sqlite"]
    default: "postgres"
    help: "Choose your database"`,
			wantConfig: Config{
				Name: "test",
				Variables: map[string]Variable{
					"database": {
						Type:    "choice",
						Prompt:  "Database type",
						Choices: []string{"postgres", "mysql", "sqlite"},
						Default: "postgres",
						Help:    "Choose your database",
					},
				},
			},
		},

		// Boolean variables
		{
			name: "boolean variable",
			input: `name: "test"
variables:
  use_docker:
    type: boolean
    prompt: "Include Docker?"
    default: true`,
			wantConfig: Config{
				Name: "test",
				Variables: map[string]Variable{
					"use_docker": {
						Type:    "boolean",
						Prompt:  "Include Docker?",
						Default: true,
					},
				},
			},
		},

		// Number variables
		{
			name: "number variable with constraints",
			input: `name: "test"
variables:
  port:
    type: number
    prompt: "HTTP port"
    default: 8080
    min: 1024
    max: 65535`,
			wantConfig: Config{
				Name: "test",
				Variables: map[string]Variable{
					"port": {
						Type:    "number",
						Prompt:  "HTTP port",
						Default: 8080,
						Min:     1024,
						Max:     65535,
					},
				},
			},
		},

		// Multiple variables
		{
			name: "multiple variables of different types",
			input: `name: "multi-var-template"
variables:
  project_name:
    type: string
    prompt: "Project name"
    default: "my-app"

  database:
    type: choice
    prompt: "Database"
    choices: ["postgres", "mysql"]
    default: "postgres"

  enable_auth:
    type: boolean
    prompt: "Enable authentication?"
    default: false

  port:
    type: number
    prompt: "Port"
    default: 3000`,
			wantConfig: Config{
				Name: "multi-var-template",
				Variables: map[string]Variable{
					"project_name": {
						Type:    "string",
						Prompt:  "Project name",
						Default: "my-app",
					},
					"database": {
						Type:    "choice",
						Prompt:  "Database",
						Choices: []string{"postgres", "mysql"},
						Default: "postgres",
					},
					"enable_auth": {
						Type:    "boolean",
						Prompt:  "Enable authentication?",
						Default: false,
					},
					"port": {
						Type:    "number",
						Prompt:  "Port",
						Default: 3000,
					},
				},
			},
		},

		// Hooks configuration
		{
			name: "template with hooks",
			input: `name: "template-with-hooks"
variables:
  project_name:
    type: string
    default: "my-project"

hooks:
  pre_generation:
    - "echo 'Starting generation'"
    - "scripts/validate.sh"
  post_generation:
    - "go mod init {{.project_name}}"
    - "echo 'Done!'"`,
			wantConfig: Config{
				Name: "template-with-hooks",
				Variables: map[string]Variable{
					"project_name": {
						Type:    "string",
						Default: "my-project",
					},
				},
				Hooks: Hooks{
					PreGeneration:  []string{"echo 'Starting generation'", "scripts/validate.sh"},
					PostGeneration: []string{"go mod init {{.project_name}}", "echo 'Done!'"},
				},
			},
		},

		// Template settings
		{
			name: "template with settings",
			input: `name: "template-with-settings"
variables:
  name:
    type: string
    default: "test"

template:
  ignore_patterns:
    - "*.tmp"
    - ".DS_Store"
  keep_permissions: true`,
			wantConfig: Config{
				Name: "template-with-settings",
				Variables: map[string]Variable{
					"name": {
						Type:    "string",
						Default: "test",
					},
				},
				Template: TemplateSettings{
					IgnorePatterns:  []string{"*.tmp", ".DS_Store"},
					KeepPermissions: true,
				},
			},
		},

		// Error cases
		{
			name:          "invalid YAML",
			input:         `name: "test"\ninvalid: yaml: content`,
			wantErr:       true,
			errorContains: "yaml:",
		},
		{
			name: "missing required name",
			input: `description: "Missing name"
variables:
  test:
    type: string`,
			wantErr:       true,
			errorContains: "name is required",
		},
		{
			name: "invalid variable type",
			input: `name: "test"
variables:
  test:
    type: "invalid_type"`,
			wantErr:       true,
			errorContains: "invalid variable type",
		},
		{
			name: "choice variable without choices",
			input: `name: "test"
variables:
  test:
    type: choice
    prompt: "Choose"`,
			wantErr:       true,
			errorContains: "choices required for choice type",
		},
		{
			name: "number variable with invalid min/max",
			input: `name: "test"
variables:
  test:
    type: number
    min: 100
    max: 50`,
			wantErr:       true,
			errorContains: "min cannot be greater than max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseKickYAML([]byte(tt.input))

			if tt.wantErr {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)

			// Compare all fields except the private variableOrder field
			assert.Equal(t, tt.wantConfig.Name, config.Name)
			assert.Equal(t, tt.wantConfig.Description, config.Description)
			assert.Equal(t, tt.wantConfig.Version, config.Version)
			assert.Equal(t, tt.wantConfig.Author, config.Author)
			assert.Equal(t, tt.wantConfig.Repository, config.Repository)
			assert.Equal(t, tt.wantConfig.Variables, config.Variables)
			assert.Equal(t, tt.wantConfig.Hooks, config.Hooks)
			assert.Equal(t, tt.wantConfig.Template, config.Template)

			// Test variable order through public method - should preserve order from YAML
			variableOrder := config.GetVariableOrder()
			if len(config.Variables) > 0 {
				assert.NotEmpty(t, variableOrder, "Variable order should be preserved")
				assert.Equal(t, len(config.Variables), len(variableOrder), "Variable order should include all variables")
			}
		})
	}
}

func TestVariable_Validate(t *testing.T) {
	tests := []struct {
		name     string
		variable Variable
		value    any
		wantErr  bool
	}{
		// String validation
		{
			name: "valid string",
			variable: Variable{
				Type:    "string",
				Pattern: "^[a-z]+$",
			},
			value: "hello",
		},
		{
			name: "invalid string pattern",
			variable: Variable{
				Type:    "string",
				Pattern: "^[a-z]+$",
			},
			value:   "Hello123",
			wantErr: true,
		},
		{
			name: "string without pattern",
			variable: Variable{
				Type: "string",
			},
			value: "any-string-works",
		},

		// Choice validation
		{
			name: "valid choice",
			variable: Variable{
				Type:    "choice",
				Choices: []string{"postgres", "mysql", "sqlite"},
			},
			value: "postgres",
		},
		{
			name: "invalid choice",
			variable: Variable{
				Type:    "choice",
				Choices: []string{"postgres", "mysql", "sqlite"},
			},
			value:   "oracle",
			wantErr: true,
		},

		// Number validation
		{
			name: "valid number in range",
			variable: Variable{
				Type: "number",
				Min:  1024,
				Max:  65535,
			},
			value: 8080,
		},
		{
			name: "number below min",
			variable: Variable{
				Type: "number",
				Min:  1024,
				Max:  65535,
			},
			value:   500,
			wantErr: true,
		},
		{
			name: "number above max",
			variable: Variable{
				Type: "number",
				Min:  1024,
				Max:  65535,
			},
			value:   70000,
			wantErr: true,
		},

		// Boolean validation
		{
			name: "valid boolean true",
			variable: Variable{
				Type: "boolean",
			},
			value: true,
		},
		{
			name: "valid boolean false",
			variable: Variable{
				Type: "boolean",
			},
			value: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.variable.Validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetVariableOrder(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name: "preserves YAML definition order",
			input: `name: "test"
variables:
  project_name:
    type: string
    default: "my-project"

  author_name:
    type: string
    default: "John Doe"

  description:
    type: string
    default: "A test project"

  enable_feature:
    type: boolean
    default: true

  version:
    type: string
    default: "1.0.0"`,
			expected: []string{"project_name", "author_name", "description", "enable_feature", "version"},
		},
		{
			name: "single variable",
			input: `name: "test"
variables:
  project_name:
    type: string
    default: "my-project"`,
			expected: []string{"project_name"},
		},
		{
			name: "alphabetically unordered variables should preserve definition order",
			input: `name: "test"
variables:
  z_last:
    type: string
    default: "last"

  a_first:
    type: string
    default: "first"

  m_middle:
    type: string
    default: "middle"`,
			expected: []string{"z_last", "a_first", "m_middle"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := ParseKickYAML([]byte(tt.input))
			require.NoError(t, err)

			order := config.GetVariableOrder()
			assert.Equal(t, tt.expected, order, "Variable order should match YAML definition order")
		})
	}
}
