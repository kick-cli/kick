package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	renderer := New()

	require.NotNil(t, renderer)
	assert.IsType(t, &Renderer{}, renderer)
}

func TestRenderer_RenderTree(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) (srcRoot, outRoot string, data map[string]any)
		cleanupFunc  func(t *testing.T, srcRoot, outRoot string)
		wantErr      bool
		errContains  string
		validateFunc func(t *testing.T, outRoot string)
	}{
		// Basic functionality tests
		{
			name: "empty source directory",
			setupFunc: func(t *testing.T) (string, string, map[string]any) {
				srcRoot, err := os.MkdirTemp("", "cutr-src-*")
				require.NoError(t, err)
				outRoot, err := os.MkdirTemp("", "cutr-out-*")
				require.NoError(t, err)

				data := map[string]any{
					"project_name": "test-project",
				}
				return srcRoot, outRoot, data
			},
			validateFunc: func(t *testing.T, outRoot string) {
				// Should create output directory even if source is empty
				info, err := os.Stat(outRoot)
				require.NoError(t, err)
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "simple file rendering",
			setupFunc: func(t *testing.T) (string, string, map[string]any) {
				srcRoot, err := os.MkdirTemp("", "cutr-src-*")
				require.NoError(t, err)
				outRoot, err := os.MkdirTemp("", "cutr-out-*")
				require.NoError(t, err)

				// Create a simple template file
				content := "Project: {{.project_name}}\nAuthor: {{.author}}"
				err = os.WriteFile(filepath.Join(srcRoot, "README.md"), []byte(content), 0644)
				require.NoError(t, err)

				data := map[string]any{
					"project_name": "MyProject",
					"author":       "John Doe",
				}
				return srcRoot, outRoot, data
			},
			validateFunc: func(t *testing.T, outRoot string) {
				content, err := os.ReadFile(filepath.Join(outRoot, "README.md"))
				require.NoError(t, err)
				expected := "Project: MyProject\nAuthor: John Doe"
				assert.Equal(t, expected, string(content))
			},
		},
		{
			name: "directory structure rendering",
			setupFunc: func(t *testing.T) (string, string, map[string]any) {
				srcRoot, err := os.MkdirTemp("", "cutr-src-*")
				require.NoError(t, err)
				outRoot, err := os.MkdirTemp("", "cutr-out-*")
				require.NoError(t, err)

				// Create nested directory structure
				subDir := filepath.Join(srcRoot, "{{.project_name}}")
				err = os.MkdirAll(subDir, 0755)
				require.NoError(t, err)

				// Create file in subdirectory
				content := "Package: {{.project_name}}"
				err = os.WriteFile(filepath.Join(subDir, "main.go"), []byte(content), 0644)
				require.NoError(t, err)

				data := map[string]any{
					"project_name": "awesome-app",
				}
				return srcRoot, outRoot, data
			},
			validateFunc: func(t *testing.T, outRoot string) {
				// Check rendered directory name
				expectedDir := filepath.Join(outRoot, "awesome-app")
				info, err := os.Stat(expectedDir)
				require.NoError(t, err)
				assert.True(t, info.IsDir())

				// Check rendered file content
				content, err := os.ReadFile(filepath.Join(expectedDir, "main.go"))
				require.NoError(t, err)
				assert.Equal(t, "Package: awesome-app", string(content))
			},
		},
		{
			name: "binary file copying",
			setupFunc: func(t *testing.T) (string, string, map[string]any) {
				srcRoot, err := os.MkdirTemp("", "cutr-src-*")
				require.NoError(t, err)
				outRoot, err := os.MkdirTemp("", "cutr-out-*")
				require.NoError(t, err)

				// Create a binary-like file (with null bytes)
				binaryContent := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00} // PNG header-ish
				err = os.WriteFile(filepath.Join(srcRoot, "image.png"), binaryContent, 0644)
				require.NoError(t, err)

				data := map[string]any{
					"project_name": "test",
				}
				return srcRoot, outRoot, data
			},
			validateFunc: func(t *testing.T, outRoot string) {
				// Binary file should be copied as-is
				content, err := os.ReadFile(filepath.Join(outRoot, "image.png"))
				require.NoError(t, err)
				expected := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
				assert.Equal(t, expected, content)
			},
		},
		{
			name: "template functions usage",
			setupFunc: func(t *testing.T) (string, string, map[string]any) {
				srcRoot, err := os.MkdirTemp("", "cutr-src-*")
				require.NoError(t, err)
				outRoot, err := os.MkdirTemp("", "cutr-out-*")
				require.NoError(t, err)

				// Create template using various functions
				content := `Snake: {{.name | snake}}
Kebab: {{.name | kebab}}
Camel: {{.name | camel}}
Pascal: {{.name | pascal}}
Trim: "{{.padded | trim}}"`
				err = os.WriteFile(filepath.Join(srcRoot, "functions.txt"), []byte(content), 0644)
				require.NoError(t, err)

				data := map[string]any{
					"name":   "My Project Name",
					"padded": "  spaced text  ",
				}
				return srcRoot, outRoot, data
			},
			validateFunc: func(t *testing.T, outRoot string) {
				content, err := os.ReadFile(filepath.Join(outRoot, "functions.txt"))
				require.NoError(t, err)

				expected := `Snake: my_project_name
Kebab: my-project-name
Camel: myProjectName
Pascal: MyProjectName
Trim: "spaced text"`
				assert.Equal(t, expected, string(content))
			},
		},
		{
			name: "skip git and cutr files",
			setupFunc: func(t *testing.T) (string, string, map[string]any) {
				srcRoot, err := os.MkdirTemp("", "cutr-src-*")
				require.NoError(t, err)
				outRoot, err := os.MkdirTemp("", "cutr-out-*")
				require.NoError(t, err)

				// Create .git directory and cutr.yaml
				gitDir := filepath.Join(srcRoot, ".git")
				err = os.MkdirAll(gitDir, 0755)
				require.NoError(t, err)
				err = os.WriteFile(filepath.Join(gitDir, "config"), []byte("git config"), 0644)
				require.NoError(t, err)

				err = os.WriteFile(filepath.Join(srcRoot, "cutr.yaml"), []byte(`name: "test"`), 0644)
				require.NoError(t, err)

				// Create normal file
				err = os.WriteFile(filepath.Join(srcRoot, "normal.txt"), []byte("normal content"), 0644)
				require.NoError(t, err)

				data := map[string]any{
					"name": "test",
				}
				return srcRoot, outRoot, data
			},
			validateFunc: func(t *testing.T, outRoot string) {
				// .git directory should not exist
				_, err := os.Stat(filepath.Join(outRoot, ".git"))
				assert.Error(t, err)

				// cutr.yaml should not exist
				_, err = os.Stat(filepath.Join(outRoot, "cutr.yaml"))
				assert.Error(t, err)

				// Normal file should exist
				content, err := os.ReadFile(filepath.Join(outRoot, "normal.txt"))
				require.NoError(t, err)
				assert.Equal(t, "normal content", string(content))
			},
		},

		// Error cases
		{
			name: "non-existent source directory",
			setupFunc: func(t *testing.T) (string, string, map[string]any) {
				outRoot, err := os.MkdirTemp("", "cutr-out-*")
				require.NoError(t, err)

				data := map[string]any{
					"name": "test",
				}
				return "/non/existent/path", outRoot, data
			},
			wantErr:     true,
			errContains: "no such file or directory",
		},
		{
			name: "template parsing error",
			setupFunc: func(t *testing.T) (string, string, map[string]any) {
				srcRoot, err := os.MkdirTemp("", "cutr-src-*")
				require.NoError(t, err)
				outRoot, err := os.MkdirTemp("", "cutr-out-*")
				require.NoError(t, err)

				// Create file with invalid template syntax
				content := "{{.name" // Missing closing braces
				err = os.WriteFile(filepath.Join(srcRoot, "invalid.txt"), []byte(content), 0644)
				require.NoError(t, err)

				data := map[string]any{
					"name": "test",
				}
				return srcRoot, outRoot, data
			},
			wantErr:     true,
			errContains: "template:",
		},
		{
			name: "missing template variable",
			setupFunc: func(t *testing.T) (string, string, map[string]any) {
				srcRoot, err := os.MkdirTemp("", "cutr-src-*")
				require.NoError(t, err)
				outRoot, err := os.MkdirTemp("", "cutr-out-*")
				require.NoError(t, err)

				// Create file referencing non-existent variable
				content := "Name: {{.nonexistent}}"
				err = os.WriteFile(filepath.Join(srcRoot, "missing.txt"), []byte(content), 0644)
				require.NoError(t, err)

				data := map[string]any{
					"name": "test",
				}
				return srcRoot, outRoot, data
			},
			wantErr:     true,
			errContains: "map has no entry for key",
		},
		{
			name: "invalid output path",
			setupFunc: func(t *testing.T) (string, string, map[string]any) {
				srcRoot, err := os.MkdirTemp("", "cutr-src-*")
				require.NoError(t, err)

				// Create a simple file
				err = os.WriteFile(filepath.Join(srcRoot, "test.txt"), []byte("test"), 0644)
				require.NoError(t, err)

				data := map[string]any{
					"name": "test",
				}
				return srcRoot, "/dev/null/invalid", data // Invalid output path
			},
			wantErr:     true,
			errContains: "not a directory",
		},
	}

	renderer := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcRoot, outRoot, data := tt.setupFunc(t)

			if tt.cleanupFunc != nil {
				defer tt.cleanupFunc(t, srcRoot, outRoot)
			} else {
				// Default cleanup
				defer func() {
					if srcRoot != "/non/existent/path" && srcRoot != "/dev/null/invalid" {
						_ = os.RemoveAll(srcRoot)
					}
					if outRoot != "/dev/null/invalid" {
						_ = os.RemoveAll(outRoot)
					}
				}()
			}

			err := renderer.RenderTree(srcRoot, outRoot, data)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)

			if tt.validateFunc != nil {
				tt.validateFunc(t, outRoot)
			}
		})
	}
}

