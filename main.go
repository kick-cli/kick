package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/text/cases"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/huh"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

//go:embed _internal/README.txt
var internalFS embed.FS

const (
	cfgCookiecutter = "cookiecutter.json"
)

// Color constants
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
)

type VarSpec struct {
	Name    string
	Prompt  string
	Default any
	Choices []string
	Kind    string // "string","bool","number","choice","any"
}

func main() {
	if len(os.Args) < 2 || hasHelpFlag(os.Args[1:]) {
		usage()
		return
	}
	src := os.Args[1]
	out := "."
	if len(os.Args) >= 3 {
		out = os.Args[2]
	}

	templatePath, cleanup, err := resolveTemplate(src)
	if err != nil {
		fatal("resolve template: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	cfgPath := filepath.Join(templatePath, cfgCookiecutter)
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		fatal("read %s: %v", cfgPath, err)
	}

	specs, order, err := parseCookiecutterJSON(cfgData)
	if err != nil {
		fatal("parse config: %v", err)
	}

	values, err := promptValues(specs, order)
	if err != nil {
		fatal("prompt: %v", err)
	}

	// Root data for templates roughly mimics cookiecutter: .cookiecutter.<var>
	data := map[string]any{"cookiecutter": values}

	// Walk and render
	if err := renderTemplateTree(templatePath, out, data); err != nil {
		fatal("render: %v", err)
	}

	doneColor := fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Teal().RGB[0], mocha.Teal().RGB[1], mocha.Teal().RGB[2])
	fmt.Printf("%s✅ Done.%s\n", doneColor, colorReset)
}

func usage() {
	_, _ = fmt.Fprintf(os.Stdout, `cutr – minimal Cookiecutter-like generator in Go

Usage:
  cutr <template> [output_dir]

<template> can be:
  - local directory path
  - git URL (https/ssh) or something ending in .git (cloned in-process)

Template expects %s at the root with variables.

Example:
  cutr gh://my-org/service-template ./my-service
  cutr /path/to/template ./out

`, cfgCookiecutter)
	_ = internalFS // just to ensure embed keeps something; not required for runtime
}

func hasHelpFlag(args []string) bool {
	for _, a := range args {
		switch a {
		case "-h", "--help", "help":
			return true
		}
	}
	return false
}

func resolveTemplate(src string) (string, func(), error) {
	// Detect git-ish sources
	if isGitLike(src) {
		tmp, err := os.MkdirTemp("", "cutr-*")
		if err != nil {
			return "", nil, err
		}
		// Best-effort shallow clone
		_, err = git.PlainClone(tmp, false, &git.CloneOptions{
			URL:      normalizeGitURL(src),
			Progress: nil,
			Depth:    1,
		})
		if err != nil {
			if errors.Is(err, transport.ErrAuthenticationRequired) {
				return "", func() { _ = os.RemoveAll(tmp) }, fmt.Errorf("git auth required for %s", src)
			}
			return "", func() { _ = os.RemoveAll(tmp) }, err
		}
		return tmp, func() { _ = os.RemoveAll(tmp) }, nil
	}

	// Local path
	info, err := os.Stat(src)
	if err != nil {
		return "", nil, err
	}
	if !info.IsDir() {
		return "", nil, fmt.Errorf("template path must be a directory")
	}
	return src, nil, nil
}

func isGitLike(s string) bool {
	if strings.HasSuffix(s, ".git") {
		return true
	}
	if strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "ssh://") {
		return true
	}
	if strings.Contains(s, "git@") || strings.HasPrefix(s, "gh://") {
		return true
	}
	return false
}

func normalizeGitURL(s string) string {
	if after, ok := strings.CutPrefix(s, "gh://"); ok {
		// gh://owner/repo[/subdir][?ref=branch]
		rest := after
		parts := strings.Split(rest, "?")
		path := parts[0]
		return "https://github.com/" + strings.TrimSuffix(path, "/")
	}
	return s
}

