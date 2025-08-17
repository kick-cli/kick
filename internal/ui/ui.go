package ui

import (
	"fmt"
	"strings"
)

// PrintHeader prints a colored header message
func PrintHeader(text string) {
	fmt.Printf("%s%s%s\n", colorHeader, text, ColorReset)
	fmt.Println()
}

// PrintHistory displays a completed question and answer in color
func PrintHistory(prompt string, answer any) {
	fmt.Printf("%s✓%s %s%s%s: %s%v%s\n\n",
		colorSuccess, ColorReset,
		colorHistoryText, prompt, ColorReset,
		colorAnswer, answer, ColorReset)
}

// PrintPrompt displays a fallback prompt for non-TTY environments
func PrintPrompt(prompt string, choices []string, kind, defStr string) {
	fmt.Printf("%s❯%s %s%s%s", colorPromptSymbol, ColorReset, colorPromptText, prompt, ColorReset)
	if len(choices) > 0 {
		fmt.Printf(" %s(choices: %s)%s", colorMuted, strings.Join(choices, ", "), ColorReset)
	} else if kind == "bool" {
		fmt.Printf(" %s(choices: Yes, No)%s", colorMuted, ColorReset)
	}
	fmt.Printf(" %s[default: %s]%s: ", colorSubtle, defStr, ColorReset)
}

// PrintSuccess prints a colored success message
func PrintSuccess(text string) {
	fmt.Printf("%s%s%s\n", colorDone, text, ColorReset)
}
