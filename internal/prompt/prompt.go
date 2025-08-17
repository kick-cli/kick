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
func Values(specs map[string]config.VarSpec, order []string) (map[string]any, error) {
	values := make(map[string]any, len(specs))

	// Print header once with colorful styling
	ui.PrintHeader("🎯 Project Configuration")

	// Process each prompt individually to avoid screen clearing
	for _, name := range order {
		s := specs[name]
		defStr := fmt.Sprint(s.Default)

		var val any
		var err error

		if len(s.Choices) > 0 {
			val, err = choice(s, defStr)
		} else {
			switch s.Kind {
			case "bool":
				val, err = boolean(s)
			case "number":
				val, err = number(s, defStr)
			default:
				val, err = text(s, defStr)
			}
		}

		if err != nil {
			// Fallback to simple prompts if Huh fails
			if strings.Contains(err.Error(), "TTY") || strings.Contains(err.Error(), "tty") {
				return fallback(specs, order)
			}
			return nil, err
		}

		values[name] = val
		ui.PrintHistory(s.Prompt, val)
	}

	return values, nil
}

func choice(spec config.VarSpec, defStr string) (any, error) {
	options := make([]huh.Option[string], len(spec.Choices))
	for i, choice := range spec.Choices {
		options[i] = huh.NewOption(choice, choice)
	}

	var selected = defStr
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(spec.Prompt).
				Options(options...).
				Value(&selected),
		),
	).WithAccessible(false)

	err := form.Run()
	return selected, err
}

func boolean(spec config.VarSpec) (any, error) {
	options := []huh.Option[string]{
		huh.NewOption("Yes", "true"),
		huh.NewOption("No", "false"),
	}

	// Set default based on spec.Default
	var selected = "false"
	if asBool(spec.Default) {
		selected = "true"
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(spec.Prompt).
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

func number(spec config.VarSpec, defStr string) (any, error) {
	var input = defStr
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(spec.Prompt).
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
		return spec.Default, nil
	}
	if n, err := strconv.ParseFloat(input, 64); err == nil {
		return n, nil
	}
	return spec.Default, nil
}

func text(spec config.VarSpec, defStr string) (any, error) {
	var input = defStr
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(spec.Prompt).
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
func fallback(specs map[string]config.VarSpec, order []string) (map[string]any, error) {
	values := make(map[string]any, len(specs))

	for _, name := range order {
		s := specs[name]
		defStr := fmt.Sprint(s.Default)

		// Format boolean default display
		if s.Kind == "bool" {
			if asBool(s.Default) {
				defStr = "Yes"
			} else {
				defStr = "No"
			}
		}

		ui.PrintPrompt(s.Prompt, s.Choices, s.Kind, defStr)

		var input string
		_, _ = fmt.Scanln(&input)
		input = strings.TrimSpace(input)

		var finalValue any
		if input == "" {
			if s.Kind == "bool" {
				finalValue = asBool(s.Default)
			} else {
				finalValue = s.Default
			}
		} else {
			switch s.Kind {
			case "bool":
				switch strings.ToLower(input) {
				case "y", "yes", "true", "1":
					finalValue = true
				case "n", "no", "false", "0":
					finalValue = false
				default:
					finalValue = asBool(s.Default)
				}
			case "number":
				if n, err := strconv.ParseFloat(input, 64); err == nil {
					finalValue = n
				} else {
					finalValue = s.Default
				}
			default:
				if len(s.Choices) > 0 {
					// Check if input matches a choice
					found := false
					for _, choice := range s.Choices {
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
		ui.PrintHistory(s.Prompt, finalValue)
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
