package ui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/yarlson/cutr/internal/config"
	"github.com/yarlson/tap/core"
	"github.com/yarlson/tap/prompts"
	"github.com/yarlson/tap/terminal"
)

// Collector handles prompting for template variables
type Collector struct {
	term *terminal.Terminal
}

// New creates a new prompt collector
func New(term *terminal.Terminal) *Collector {
	return &Collector{term: term}
}

// CollectValues prompts for and collects user input for template variables
func (c *Collector) CollectValues(variables map[string]config.Variable, order []string) (map[string]any, error) {
	values := make(map[string]any, len(variables))

	// Project scaffolding intro
	prompts.Intro("🏗️  Project Scaffolding", prompts.MessageOptions{Output: c.term.Writer})

	// Process each variable in order
	for _, name := range order {
		variable := variables[name]
		defStr := fmt.Sprint(variable.Default)

		var result any
		var err error

		// Handle different variable types
		if len(variable.Choices) > 0 {
			result, err = c.promptChoice(variable, defStr)
		} else {
			switch variable.Type {
			case "boolean":
				result, err = c.promptBoolean(variable)
			case "number":
				result, err = c.promptNumber(variable, defStr)
			default:
				result, err = c.promptText(variable, defStr)
			}
		}

		if err != nil {
			return nil, err
		}

		// Check for cancellation
		if core.IsCancel(result) {
			prompts.Cancel("Generation aborted", prompts.MessageOptions{Output: c.term.Writer})
			return nil, fmt.Errorf("user cancelled")
		}

		values[name] = result
	}

	return values, nil
}

// promptChoice handles selection from predefined choices
func (c *Collector) promptChoice(variable config.Variable, defStr string) (any, error) {
	options := make([]prompts.SelectOption[string], len(variable.Choices))
	for i, choice := range variable.Choices {
		options[i] = prompts.SelectOption[string]{
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

	return prompts.Select(prompts.SelectOptions[string]{
		Message:      variable.Prompt,
		Options:      options,
		InitialValue: initialValue,
		Input:        c.term.Reader,
		Output:       c.term.Writer,
	}), nil
}

// promptBoolean handles yes/no prompts
func (c *Collector) promptBoolean(variable config.Variable) (any, error) {
	initialValue := asBool(variable.Default)

	return prompts.Confirm(prompts.ConfirmOptions{
		Message:      variable.Prompt,
		Active:       "Yes",
		Inactive:     "No",
		InitialValue: initialValue,
		Input:        c.term.Reader,
		Output:       c.term.Writer,
	}), nil
}

// promptNumber handles numeric input with validation
func (c *Collector) promptNumber(variable config.Variable, defStr string) (any, error) {
	result := prompts.Text(prompts.TextOptions{
		Message:      variable.Prompt,
		Placeholder:  defStr,
		DefaultValue: defStr,
		Input:        c.term.Reader,
		Output:       c.term.Writer,
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

	if core.IsCancel(result) {
		return result, nil
	}

	input := result.(string)

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
func (c *Collector) promptText(variable config.Variable, defStr string) (any, error) {
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

	result := prompts.Text(prompts.TextOptions{
		Message:      variable.Prompt,
		Placeholder:  defStr,
		DefaultValue: defStr,
		Input:        c.term.Reader,
		Output:       c.term.Writer,
		Validate:     validate,
	})

	if core.IsCancel(result) {
		return result, nil
	}

	input := result.(string)
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
