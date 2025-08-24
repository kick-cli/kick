package internal

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"text/template"

	"github.com/yarlson/tap"
)

// Executor handles hook execution operations.
type Executor struct {
	stream *tap.Stream
}

// New creates a new hook executor.
func New() *Executor {
	return &Executor{}
}

// NewWithStream creates a new hook executor with a provided tap stream.
func NewWithStream(stream *tap.Stream) *Executor {
	return &Executor{stream: stream}
}

// ExecutePreGeneration executes pre-generation hooks.
func (e *Executor) ExecutePreGeneration(ctx context.Context, hooks Hooks, workDir string, data map[string]any) error {
	for _, command := range hooks.PreGeneration {
		if err := e.executeCommand(ctx, command, workDir, data); err != nil {
			return fmt.Errorf("execute pre-generation hook: %w", err)
		}
	}
	return nil
}

// ExecutePostGeneration executes post-generation hooks.
func (e *Executor) ExecutePostGeneration(ctx context.Context, hooks Hooks, workDir string, data map[string]any) error {
	for _, command := range hooks.PostGeneration {
		if err := e.executeCommand(ctx, command, workDir, data); err != nil {
			return fmt.Errorf("execute post-generation hook: %w", err)
		}
	}
	return nil
}

// executeCommand executes a single hook command with template rendering.
func (e *Executor) executeCommand(ctx context.Context, command string, workDir string, data map[string]any) error {
	// Render the command template
	renderedCommand, err := e.renderCommand(command, data)
	if err != nil {
		return fmt.Errorf("render hook command: %w", err)
	}

	// Execute the command
	cmd := exec.CommandContext(ctx, "sh", "-c", renderedCommand)
	cmd.Dir = workDir
	cmd.Env = os.Environ()

	if e.stream != nil {
		// Use tap stream's built-in Pipe method for simple streaming
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("create stdout pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("start hook command: %w", err)
		}

		// Let tap stream handle the output streaming
		e.stream.Pipe(stdout)

		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("execute hook command: %w", err)
		}
	} else {
		// Fallback to original implementation
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			if stderr.Len() > 0 {
				return fmt.Errorf("execute hook command: %w (stderr: %s)", err, stderr.String())
			}
			return fmt.Errorf("execute hook command: %w", err)
		}
	}

	return nil
}

// renderCommand renders a command string as a template.
func (e *Executor) renderCommand(command string, data map[string]any) (string, error) {
	tmpl, err := template.New("hook").Option("missingkey=error").Parse(command)
	if err != nil {
		return "", fmt.Errorf("parse command template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute command template: %w", err)
	}

	return buf.String(), nil
}
