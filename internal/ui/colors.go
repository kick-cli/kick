package ui

import (
	"fmt"

	catppuccin "github.com/catppuccin/go"
)

const ColorReset = "\033[0m"

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
