package source

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

// Resolver handles template source resolution (local paths or git repositories)
type Resolver struct{}

// New creates a new source resolver
func New() *Resolver {
	return &Resolver{}
}

// Resolve resolves a template source and returns the local path and optional cleanup function
func (r *Resolver) Resolve(src string) (string, func(), error) {
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