func parseCookiecutterJSON(b []byte) (map[string]VarSpec, []string, error) {
	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, err
	}

	specs := make(map[string]VarSpec, len(raw))
	order := make([]string, 0, len(raw))
	for k := range raw {
		order = append(order, k)
	}
	sort.Strings(order)

	for _, k := range order {
		v := raw[k]
		spec := VarSpec{Name: k, Prompt: k}
		switch t := v.(type) {
		case string, float64, bool:
			spec.Default = t
			spec.Kind = kindOf(t)
		case []any:
			spec.Kind = "choice"
			spec.Choices = toStringSlice(t)
			if len(spec.Choices) > 0 {
				spec.Default = spec.Choices[0]
			}
		case map[string]any:
			// Support {"default": "...", "choices": [...], "prompt": "..."}
			if p, ok := t["prompt"].(string); ok && p != "" {
				spec.Prompt = p
			}
			if ch, ok := t["choices"].([]any); ok {
				spec.Choices = toStringSlice(ch)
				spec.Kind = "choice"
			}
			if d, ok := t["default"]; ok {
				spec.Default = d
				if spec.Kind == "" {
					spec.Kind = kindOf(d)
				}
			}
			if spec.Kind == "" {
				spec.Kind = "any"
			}
		default:
			spec.Kind = "any"
		}
		specs[k] = spec
	}
	return specs, order, nil
}

func kindOf(v any) string {
	switch t := v.(type) {
	case string:
		// Check if string looks like a boolean
		if isBooleanLikeString(t) {
			return "bool"
		}
		return "string"
	case bool:
		return "bool"
	case float64:
		return "number"
	default:
		return "any"
	}
}

func isBooleanLikeString(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	switch lower {
	case "y", "n", "yes", "no", "true", "false", "1", "0":
		return true
	default:
		return false
	}
}

func toStringSlice(vs []any) []string {
	out := make([]string, 0, len(vs))
	for _, v := range vs {
		out = append(out, fmt.Sprint(v))
	}
	return out
}

func promptValues(specs map[string]VarSpec, order []string) (map[string]any, error) {
	values := make(map[string]any, len(specs))

	// Print header once with colorful styling
	headerColor := fmt.Sprintf("\033[38;2;%d;%d;%dm", mocha.Maroon().RGB[0], mocha.Maroon().RGB[1], mocha.Maroon().RGB[2])
	fmt.Printf("%s🎯 Project Configuration%s\n", headerColor, colorReset)
	fmt.Println()

	// Process each prompt individually to avoid screen clearing
	for _, name := range order {
		s := specs[name]
		defStr := fmt.Sprint(s.Default)

		var val any
		var err error

		if len(s.Choices) > 0 {
			val, err = promptChoice(s, defStr)
		} else {
			switch s.Kind {
			case "bool":
				val, err = promptBool(s)
			case "number":
				val, err = promptNumber(s, defStr)
			default:
				val, err = promptString(s, defStr)
			}
		}

		if err != nil {
			// Fallback to simple prompts if Huh fails
			if strings.Contains(err.Error(), "TTY") || strings.Contains(err.Error(), "tty") {
				return promptValuesFallback(specs, order)
			}
			return nil, err
		}

		values[name] = val
		fmt.Printf("%s✓%s %s%s%s: %s%v%s\n\n",
			colorSuccess, colorReset,
			colorHistoryText, s.Prompt, colorReset,
			colorAnswer, val, colorReset) // Show colored question and answer
	}

	return values, nil
}

