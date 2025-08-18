#!/bin/bash

# Build script for {{.project_name}}
# This script builds the application with version information

set -e

# Get version information
VERSION=${VERSION:-"dev"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +"%Y-%m-%d %H:%M:%S UTC")

# Build flags
LDFLAGS="-X '{{.module_name}}/cmd.Version=${VERSION}' -X '{{.module_name}}/cmd.Commit=${COMMIT}' -X '{{.module_name}}/cmd.Date=${DATE}'"

echo "🔨 Building {{.project_name}}..."
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT}"
echo "Date: ${DATE}"
echo ""

# Build the application
go build -ldflags "${LDFLAGS}" -o {{.project_name}} .

echo "✅ Build complete! Binary: ./{{.project_name}}"
echo ""
echo "Usage:"
echo "  ./{{.project_name}} --help"
{{- if .include_version}}
echo "  ./{{.project_name}} version"
{{- end}}
echo "  ./{{.project_name}} hello"