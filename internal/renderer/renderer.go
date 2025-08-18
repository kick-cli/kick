package renderer

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/yarlson/cutr/internal/config"
)

// Renderer handles template rendering operations.
type Renderer struct {
	// funcMap is cached to avoid recreating it for each template
	funcMap template.FuncMap
}

// New creates a new template renderer.
func New() *Renderer {
	return &Renderer{
		funcMap: newTemplateFuncs(),
	}
}

// RenderTree walks the source template directory and renders all files to the output directory.
func (r *Renderer) RenderTree(srcRoot, outRoot string, data map[string]any) error {
	// Make sure output exists
	if err := os.MkdirAll(outRoot, 0o755); err != nil {
		return err
	}

	return filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return fmt.Errorf("compute relative path: %w", err)
		}
		if rel == "." {
			return nil
		}

		// Skip version control and config files
		if r.shouldSkip(filepath.Base(path), d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Render each path segment
		targetRel, err := r.renderPath(rel, data)
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

		// Process file: copy binary files as-is, render text files
		return r.processFile(path, targetPath, data)
	})
}

// RenderTreeWithSettings walks the source template directory and renders all files to the output directory
// using the provided template settings.
func (r *Renderer) RenderTreeWithSettings(srcRoot, outRoot string, data map[string]any, settings config.TemplateSettings) error {
	// Make sure output exists
	if err := os.MkdirAll(outRoot, 0o755); err != nil {
		return err
	}

	return filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return fmt.Errorf("compute relative path: %w", err)
		}
		if rel == "." {
			return nil
		}

		// Skip version control and config files
		if r.shouldSkip(filepath.Base(path), d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check ignore patterns
		if r.shouldIgnoreWithSettings(rel, d.IsDir(), settings) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Render each path segment
		targetRel, err := r.renderPath(rel, data)
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

		// Process file with settings
		return r.processFileWithSettings(path, targetPath, data, settings)
	})
}

func (r *Renderer) renderPath(rel string, data map[string]any) (string, error) {
	segs := strings.Split(rel, string(os.PathSeparator))
	outSegs := make([]string, 0, len(segs))
	for _, s := range segs {
		trim := strings.TrimSpace(s)
		if trim == "" {
			continue
		}
		rendered, err := r.renderString(trim, data)
		if err != nil {
			return "", err
		}
		rendered = strings.TrimSpace(rendered)
		if rendered == "" {
			// Skip empty-rendered segments
			continue
		}
		outSegs = append(outSegs, rendered)
	}
	return filepath.Join(outSegs...), nil
}

// shouldSkip determines if a file or directory should be skipped during rendering.
func (r *Renderer) shouldSkip(basename string, _ bool) bool {
	return basename == ".git" || basename == config.CutrYAML
}

// processFile handles copying binary files or rendering text files.
func (r *Renderer) processFile(srcPath, targetPath string, data map[string]any) error {
	// Get file info for permissions
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	mode := srcInfo.Mode()

	// Read file content
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	// Binary files are copied as-is, text files are rendered
	if isBinary(content) {
		return os.WriteFile(targetPath, content, mode.Perm())
	}

	// Render text file
	rendered, err := r.renderBytes(content, data)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	return os.WriteFile(targetPath, rendered, mode.Perm())
}

func (r *Renderer) renderString(tmpl string, data map[string]any) (string, error) {
	t, err := template.New("str").
		Funcs(r.funcMap).
		Option("missingkey=error").
		Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	return buf.String(), nil
}

func (r *Renderer) renderBytes(b []byte, data map[string]any) ([]byte, error) {
	t, err := template.New("file").
		Funcs(r.funcMap).
		Option("missingkey=error").
		Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	return buf.Bytes(), nil
}

