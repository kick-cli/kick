package hooks

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yarlson/cutr/internal/config"
)

func TestExecutor_ExecutePreGeneration(t *testing.T) {
	tests := []struct {
		name        string
		hooks       config.Hooks
		workDir     string
		data        map[string]any
		wantErr     bool
		errContains string
		setupFunc   func(t *testing.T, workDir string)
		validateFunc func(t *testing.T, workDir string)
	}{
		{
			name: "empty hooks",
			hooks: config.Hooks{},
			data: map[string]any{"name": "test"},
		},
		{
			name: "single echo command",
			hooks: config.Hooks{
				PreGeneration: []string{"echo 'Starting generation'"},
			},
			data: map[string]any{"name": "test"},
		},
		{
			name: "multiple commands",
			hooks: config.Hooks{
				PreGeneration: []string{
					"echo 'First command'",
					"echo 'Second command'",
				},
			},
			data: map[string]any{"name": "test"},
		},
		{
			name: "template rendering in command",
			hooks: config.Hooks{
				PreGeneration: []string{
					"echo 'Project: {{.name}}'",
				},
			},
			data: map[string]any{"name": "my-project"},
		},
		{
			name: "create file hook",
			hooks: config.Hooks{
				PreGeneration: []string{
					"touch pre-generation-marker.txt",
				},
			},
			data: map[string]any{"name": "test"},
			validateFunc: func(t *testing.T, workDir string) {
				_, err := os.Stat(filepath.Join(workDir, "pre-generation-marker.txt"))
				assert.NoError(t, err, "pre-generation marker file should exist")
			},
		},
		{
			name: "invalid command",
			hooks: config.Hooks{
				PreGeneration: []string{
					"nonexistent-command-xyz",
				},
			},
			data:        map[string]any{"name": "test"},
			wantErr:     true,
			errContains: "execute hook",
		},
		{
			name: "command with exit code 1",
			hooks: config.Hooks{
				PreGeneration: []string{
					"exit 1",
				},
			},
			data:        map[string]any{"name": "test"},
			wantErr:     true,
			errContains: "exit status 1",
		},
		{
			name: "template rendering error",
			hooks: config.Hooks{
				PreGeneration: []string{
					"echo '{{.nonexistent}}'",
				},
			},
			data:        map[string]any{"name": "test"},
			wantErr:     true,
			errContains: "render hook command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary work directory
			workDir, err := os.MkdirTemp("", "cutr-hooks-*")
			require.NoError(t, err)
			defer func() { _ = os.RemoveAll(workDir) }()

			// Setup if needed
			if tt.setupFunc != nil {
				tt.setupFunc(t, workDir)
			}

			executor := New()
			err = executor.ExecutePreGeneration(context.Background(), tt.hooks, workDir, tt.data)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// Validate if needed
			if tt.validateFunc != nil {
				tt.validateFunc(t, workDir)
			}
		})
	}
}

func TestExecutor_ExecutePostGeneration(t *testing.T) {
	tests := []struct {
		name        string
		hooks       config.Hooks
		workDir     string
		data        map[string]any
		wantErr     bool
		errContains string
		setupFunc   func(t *testing.T, workDir string)
		validateFunc func(t *testing.T, workDir string)
	}{
		{
			name: "empty hooks",
			hooks: config.Hooks{},
			data: map[string]any{"name": "test"},
		},
		{
			name: "git init command",
			hooks: config.Hooks{
				PostGeneration: []string{
					"git init",
				},
			},
			data: map[string]any{"name": "test"},
			validateFunc: func(t *testing.T, workDir string) {
				_, err := os.Stat(filepath.Join(workDir, ".git"))
				assert.NoError(t, err, ".git directory should exist")
			},
		},
		{
			name: "multiple post commands",
			hooks: config.Hooks{
				PostGeneration: []string{
					"touch post-marker-1.txt",
					"touch post-marker-2.txt",
				},
			},
			data: map[string]any{"name": "test"},
			validateFunc: func(t *testing.T, workDir string) {
				_, err := os.Stat(filepath.Join(workDir, "post-marker-1.txt"))
				assert.NoError(t, err, "post marker 1 should exist")
				_, err = os.Stat(filepath.Join(workDir, "post-marker-2.txt"))
				assert.NoError(t, err, "post marker 2 should exist")
			},
		},
		{
			name: "template with project name",
			hooks: config.Hooks{
				PostGeneration: []string{
					"echo '{{.project_name}}' > project-name.txt",
				},
			},
			data: map[string]any{"project_name": "awesome-project"},
			validateFunc: func(t *testing.T, workDir string) {
				content, err := os.ReadFile(filepath.Join(workDir, "project-name.txt"))
				require.NoError(t, err)
				assert.Equal(t, "awesome-project\n", string(content))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary work directory
			workDir, err := os.MkdirTemp("", "cutr-hooks-*")
			require.NoError(t, err)
			defer func() { _ = os.RemoveAll(workDir) }()

			// Setup if needed
			if tt.setupFunc != nil {
				tt.setupFunc(t, workDir)
			}

			executor := New()
			err = executor.ExecutePostGeneration(context.Background(), tt.hooks, workDir, tt.data)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			// Validate if needed
			if tt.validateFunc != nil {
				tt.validateFunc(t, workDir)
			}
		})
	}
}

func TestExecutor_ExecuteWithTimeout(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		timeout     time.Duration
		wantErr     bool
		errContains string
	}{
		{
			name:    "quick command within timeout",
			command: "echo 'quick'",
			timeout: 5 * time.Second,
		},
		{
			name:        "slow command exceeds timeout",
			command:     "sleep 2",
			timeout:     100 * time.Millisecond,
			wantErr:     true,
			errContains: "signal: killed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir, err := os.MkdirTemp("", "cutr-timeout-*")
			require.NoError(t, err)
			defer func() { _ = os.RemoveAll(workDir) }()

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			executor := New()
			err = executor.executeCommand(ctx, tt.command, workDir, map[string]any{})

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestExecutor_ExecuteBothHooks(t *testing.T) {
	t.Run("complete hook workflow", func(t *testing.T) {
		workDir, err := os.MkdirTemp("", "cutr-workflow-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workDir) }()

		hooks := config.Hooks{
			PreGeneration: []string{
				"echo 'pre: {{.name}}' > pre.txt",
				"mkdir -p src",
			},
			PostGeneration: []string{
				"echo 'post: {{.name}}' > post.txt",
				"touch src/main.go",
			},
		}

		data := map[string]any{"name": "test-project"}

		executor := New()

		// Execute pre-generation hooks
		err = executor.ExecutePreGeneration(context.Background(), hooks, workDir, data)
		require.NoError(t, err)

		// Verify pre-generation results
		preContent, err := os.ReadFile(filepath.Join(workDir, "pre.txt"))
		require.NoError(t, err)
		assert.Equal(t, "pre: test-project\n", string(preContent))

		_, err = os.Stat(filepath.Join(workDir, "src"))
		require.NoError(t, err)

		// Execute post-generation hooks
		err = executor.ExecutePostGeneration(context.Background(), hooks, workDir, data)
		require.NoError(t, err)

		// Verify post-generation results
		postContent, err := os.ReadFile(filepath.Join(workDir, "post.txt"))
		require.NoError(t, err)
		assert.Equal(t, "post: test-project\n", string(postContent))

		_, err = os.Stat(filepath.Join(workDir, "src", "main.go"))
		require.NoError(t, err)
	})
}