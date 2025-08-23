#!/usr/bin/env bash
set -euo pipefail

APP_NAME="{{.project_name}}"

GOFLAGS="-trimpath"
LDFLAGS="-s -w -X {{.module_name}}/internal/version.Version=${VERSION:-dev} -X {{.module_name}}/internal/version.Commit=${COMMIT:-$(git rev-parse --short HEAD || echo local)} -X {{.module_name}}/internal/version.Date=${DATE:-$(date -u +%Y-%m-%d)}"

echo "Building $APP_NAME"
GOFLAGS="$GOFLAGS" go build -ldflags "$LDFLAGS" -o "$APP_NAME" .

echo "Built $APP_NAME"
