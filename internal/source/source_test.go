package source

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	resolver := New()

	require.NotNil(t, resolver)
	assert.IsType(t, &Resolver{}, resolver)
}

func TestResolver_Resolve(t *testing.T) {
	// Create test directories for local path testing
	testDir, err := os.MkdirTemp("", "cutr-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(testDir) }()

	testFile := filepath.Join(testDir, "testfile.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	nonExistentDir := filepath.Join(testDir, "nonexistent")

	tests := []struct {
		name           string
		src            string
		wantPath       string
		wantCleanup    bool
		wantErr        bool
		errContains    string
		skipReason     string
		validateResult func(t *testing.T, path string, cleanup func())
	}{
		// Local path tests
		{
			name:        "valid local directory",
			src:         testDir,
			wantPath:    testDir,
			wantCleanup: false,
			validateResult: func(t *testing.T, path string, cleanup func()) {
				assert.Equal(t, testDir, path)
				assert.Nil(t, cleanup)
			},
		},
		{
			name:        "local file instead of directory",
			src:         testFile,
			wantErr:     true,
			errContains: "template path must be a directory",
		},
		{
			name:        "non-existent local path",
			src:         nonExistentDir,
			wantErr:     true,
			errContains: "no such file or directory",
		},
		{
			name:        "empty path",
			src:         "",
			wantErr:     true,
			errContains: "no such file or directory",
		},

		// Git URL tests (these will likely fail in test environment)
		{
			name:       "github https URL",
			src:        "https://github.com/octocat/Hello-World.git",
			wantErr:    true, // Expected to fail in test environment
			skipReason: "requires network access and git",
		},
		{
			name:       "github ssh URL",
			src:        "git@github.com:octocat/Hello-World.git",
			wantErr:    true, // Expected to fail in test environment
			skipReason: "requires network access and git",
		},
		{
			name:       "gh:// shorthand URL",
			src:        "gh://octocat/Hello-World",
			wantErr:    true, // Expected to fail in test environment
			skipReason: "requires network access and git",
		},
		{
			name:       "generic .git URL",
			src:        "https://example.com/repo.git",
			wantErr:    true, // Expected to fail in test environment
			skipReason: "requires network access and git",
		},
		{
			name:       "ssh:// git URL",
			src:        "ssh://git@github.com/octocat/Hello-World.git",
			wantErr:    true, // Expected to fail in test environment
			skipReason: "requires network access and git",
		},

		// Edge cases for git-like detection
		{
			name:       "local path ending with .git",
			src:        filepath.Join(testDir, "repo.git"),
			wantErr:    true,
			skipReason: "git clone will fail for non-existent local .git path",
		},
		{
			name:       "local path with git@ in name",
			src:        filepath.Join(testDir, "git@test"),
			wantErr:    true,
			skipReason: "git@ pattern triggers git clone behavior",
		},
	}

	resolver := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipReason != "" {
				t.Skip("Skipping:", tt.skipReason)
				return
			}

			path, cleanup, err := resolver.Resolve(tt.src)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				// For error cases, ensure cleanup is provided if path was created
				if cleanup != nil {
					defer cleanup()
				}
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, path)

			// If cleanup function is provided, defer it and validate it works
			if cleanup != nil {
				defer func() {
					// Verify path exists before cleanup
					_, err := os.Stat(path)
					assert.NoError(t, err, "Path should exist before cleanup")

					// Call cleanup
					cleanup()

					// Verify path is cleaned up (for temp directories)
					if strings.Contains(path, "cutr-") {
						_, err = os.Stat(path)
						assert.Error(t, err, "Temporary path should be cleaned up")
					}
				}()
			}

			// Run custom validation if provided
			if tt.validateResult != nil {
				tt.validateResult(t, path, cleanup)
			}

			// Basic validation for all successful cases
			if tt.wantPath != "" {
				assert.Equal(t, tt.wantPath, path)
			}

			if tt.wantCleanup {
				assert.NotNil(t, cleanup, "Expected cleanup function")
			} else {
				assert.Nil(t, cleanup, "Expected no cleanup function for local paths")
			}
		})
	}
}

func TestResolver_Resolve_GitIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping git integration test in short mode")
	}

	resolver := New()

	t.Run("public github repository", func(t *testing.T) {
		// Try to clone a small public repository
		src := "https://github.com/octocat/Hello-World.git"

		path, cleanup, err := resolver.Resolve(src)

		if err != nil {
			// If it fails, check if it's due to network/git issues
			errStr := err.Error()
			if strings.Contains(errStr, "repository not found") ||
				strings.Contains(errStr, "network") ||
				strings.Contains(errStr, "timeout") ||
				strings.Contains(errStr, "connection") {
				t.Skip("Skipping due to network/git issues:", err)
				return
			}
			t.Fatalf("Unexpected error: %v", err)
		}

		require.NoError(t, err)
		require.NotEmpty(t, path)
		require.NotNil(t, cleanup)

		// Verify the cloned directory exists and has git content
		info, err := os.Stat(path)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		// Check for typical git clone contents
		entries, err := os.ReadDir(path)
		require.NoError(t, err)
		assert.NotEmpty(t, entries, "Cloned directory should not be empty")

		// Cleanup
		cleanup()

		// Verify cleanup worked
		_, err = os.Stat(path)
		assert.Error(t, err, "Directory should be cleaned up")
	})
}

func TestResolver_Resolve_ErrorScenarios(t *testing.T) {
	resolver := New()

	t.Run("nil resolver method call", func(t *testing.T) {
		// Test that the method handles basic cases properly
		// This is more about ensuring the method signature works correctly
		testDir, err := os.MkdirTemp("", "cutr-error-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(testDir) }()

		path, cleanup, err := resolver.Resolve(testDir)
		require.NoError(t, err)
		assert.Equal(t, testDir, path)
		assert.Nil(t, cleanup)
	})
}