func TestRenderer_RenderTree_EdgeCases(t *testing.T) {
	renderer := New()

	t.Run("empty template path rendering", func(t *testing.T) {
		srcRoot, err := os.MkdirTemp("", "cutr-src-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(srcRoot) }()

		outRoot, err := os.MkdirTemp("", "cutr-out-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(outRoot) }()

		// Create directory that renders to empty (should be skipped)
		emptyDir := filepath.Join(srcRoot, "{{.empty}}")
		err = os.MkdirAll(emptyDir, 0755)
		require.NoError(t, err)

		data := map[string]any{
			"empty": "", // Will render directory name to empty
		}

		err = renderer.RenderTree(srcRoot, outRoot, data)
		require.NoError(t, err)

		// Check that output directory exists but is empty (no rendered subdirectory)
		entries, err := os.ReadDir(outRoot)
		require.NoError(t, err)
		assert.Empty(t, entries, "Empty-rendered directories should be skipped")
	})

	t.Run("file permissions preservation", func(t *testing.T) {
		srcRoot, err := os.MkdirTemp("", "cutr-src-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(srcRoot) }()

		outRoot, err := os.MkdirTemp("", "cutr-out-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(outRoot) }()

		// Create executable file
		content := "#!/bin/bash\necho 'Hello {{.name}}'"
		srcFile := filepath.Join(srcRoot, "script.sh")
		err = os.WriteFile(srcFile, []byte(content), 0755)
		require.NoError(t, err)

		data := map[string]any{
			"name": "World",
		}

		err = renderer.RenderTree(srcRoot, outRoot, data)
		require.NoError(t, err)

		// Check that permissions are preserved
		outFile := filepath.Join(outRoot, "script.sh")
		info, err := os.Stat(outFile)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0755), info.Mode().Perm())

		// Check that content was rendered
		renderedContent, err := os.ReadFile(outFile)
		require.NoError(t, err)
		expected := "#!/bin/bash\necho 'Hello World'"
		assert.Equal(t, expected, string(renderedContent))
	})
}
