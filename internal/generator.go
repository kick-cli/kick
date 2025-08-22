package internal

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yarlson/tap"
)

// Options contains the configuration for template generation
type Options struct {
	Source    string // Template source (path or URL)
	OutputDir string // Output directory
}

// Generate performs the complete template generation workflow
func Generate(opts Options) error {
	// Resolve template source
	resolver := NewResolver()
	templatePath, cleanup, err := resolver.Resolve(opts.Source)
	if err != nil {
		return fmt.Errorf("resolve template: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Parse configuration
	cfg, err := loadConfig(templatePath)
	if err != nil {
		return err
	}

	// Collect user input
	values, err := CollectValues(cfg.Variables, cfg.GetVariableOrder())
	if err != nil {
		return fmt.Errorf("collect values: %v", err)
	}

	// Execute pre-generation hooks
	if err := executeHooks(cfg.Hooks.PreGeneration, "pre-generation", templatePath, values); err != nil {
		return err
	}

	// Generate files
	if err := generateFiles(templatePath, opts.OutputDir, values, cfg.Template); err != nil {
		return err
	}

	// Execute post-generation hooks
	if err := executeHooks(cfg.Hooks.PostGeneration, "post-generation", opts.OutputDir, values); err != nil {
		return err
	}

	// Success message
	tap.Outro("✓ Project scaffolded")
	return nil
}

// loadConfig loads and parses the template configuration
func loadConfig(templatePath string) (Config, error) {
	cfgPath := filepath.Join(templatePath, CutrYAML)
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %v", cfgPath, err)
	}

	cfg, err := ParseCutrYAML(cfgData)
	if err != nil {
		return Config{}, fmt.Errorf("parse config: %v", err)
	}

	return cfg, nil
}

// executeHooks runs pre or post generation hooks with progress display
func executeHooks(hookCommands []string, hookType, workDir string, data map[string]any) error {
	if len(hookCommands) == 0 {
		return nil
	}

	message := fmt.Sprintf("⚡ Executing %s hooks", hookType)
	// Capitalize first letter of hook type for display
	displayType := strings.ReplaceAll(hookType, "-", " ")
	if len(displayType) > 0 {
		displayType = strings.ToUpper(string(displayType[0])) + displayType[1:]
	}
	successMessage := fmt.Sprintf("%s hooks executed", displayType)
	return ShowProgress(message, successMessage, func() error {
		hookExecutor := New()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if hookType == "pre-generation" {
			return hookExecutor.ExecutePreGeneration(ctx, Hooks{PreGeneration: hookCommands}, workDir, data)
		} else {
			return hookExecutor.ExecutePostGeneration(ctx, Hooks{PostGeneration: hookCommands}, workDir, data)
		}
	})
}

// generateFiles renders the template tree with progress display
func generateFiles(templatePath, outputDir string, data map[string]any, settings TemplateSettings) error {
	return ShowProgress("📁 Rendering template files", "Template rendering complete", func() error {
		rend := NewRenderer()
		return rend.RenderTreeWithSettings(templatePath, outputDir, data, settings)
	})
}
