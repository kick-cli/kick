package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yarlson/tap"
)

// CollectValues prompts for and collects user input for template variables
func CollectValues(variables map[string]Variable, order []string) (map[string]any, error) {
	values := make(map[string]any, len(variables))

	// Project scaffolding intro
	tap.Intro("🏗️  Project Scaffolding")

	// Process each variable in order
	for _, name := range order {
		variable := variables[name]
		defStr := fmt.Sprint(variable.Default)

		var result any
		var err error

		// Handle different variable types
		if len(variable.Choices) > 0 {
			result, err = promptChoice(variable, defStr)
		} else {
			switch variable.Type {
			case "boolean":
				result, err = promptBoolean(variable)
			case "number":
				result, err = promptNumber(variable, defStr)
			default:
				result, err = promptText(variable, defStr)
			}
		}

		if err != nil {
			return nil, err
		}

		values[name] = result
	}

	return values, nil
}

// promptChoice handles selection from predefined choices
func promptChoice(variable Variable, defStr string) (any, error) {
	options := make([]tap.SelectOption[string], len(variable.Choices))
	for i, choice := range variable.Choices {
		options[i] = tap.SelectOption[string]{
			Value: choice,
			Label: choice,
		}
	}

	// Set initial value if default matches a choice
	var initialValue *string
	for _, choice := range variable.Choices {
		if choice == defStr {
			initialValue = &choice
			break
		}
	}

	return tap.Select(tap.SelectOptions[string]{
		Message:      variable.Prompt,
		Options:      options,
		InitialValue: initialValue,
	}), nil
}

// promptBoolean handles yes/no prompts
func promptBoolean(variable Variable) (any, error) {
	initialValue := asBool(variable.Default)

	return tap.Confirm(tap.ConfirmOptions{
		Message:      variable.Prompt,
		Active:       "Yes",
		Inactive:     "No",
		InitialValue: initialValue,
	}), nil
}

// promptNumber handles numeric input with validation
func promptNumber(variable Variable, defStr string) (any, error) {
	input := tap.Text(tap.TextOptions{
		Message:      variable.Prompt,
		Placeholder:  defStr,
		DefaultValue: defStr,
		Validate: func(input string) error {
			if input == "" {
				return nil // Allow empty input to use default
			}
			if _, err := strconv.ParseFloat(input, 64); err != nil {
				return fmt.Errorf("invalid numeric value")
			}
			return nil
		},
	})

	// Convert to number
	if input == "" {
		return variable.Default, nil
	}
	if n, err := strconv.ParseFloat(input, 64); err == nil {
		return n, nil
	}
	return variable.Default, nil
}

// promptText handles text input with optional pattern validation
func promptText(variable Variable, defStr string) (any, error) {
	var validate func(string) error
	if variable.Pattern != "" {
		validate = func(input string) error {
			if input == "" {
				return nil // Allow empty input to use default
			}
			matched, err := regexp.MatchString(variable.Pattern, input)
			if err != nil {
				return fmt.Errorf("pattern validation error: %v", err)
			}
			if !matched {
				if variable.Help != "" {
					return fmt.Errorf("%s", variable.Help)
				}
				return fmt.Errorf("value does not match pattern: %s", variable.Pattern)
			}
			return nil
		}
	}

	input := tap.Text(tap.TextOptions{
		Message:      variable.Prompt,
		Placeholder:  defStr,
		DefaultValue: defStr,
		Validate:     validate,
	})
	if input == "" {
		return defStr, nil
	}
	return input, nil
}

// asBool converts various types to boolean
func asBool(v any) bool {
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
