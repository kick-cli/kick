package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yarlson/cutr/internal/config"
	"github.com/yarlson/cutr/internal/prompt"
	"github.com/yarlson/cutr/internal/source"
	"github.com/yarlson/cutr/internal/template"
	"github.com/yarlson/cutr/internal/ui"
)

//go:embed _internal/README.txt
var internalFS embed.FS

func main() {
	if len(os.Args) < 2 || hasHelpFlag(os.Args[1:]) {
		usage()
		return
	}
	src := os.Args[1]
	out := "."
	if len(os.Args) >= 3 {
		out = os.Args[2]
	}

	// Resolve template source
	resolver := source.New()
	templatePath, cleanup, err := resolver.Resolve(src)
	if err != nil {
		fatal("resolve template: %v", err)
	}
	if cleanup != nil {
		defer cleanup()
	}

	// Parse configuration
	cfgPath := filepath.Join(templatePath, config.CookiecutterJSON)
	cfgData, err := os.ReadFile(cfgPath)
	if err != nil {
		fatal("read %s: %v", cfgPath, err)
	}

	specs, order, err := config.ParseCookiecutterJSON(cfgData)
	if err != nil {
		fatal("parse config: %v", err)
	}

	// Prompt for values
	values, err := prompt.Values(specs, order)
	if err != nil {
		fatal("prompt: %v", err)
	}

	// Root data for templates roughly mimics cookiecutter: .cookiecutter.<var>
	data := map[string]any{"cookiecutter": values}

	// Render template tree
	renderer := template.New()
	if err := renderer.RenderTree(templatePath, out, data); err != nil {
		fatal("render: %v", err)
	}

	ui.PrintSuccess("✅ Done.")
}

func usage() {
	_, _ = fmt.Fprintf(os.Stdout, `cutr – minimal Cookiecutter-like generator in Go

Usage:
  cutr <template> [output_dir]

<template> can be:
  - local directory path
  - git URL (https/ssh) or something ending in .git (cloned in-process)

Template expects %s at the root with variables.

Example:
  cutr gh://my-org/service-template ./my-service
  cutr /path/to/template ./out

`, config.CookiecutterJSON)
	_ = internalFS // just to ensure embed keeps something; not required for runtime
}

func hasHelpFlag(args []string) bool {
	for _, a := range args {
		switch a {
		case "-h", "--help", "help":
			return true
		}
	}
	return false
}

func fatal(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "error: "+format+"\n", a...)
	os.Exit(1)
}
