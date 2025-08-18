# {{.project_name | title}}

{{.description}}

## Description

This is a modern Go CLI application built with the [Cobra](https://github.com/spf13/cobra) framework. It provides a clean, extensible structure for building command-line tools.

## Features

- 🐍 Built with Cobra CLI framework
- 📁 Clean project structure
{{- if .include_config}}
- ⚙️ Configuration file support with Viper
{{- end}}
{{- if .include_version}}
- 📦 Version command with build information
{{- end}}
- 🚀 Ready for development and deployment

## Installation

### From source

```bash
git clone <your-repo-url>
cd {{.project_name}}
go build -o {{.project_name}}
```

### Using go install

```bash
go install {{.module_name}}@latest
```

## Usage

```bash
# Show help
./{{.project_name}} --help

# Run the hello command
./{{.project_name}} hello
./{{.project_name}} hello "Your Name"
./{{.project_name}} hello --uppercase "Your Name"

{{- if .include_version}}
# Show version
./{{.project_name}} version
{{- end}}
```

## Development

### Building

Use the provided build script:

```bash
./scripts/build.sh
```

Or build manually:

```bash
go build -o {{.project_name}} .
```

### Adding new commands

1. Create a new file in the `cmd/` directory
2. Follow the pattern from `cmd/hello.go`
3. Add your command logic

Example:

```bash
cobra-cli add mycommand
```

## Project Structure

```
{{.project_name}}/
├── main.go              # Application entry point
├── cmd/                 # Command definitions
│   ├── root.go          # Root command and CLI setup
{{- if .include_version}}
│   ├── version.go       # Version command
{{- end}}
│   └── hello.go         # Example hello command
├── scripts/             # Build and utility scripts
│   └── build.sh         # Build script
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
└── README.md            # This file
```

{{- if .include_config}}

## Configuration

The application supports configuration files in YAML format. By default, it looks for:

- `$HOME/.{{.project_name}}.yaml`
- `./.{{.project_name}}.yaml`

You can also specify a custom config file:

```bash
./{{.project_name}} --config /path/to/config.yaml
```

Example configuration:

```yaml
# ~/.{{.project_name}}.yaml
verbose: true
output: json
```
{{- end}}

## License

This project is licensed under the {{.license}} License - see the [LICENSE](LICENSE) file for details.

## Author

{{.author_name}} <{{.author_email}}>

---

*Generated with [cutr](https://github.com/yarlson/cutr)*