func promptChoice(spec VarSpec, defStr string) (any, error) {
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

func promptBool(spec VarSpec) (any, error) {
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

func promptNumber(spec VarSpec, defStr string) (any, error) {
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

func promptString(spec VarSpec, defStr string) (any, error) {
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

// Simple fallback prompts when Huh can't initialize (e.g., no TTY)
func promptValuesFallback(specs map[string]VarSpec, order []string) (map[string]any, error) {
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

		fmt.Printf("%s❯%s %s%s%s", colorPromptSymbol, colorReset, colorPromptText, s.Prompt, colorReset)
		if len(s.Choices) > 0 {
			fmt.Printf(" %s(choices: %s)%s", colorMuted, strings.Join(s.Choices, ", "), colorReset)
		} else if s.Kind == "bool" {
			fmt.Printf(" %s(choices: Yes, No)%s", colorMuted, colorReset)
		}
		fmt.Printf(" %s[default: %s]%s: ", colorSubtle, defStr, colorReset)

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
		fmt.Printf("%s✓%s %s%s%s: %s%v%s\n\n",
			colorSuccess, colorReset,
			colorHistoryText, s.Prompt, colorReset,
			colorAnswer, finalValue, colorReset) // Show colored question and answer
	}

	return values, nil
}

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

func renderTemplateTree(srcRoot, outRoot string, data map[string]any) error {
	// Make sure output exists
	if err := os.MkdirAll(outRoot, 0o755); err != nil {
		return err
	}

	return filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(srcRoot, path)
		if rel == "." {
			return nil
		}

		// Skip self files
		base := filepath.Base(path)
		if base == ".git" || base == cfgCookiecutter {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Render each path segment
		targetRel, err := renderPath(rel, data)
		if err != nil {
			return fmt.Errorf("render path %q: %w", rel, err)
		}
		// Skip empty results (if a segment renders to empty, drop it)
		if targetRel == "" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		targetPath := filepath.Join(outRoot, targetRel)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		// File: copy or render
		srcInfo, err := os.Stat(path)
		if err != nil {
			return err
		}
		mode := srcInfo.Mode()

		dataBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if looksBinary(dataBytes) {
			// Copy as-is
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return err
			}
			return os.WriteFile(targetPath, dataBytes, mode.Perm())
		}

		outBytes, err := renderBytes(dataBytes, data)
		if err != nil {
			return fmt.Errorf("render file %q: %w", rel, err)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, outBytes, mode.Perm())
	})
}

func renderPath(rel string, data map[string]any) (string, error) {
	segs := strings.Split(rel, string(os.PathSeparator))
	outSegs := make([]string, 0, len(segs))
	for _, s := range segs {
		trim := strings.TrimSpace(s)
		if trim == "" {
			continue
		}
		r, err := renderString(trim, data)
		if err != nil {
			return "", err
		}
		r = strings.TrimSpace(r)
		if r == "" {
			// Skip empty-rendered segments
			continue
		}
		outSegs = append(outSegs, r)
	}
	return filepath.Join(outSegs...), nil
}

func renderString(tmpl string, data map[string]any) (string, error) {
	t := template.New("str").Funcs(funcs())
	t, err := t.Option("missingkey=error").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderBytes(b []byte, data map[string]any) ([]byte, error) {
	t := template.New("file").Funcs(funcs())
	t, err := t.Option("missingkey=error").Parse(string(b))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func looksBinary(b []byte) bool {
	// Simple heuristic: if contains NUL or not valid UTF-8 and > 30% non-printables in first 1024 bytes
	n := min(len(b), 1024)
	sample := b[:n]
	if bytes.IndexByte(sample, 0) >= 0 {
		return true
	}
	if utf8.Valid(sample) {
		return false
	}
	var nonPrintable int
	for _, c := range sample {
		if c < 0x09 || (c > 0x0D && c < 0x20) {
			nonPrintable++
		}
	}
	return float64(nonPrintable)/float64(len(sample)) > 0.3
}

func funcs() template.FuncMap {
	return template.FuncMap{
		"upper": cases.Upper,
		"lower": cases.Lower,
		"title": cases.Title,
		"trim":  strings.TrimSpace,
		"snake": toSnake,
		"kebab": toKebab,
		"camel": toCamel,
		"pascal": func(s string) string {
			c := toCamel(s)
			if c == "" {
				return c
			}
			return strings.ToUpper(c[:1]) + c[1:]
		},
		"replace": strings.ReplaceAll,
	}
}

// naive casers for convenience
func toSnake(s string) string {
	s = strings.TrimSpace(s)
	var out []rune
	for i, r := range s {
		if r == ' ' || r == '-' {
			out = append(out, '_')
			continue
		}
		if i > 0 && isUpper(r) && (i+1 < len(s) && isLower(rune(s[i+1]))) {
			out = append(out, '_')
		}
		out = append(out, toLowerRune(r))
	}
	return strings.Trim(strings.ReplaceAll(string(out), "__", "_"), "_")
}

func toKebab(s string) string {
	return strings.ReplaceAll(toSnake(s), "_", "-")
}

func toCamel(s string) string {
	s = strings.TrimSpace(s)
	parts := splitWords(s)
	if len(parts) == 0 {
		return ""
	}
	var out strings.Builder
	for i, p := range parts {
		p = strings.ToLower(p)
		if i == 0 {
			out.WriteString(p)
			continue
		}
		out.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			out.WriteString(p[1:])
		}
	}
	return out.String()
}

func splitWords(s string) []string {
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	fields := strings.Fields(s)
	return fields
}

func isUpper(r rune) bool { return r >= 'A' && r <= 'Z' }
func isLower(r rune) bool { return r >= 'a' && r <= 'z' }
func toLowerRune(r rune) rune {
	if isUpper(r) {
		return r + ('a' - 'A')
	}
	return r
}

func fatal(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "error: "+format+"\n", a...)
	os.Exit(1)
}
