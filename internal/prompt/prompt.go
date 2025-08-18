package prompt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/yarlson/cutr/internal/config"
	"github.com/yarlson/cutr/internal/ui"
)

// Values prompts for and collects user input for template variables
func Values(variables map[string]config.Variable, order []string) (map[string]any, error) {
	values := make(map[string]any, len(variables))

	// Print header once with colorful styling
	ui.PrintHeader("🎯 Project Configuration")

	// Process each prompt individually to avoid screen clearing
	for _, name := range order {
		v := variables[name]
		defStr := fmt.Sprint(v.Default)

		var val any
		var err error

		if len(v.Choices) > 0 {
			val, err = choice(v, defStr)
		} else {
			switch v.Type {
			case "boolean":
				val, err = boolean(v)
			case "number":
				val, err = number(v, defStr)
			default:
				val, err = text(v, defStr)
			}
		}

		if err != nil {
			// Fallback to simple prompts if Huh fails
			if strings.Contains(err.Error(), "TTY") || strings.Contains(err.Error(), "tty") {
				return fallback(variables, order)
			}
			return nil, err
		}

		values[name] = val
		ui.PrintHistory(v.Prompt, val)
	}

	return values, nil
}

func choice(variable config.Variable, defStr string) (any, error) {
	options := make([]huh.Option[string], len(variable.Choices))
	for i, choice := range variable.Choices {
		options[i] = huh.NewOption(choice, choice)
	}

	var selected = defStr
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(variable.Prompt).
				Options(options...).
				Value(&selected),
		),
	).WithAccessible(false)

	err := form.Run()
	return selected, err
}

func boolean(variable config.Variable) (any, error) {
	options := []huh.Option[string]{
		huh.NewOption("Yes", "true"),
		huh.NewOption("No", "false"),
	}

	// Set default based on variable.Default
	var selected = "false"
	if asBool(variable.Default) {
		selected = "true"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(variable.Prompt).
				Options(options...).
				Value(&selected),
		),
	).WithAccessible(false)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Convert back to boolean
	return selected == "true", nil
}

func number(variable config.Variable, defStr string) (any, error) {
	var input = defStr
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(variable.Prompt).
				Value(&input).
				Placeholder(defStr),
		),
	).WithAccessible(false)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	// Convert to number
	if input == "" {
		return variable.Default, nil
	}
	if n, err := strconv.ParseFloat(input, 64); err == nil {
		return n, nil
	}
	return variable.Default, nil
}

func text(variable config.Variable, defStr string) (any, error) {
	var input = defStr
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(variable.Prompt).
				Value(&input).
				Placeholder(defStr),
		),
	).WithAccessible(false)

	err := form.Run()
	if err != nil {
		return nil, err
	}

	if input == "" {
		return defStr, nil
	}
	return input, nil
}

// fallback provides simple prompts when Huh can't initialize (e.g., no TTY)
func fallback(variables map[string]config.Variable, order []string) (map[string]any, error) {
	values := make(map[string]any, len(variables))

	for _, name := range order {
		v := variables[name]
		defStr := fmt.Sprint(v.Default)

		// Format boolean default display
		if v.Type == "boolean" {
			if asBool(v.Default) {
				defStr = "Yes"
			} else {
				defStr = "No"
			}
		}

		ui.PrintPrompt(v.Prompt, v.Choices, v.Type, defStr)

		var input string
		_, _ = fmt.Scanln(&input)
		input = strings.TrimSpace(input)

		var finalValue any
		if input == "" {
			if v.Type == "boolean" {
				finalValue = asBool(v.Default)
			} else {
				finalValue = v.Default
			}
		} else {
			switch v.Type {
			case "boolean":
				switch strings.ToLower(input) {
				case "y", "yes", "true", "1":
					finalValue = true
				case "n", "no", "false", "0":
					finalValue = false
				default:
					finalValue = asBool(v.Default)
				}
			case "number":
				if n, err := strconv.ParseFloat(input, 64); err == nil {
					finalValue = n
				} else {
					finalValue = v.Default
				}
			default:
				if len(v.Choices) > 0 {
					// Check if input matches a choice
					found := false
					for _, choice := range v.Choices {
						if input == choice {
							finalValue = choice
							found = true
							break
						}
					}
					if !found {
						finalValue = defStr
					}
				} else {
					finalValue = input
				}
			}
		}

		values[name] = finalValue
		ui.PrintHistory(v.Prompt, finalValue)
	}

	return values, nil
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
