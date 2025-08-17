package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode/utf8"

	"golang.org/x/text/cases"

	"github.com/yarlson/cutr/internal/config"
)

// Renderer handles template rendering operations
type Renderer struct{}

// New creates a new template renderer
func New() *Renderer {
	return &Renderer{}
}

// RenderTree walks the source template directory and renders all files to the output directory
func (r *Renderer) RenderTree(srcRoot, outRoot string, data map[string]any) error {
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
		if base == ".git" || base == config.CookiecutterJSON {
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

		outBytes, err := r.renderBytes(dataBytes, data)
		if err != nil {
			return fmt.Errorf("render file %q: %w", rel, err)
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(targetPath, outBytes, mode.Perm())
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

func (r *Renderer) renderString(tmpl string, data map[string]any) (string, error) {
	t := template.New("str").Funcs(templateFuncs())
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

func (r *Renderer) renderBytes(b []byte, data map[string]any) ([]byte, error) {
	t := template.New("file").Funcs(templateFuncs())
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

func templateFuncs() template.FuncMap {
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