// isBinary detects if data appears to be binary using heuristics.
// It checks for null bytes and counts non-printable characters.
func isBinary(data []byte) bool {
	const (
		maxSampleSize     = 1024
		nonPrintableRatio = 0.3
	)

	// Check a sample of the data
	sampleSize := min(len(data), maxSampleSize)
	sample := data[:sampleSize]

	// Binary files often contain null bytes
	if bytes.IndexByte(sample, 0) >= 0 {
		return true
	}

	// If it's valid UTF-8, it's likely text
	if utf8.Valid(sample) {
		return false
	}

	// Count non-printable characters
	nonPrintable := 0
	for _, b := range sample {
		if !isPrintableOrWhitespace(b) {
			nonPrintable++
		}
	}

	// If more than 30% non-printable, consider it binary
	return float64(nonPrintable)/float64(len(sample)) > nonPrintableRatio
}

// newTemplateFuncs creates the template function map.
func newTemplateFuncs() template.FuncMap {
	caser := cases.Title(language.English)
	return template.FuncMap{
		"upper":   func(s string) string { return cases.Upper(language.English).String(s) },
		"lower":   func(s string) string { return cases.Lower(language.English).String(s) },
		"title":   func(s string) string { return caser.String(s) },
		"trim":    strings.TrimSpace,
		"snake":   toSnakeCase,
		"kebab":   toKebabCase,
		"camel":   toCamelCase,
		"pascal":  toPascalCase,
		"replace": strings.ReplaceAll,
	}
}

// Case conversion functions

// toSnakeCase converts a string to snake_case.
func toSnakeCase(s string) string {
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

// toKebabCase converts a string to kebab-case.
func toKebabCase(s string) string {
	return strings.ReplaceAll(toSnakeCase(s), "_", "-")
}

// toCamelCase converts a string to camelCase.
func toCamelCase(s string) string {
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

// toPascalCase converts a string to PascalCase.
func toPascalCase(s string) string {
	camel := toCamelCase(s)
	if camel == "" {
		return ""
	}
	return strings.ToUpper(camel[:1]) + camel[1:]
}

// isPrintableOrWhitespace checks if a byte is printable or whitespace.
func isPrintableOrWhitespace(b byte) bool {
	return b >= 0x20 && b <= 0x7E || b == 0x09 || b == 0x0A || b == 0x0D
}

// Helper functions

// isUpper checks if a rune is uppercase.
func isUpper(r rune) bool { return unicode.IsUpper(r) }

// isLower checks if a rune is lowercase.
func isLower(r rune) bool { return unicode.IsLower(r) }

// toLowerRune converts a rune to lowercase.
func toLowerRune(r rune) rune { return unicode.ToLower(r) }

// shouldIgnoreWithSettings checks if a file should be ignored based on template settings.
func (r *Renderer) shouldIgnoreWithSettings(relPath string, isDir bool, settings config.TemplateSettings) bool {
	basename := filepath.Base(relPath)
	
	for _, pattern := range settings.IgnorePatterns {
		// Check if pattern matches the basename
		if matched, _ := filepath.Match(pattern, basename); matched {
			return true
		}
		// Check if pattern matches the full relative path
		if matched, _ := filepath.Match(pattern, relPath); matched {
			return true
		}
		// For directories, also check if the pattern matches the directory name exactly
		if isDir && pattern == basename {
			return true
		}
	}
	
	return false
}

// processFileWithSettings handles copying binary files or rendering text files with template settings.
func (r *Renderer) processFileWithSettings(srcPath, targetPath string, data map[string]any, settings config.TemplateSettings) error {
	// Get file info for permissions
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	mode := srcInfo.Mode()

	// Read file content
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create target directory: %w", err)
	}

	// Determine file permissions
	var targetMode os.FileMode
	if settings.KeepPermissions {
		targetMode = mode.Perm()
	} else {
		targetMode = 0644 // Default permissions
	}

	// Binary files are copied as-is, text files are rendered
	if isBinary(content) {
		return os.WriteFile(targetPath, content, targetMode)
	}

	// Render text file
	rendered, err := r.renderBytes(content, data)
	if err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	return os.WriteFile(targetPath, rendered, targetMode)
}
