package internal

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yarlson/tap"
)

func TestExecutor_ExecutePreGeneration(t *testing.T) {
	tests := []struct {
		name         string
		hooks        Hooks
		workDir      string
		data         map[string]any
		wantErr      bool
		errContains  string
		setupFunc    func(t *testing.T, workDir string)
		validateFunc func(t *testing.T, workDir string)
	}{
		{
			name:  "empty hooks",
			hooks: Hooks{},
			data:  map[string]any{"name": "test"},
		},
		{
			name: "single echo command",
			hooks: Hooks{
				PreGeneration: []string{"echo 'Starting generation'"},
			},
			data: map[string]any{"name": "test"},
		},
		{
			name: "multiple commands",
			hooks: Hooks{
				PreGeneration: []string{
					"echo 'First command'",
					"echo 'Second command'",
				},
			},
			data: map[string]any{"name": "test"},
		},
		{
			name: "template rendering in command",
			hooks: Hooks{
				PreGeneration: []string{
					"echo 'Project: {{.name}}'",
				},
			},
			data: map[string]any{"name": "my-project"},
		},
		{
			name: "create file hook",
			hooks: Hooks{
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
			hooks: Hooks{
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
			hooks: Hooks{
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
			hooks: Hooks{
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
			workDir, err := os.MkdirTemp("", "kick-hooks-*")
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
		name         string
		hooks        Hooks
		workDir      string
		data         map[string]any
		wantErr      bool
		errContains  string
		setupFunc    func(t *testing.T, workDir string)
		validateFunc func(t *testing.T, workDir string)
	}{
		{
			name:  "empty hooks",
			hooks: Hooks{},
			data:  map[string]any{"name": "test"},
		},
		{
			name: "git init command",
			hooks: Hooks{
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
			hooks: Hooks{
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
			hooks: Hooks{
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
			workDir, err := os.MkdirTemp("", "kick-hooks-*")
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
			workDir, err := os.MkdirTemp("", "kick-timeout-*")
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
		workDir, err := os.MkdirTemp("", "kick-workflow-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workDir) }()

		hooks := Hooks{
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

func TestExecutor_WithStreamingOutput(t *testing.T) {
	t.Run("executor with stream executes hooks successfully", func(t *testing.T) {
		workDir, err := os.MkdirTemp("", "kick-stream-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workDir) }()

		// Create stream and executor with streaming
		stream := tap.NewStream(tap.StreamOptions{ShowTimer: false})
		executor := NewWithStream(stream)

		hooks := Hooks{
			PreGeneration: []string{
				"echo 'Line 1'",
				"echo 'Line 2 from command'",
				"touch streaming-test.txt",
			},
		}

		data := map[string]any{"name": "test-project"}

		// Start stream (this would normally show in terminal)
		stream.Start("Testing streaming hooks")

		// Execute hooks with streaming
		err = executor.ExecutePreGeneration(context.Background(), hooks, workDir, data)
		require.NoError(t, err, "Hooks should execute successfully with stream")

		// Stop stream
		stream.Stop("Hooks completed", 0)

		// Verify that the commands actually executed (file created)
		_, err = os.Stat(filepath.Join(workDir, "streaming-test.txt"))
		require.NoError(t, err, "Hook commands should have been executed")
	})

	t.Run("executor with stream handles commands with mixed output", func(t *testing.T) {
		workDir, err := os.MkdirTemp("", "kick-stream-mixed-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workDir) }()

		// Create stream and executor with streaming
		stream := tap.NewStream(tap.StreamOptions{ShowTimer: false})
		executor := NewWithStream(stream)

		hooks := Hooks{
			PreGeneration: []string{
				"echo 'stdout message'",
				"echo 'another line'",
				"touch mixed-test.txt",
			},
		}

		data := map[string]any{"name": "test-project"}

		// Start stream
		stream.Start("Testing mixed output streaming")

		// Execute hooks with streaming
		err = executor.ExecutePreGeneration(context.Background(), hooks, workDir, data)
		require.NoError(t, err, "Should handle mixed output correctly")

		// Stop stream
		stream.Stop("Mixed output streaming completed", 0)

		// Verify the file was created (proving commands executed)
		_, err = os.Stat(filepath.Join(workDir, "mixed-test.txt"))
		require.NoError(t, err, "Hook commands should execute successfully")
	})

	t.Run("executor without stream falls back to original behavior", func(t *testing.T) {
		workDir, err := os.MkdirTemp("", "kick-no-stream-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workDir) }()

		// Use executor without stream
		executor := New()

		hooks := Hooks{
			PreGeneration: []string{
				"echo 'This should work without streaming'",
				"touch fallback-test.txt",
			},
		}

		data := map[string]any{"name": "test-project"}

		// Execute hooks without streaming (should fall back to original behavior)
		err = executor.ExecutePreGeneration(context.Background(), hooks, workDir, data)
		require.NoError(t, err, "Should work fine without streaming")

		// Verify file was created with fallback behavior
		_, err = os.Stat(filepath.Join(workDir, "fallback-test.txt"))
		require.NoError(t, err, "Fallback execution should work")
	})

	t.Run("NewWithStream creates executor with stream", func(t *testing.T) {
		stream := tap.NewStream(tap.StreamOptions{ShowTimer: false})
		executor := NewWithStream(stream)

		assert.NotNil(t, executor, "Should create executor")
		assert.Equal(t, stream, executor.stream, "Should store stream reference")
	})

	t.Run("hook commands output only, no command logging", func(t *testing.T) {
		workDir, err := os.MkdirTemp("", "kick-output-only-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(workDir) }()

		// Create stream and executor with streaming
		stream := tap.NewStream(tap.StreamOptions{ShowTimer: false})
		executor := NewWithStream(stream)

		hooks := Hooks{
			PreGeneration: []string{
				"echo 'This is output only'",
				"echo 'No commands should be shown'",
			},
		}

		data := map[string]any{"name": "test-project"}

		// Start stream
		stream.Start("Testing output-only streaming")

		// Execute hooks with streaming
		err = executor.ExecutePreGeneration(context.Background(), hooks, workDir, data)
		require.NoError(t, err, "Should execute successfully")

		// Stop stream
		stream.Stop("Output-only streaming completed", 0)

		// Note: We can't easily test the stream output content in this test setup,
		// but we can verify the functionality works by checking the commands executed.
		// The key point is that only command output should appear in the stream.
	})
}
