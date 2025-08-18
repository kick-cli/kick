package config

import (
	"fmt"
	"regexp"
	"slices"
	"sort"

	"gopkg.in/yaml.v3"
)

const CutrYAML = "cutr.yaml"

// Config represents the complete cutr template configuration
type Config struct {
	// Template metadata
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Version     string `yaml:"version,omitempty"`
	Author      string `yaml:"author,omitempty"`
	Repository  string `yaml:"repository,omitempty"`

	// Template variables
	Variables map[string]Variable `yaml:"variables"`

	// Optional hooks
	Hooks Hooks `yaml:"hooks,omitempty"`

	// Template settings
	Template TemplateSettings `yaml:"template,omitempty"`
}

// Variable represents a template variable definition
type Variable struct {
	Type    string   `yaml:"type"`
	Prompt  string   `yaml:"prompt,omitempty"`
	Default any      `yaml:"default,omitempty"`
	Choices []string `yaml:"choices,omitempty"`
	Pattern string   `yaml:"pattern,omitempty"`
	Help    string   `yaml:"help,omitempty"`
	Min     int      `yaml:"min,omitempty"`
	Max     int      `yaml:"max,omitempty"`
}

// Hooks defines pre and post generation commands
type Hooks struct {
	PreGeneration  []string `yaml:"pre_generation,omitempty"`
	PostGeneration []string `yaml:"post_generation,omitempty"`
}

// TemplateSettings defines template engine configuration
type TemplateSettings struct {
	IgnorePatterns  []string `yaml:"ignore_patterns,omitempty"`
	KeepPermissions bool     `yaml:"keep_permissions,omitempty"`
}

// ParseCutrYAML parses a cutr.yaml configuration file
func ParseCutrYAML(data []byte) (Config, error) {
	var config Config

	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("yaml: %w", err)
	}

	// Validate required fields
	if config.Name == "" {
		return Config{}, fmt.Errorf("name is required")
	}

	// Validate variables
	for name, variable := range config.Variables {
		if err := validateVariable(name, variable); err != nil {
			return Config{}, fmt.Errorf("variable %q: %w", name, err)
		}
	}

	return config, nil
}

// GetVariableOrder returns variable names in alphabetical order
func (c Config) GetVariableOrder() []string {
	order := make([]string, 0, len(c.Variables))
	for name := range c.Variables {
		order = append(order, name)
	}
	sort.Strings(order)
	return order
}

// Validate validates a value against the variable constraints
func (v Variable) Validate(value any) error {
	switch v.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

		if v.Pattern != "" {
			matched, err := regexp.MatchString(v.Pattern, str)
			if err != nil {
				return fmt.Errorf("invalid pattern: %w", err)
			}
			if !matched {
				return fmt.Errorf("value %q does not match pattern %q", str, v.Pattern)
			}
		}

	case "choice":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for choice, got %T", value)
		}

		if slices.Contains(v.Choices, str) {
			return nil
		}

		return fmt.Errorf("value %q is not a valid choice, must be one of %v", str, v.Choices)

	case "number":
		var num float64
		switch n := value.(type) {
		case int:
			num = float64(n)
		case float64:
			num = n
		default:
			return fmt.Errorf("expected number, got %T", value)
		}

		if v.Min != 0 && num < float64(v.Min) {
			return fmt.Errorf("value %g is below minimum %d", num, v.Min)
		}

		if v.Max != 0 && num > float64(v.Max) {
			return fmt.Errorf("value %g is above maximum %d", num, v.Max)
		}

	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	}

	return nil
}

func validateVariable(_ string, variable Variable) error {
	// Validate variable type
	validTypes := map[string]bool{
		"string":  true,
		"choice":  true,
		"number":  true,
		"boolean": true,
	}

	if !validTypes[variable.Type] {
		return fmt.Errorf("invalid variable type %q, must be one of [string, choice, number, boolean]", variable.Type)
	}

	// Type-specific validation
	switch variable.Type {
	case "choice":
		if len(variable.Choices) == 0 {
			return fmt.Errorf("choices required for choice type")
		}

	case "number":
		if variable.Min != 0 && variable.Max != 0 && variable.Min > variable.Max {
			return fmt.Errorf("min cannot be greater than max")
		}

	case "string":
		if variable.Pattern != "" {
			_, err := regexp.Compile(variable.Pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern: %w", err)
			}
		}
	}

	return nil
}
