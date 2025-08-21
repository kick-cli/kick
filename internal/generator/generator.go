package generator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/yarlson/cutr/internal/config"
	"github.com/yarlson/cutr/internal/hooks"
	"github.com/yarlson/cutr/internal/renderer"
	"github.com/yarlson/cutr/internal/source"
	"github.com/yarlson/cutr/internal/ui"
	"github.com/yarlson/tap/prompts"
	"github.com/yarlson/tap/terminal"
)

// Options contains the configuration for template generation
type Options struct {
	Source    string // Template source (path or URL)
	OutputDir string // Output directory
}

// Generator orchestrates the template generation workflow
type Generator struct {
	term            *terminal.Terminal
	promptCollector *ui.Collector
	progressHandler *ui.ProgressHandler
}

// New creates a new generator instance
func New() (*Generator, error) {
	term, err := terminal.New()
	if err != nil {
		return nil, fmt.Errorf("initialize terminal: %v", err)
	}

	return &Generator{
		term:            term,
		promptCollector: ui.New(term),
		progressHandler: ui.NewProgressHandler(term),
	}, nil
}

// Close releases terminal resources
func (g *Generator) Close() {
	if g.term != nil {
		g.term.Close()
	}
}

// Generate performs the complete template generation workflow
func (g *Generator) Generate(opts Options) error {
	// Resolve template source
	resolver := source.New()
	templatePath, cleanup, err := resolver.Resolve(opts.Source)
	if err != nil {
		return fmt.Errorf("resolve template: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Parse configuration
	cfg, err := g.loadConfig(templatePath)
	if err != nil {
		return err
	}

	// Collect user input
	values, err := g.promptCollector.CollectValues(cfg.Variables, cfg.GetVariableOrder())
	if err != nil {
		return fmt.Errorf("collect values: %v", err)
	}

	// Execute pre-generation hooks
	if err := g.executeHooks(cfg.Hooks.PreGeneration, "pre-generation", templatePath, values); err != nil {
		return err
	}

	// Generate files
	if err := g.generateFiles(templatePath, opts.OutputDir, values, cfg.Template); err != nil {
		return err
	}

	// Execute post-generation hooks
	if err := g.executeHooks(cfg.Hooks.PostGeneration, "post-generation", opts.OutputDir, values); err != nil {
		return err
	}

	// Success message
	prompts.Outro("🎉 Template generated successfully!", prompts.MessageOptions{Output: g.term.Writer})
	return nil
}

// loadConfig loads and parses the template configuration
func (g *Generator) loadConfig(templatePath string) (config.Config, error) {
	cfgPath := filepath.Join(templatePath, config.CutrYAML)
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		return config.Config{}, fmt.Errorf("read %s: %v", cfgPath, err)
	}

	cfg, err := config.ParseCutrYAML(cfgData)
	if err != nil {
		return config.Config{}, fmt.Errorf("parse config: %v", err)
	}

	return cfg, nil
}

// executeHooks runs pre or post generation hooks with progress display
func (g *Generator) executeHooks(hookCommands []string, hookType, workDir string, data map[string]any) error {
	if len(hookCommands) == 0 {
		return nil
	}

	message := fmt.Sprintf("🔧 Running %s hooks...", hookType)
	return g.progressHandler.ShowProgress(message, func() error {
		hookExecutor := hooks.New()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		if hookType == "pre-generation" {
			return hookExecutor.ExecutePreGeneration(ctx, config.Hooks{PreGeneration: hookCommands}, workDir, data)
		} else {
			return hookExecutor.ExecutePostGeneration(ctx, config.Hooks{PostGeneration: hookCommands}, workDir, data)
		}
	})
}

// generateFiles renders the template tree with progress display
func (g *Generator) generateFiles(templatePath, outputDir string, data map[string]any, settings config.TemplateSettings) error {
	return g.progressHandler.ShowProgress("🎨 Generating files...", func() error {
		rend := renderer.New()
		return rend.RenderTreeWithSettings(templatePath, outputDir, data, settings)
	})
}
