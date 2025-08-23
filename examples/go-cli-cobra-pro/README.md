# {{.project_name | title}}

{{.description}}

## Features

- ⚡️ Cobra-based CLI with ergonomic DX
- 🧪 Tests using Testify
- 🧹 Linting with golangci-lint
- 🧰 Makefile for common tasks
- 🐳 Reproducible builds with GoReleaser config
- 📝 Conventional commits and CHANGELOG support
- 🔧 Structured logging with zerolog
{{- if .include_config}}
- ⚙️ Configuration with Viper (file + env)
{{- end}}
{{- if .include_version}}
- 📦 Version command with build-time info
{{- end}}
{{- if .include_ci}}
- 🚦 GitHub Actions CI (test + lint + build)
{{- end}}

## Quick Start

```bash
# Generate project
cutr gh://yarlson/cutr/examples/go-cli-cobra-pro ./{{.project_name}}
cd {{.project_name}}

# Run
go run . --help

# Lint
make lint

# Test
make test

# Build binary
make build
```

## Project Structure

```
{{.project_name}}/
├── cmd/
│   ├── root.go
│   ├── hello.go
{{- if .include_version}}
│   └── version.go
{{- end}}
├── internal/
│   ├── greet/
│   │   └── greet.go
│   └── version/
│       └── version.go
├── scripts/
│   └── build.sh
├── .golangci.yml
├── .goreleaser.yaml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

## Makefile targets

```bash
make help     # List targets
make lint     # Run golangci-lint
make test     # Run tests
make build    # Build binary
make release  # Dry-run goreleaser build
```

## License

{{.license}}
