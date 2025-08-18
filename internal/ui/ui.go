package ui

import (
	"fmt"
	"strings"

	catppuccin "github.com/catppuccin/go"
)

const (
	colorReset = "\033[0m"
)

// Catppuccin Mocha theme for beautiful CLI styling - MAXIMUM VIBRANCY
var (
	mocha = catppuccin.Mocha

	// Main colors for prompts and UI - most vibrant Catppuccin colors
	colorPromptSymbol = fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Rosewater().RGB[0], mocha.Rosewater().RGB[1], mocha.Rosewater().RGB[2]) // ❯ symbol - bright rosewater pink
	colorPromptText   = fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Sky().RGB[0], mocha.Sky().RGB[1], mocha.Sky().RGB[2])                   // question text - vivid sky blue
	colorMuted        = fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Lavender().RGB[0], mocha.Lavender().RGB[1], mocha.Lavender().RGB[2])    // choices, meta - bright lavender
	colorSubtle       = fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Yellow().RGB[0], mocha.Yellow().RGB[1], mocha.Yellow().RGB[2])          // defaults - bright yellow

	// History colors - most vibrant variety
	colorSuccess     = fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Green().RGB[0], mocha.Green().RGB[1], mocha.Green().RGB[2]) // ✓ checkmark - bright green
	colorHistoryText = fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Pink().RGB[0], mocha.Pink().RGB[1], mocha.Pink().RGB[2])    // completed questions - hot pink
	colorAnswer      = fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Peach().RGB[0], mocha.Peach().RGB[1], mocha.Peach().RGB[2]) // user answers - bright peach

	// Special colors
	colorHeader = fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Maroon().RGB[0], mocha.Maroon().RGB[1], mocha.Maroon().RGB[2]) // header text
	colorDone   = fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Teal().RGB[0], mocha.Teal().RGB[1], mocha.Teal().RGB[2])       // completion message
)

// PrintHeader prints a colored header message
func PrintHeader(text string) {
	fmt.Printf("%s%s%s\n", colorHeader, text, colorReset)
	fmt.Println()
}

// PrintHistory displays a completed question and answer in color
func PrintHistory(prompt string, answer any) {
	fmt.Printf("%s✓%s %s%s%s: %s%v%s\n\n",
		colorSuccess, colorReset,
		colorHistoryText, prompt, colorReset,
		colorAnswer, answer, colorReset,
	)
}

// PrintPrompt displays a fallback prompt for non-TTY environments
func PrintPrompt(prompt string, choices []string, kind, defStr string) {
	fmt.Printf("%s❯%s %s%s%s", colorPromptSymbol, colorReset, colorPromptText, prompt, colorReset)

	if len(choices) > 0 {
		fmt.Printf(" %s(choices: %s)%s", colorMuted, strings.Join(choices, ", "), colorReset)
	} else if kind == "bool" {
		fmt.Printf(" %s(choices: Yes, No)%s", colorMuted, colorReset)
	}

	fmt.Printf(" %s[default: %s]%s: ", colorSubtle, defStr, colorReset)
}

// PrintSuccess prints a colored success message
func PrintSuccess(text string) {
	fmt.Printf("%s%s%s\n", colorDone, text, colorReset)
}
