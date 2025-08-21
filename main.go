package main

import (
	"fmt"
	"os"

	"github.com/yarlson/cutr/internal/config"
	"github.com/yarlson/cutr/internal/generator"
)

func main() {
	if len(os.Args) < 2 || hasHelpFlag(os.Args[1:]) {
		usage()
		return
	}

	// Parse command line arguments
	opts := generator.Options{
		Source:    os.Args[1],
		OutputDir: ".",
	}
	if len(os.Args) >= 3 {
		opts.OutputDir = os.Args[2]
	}

	// Create generator and run
	gen, err := generator.New()
	if err != nil {
		fatal("initialize generator: %v", err)
	}
	defer gen.Close()

	if err := gen.Generate(opts); err != nil {
		fatal("generate template: %v", err)
	}
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

`, config.CutrYAML)
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
