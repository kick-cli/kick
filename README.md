## cutr

A tiny, fast project scaffolder for developers. Point it at a template (local folder or Git repo), answer a few prompts, and it renders files and directory names using Go templates. Inspired by Cookiecutter, with a minimal surface area and zero config flags.

## Highlights

- **Simple sources**: local directory, any `https://` or `ssh://` Git URL, anything ending with `.git`, or `gh://owner/repo` shorthand
- **Interactive prompts**: string, choice, number, and boolean with defaults, choices, and basic validation hints
- **Path templating**: both file contents and directory/file names can contain `{{...}}`
- **Safe rendering**: binary files are copied as-is; file permissions are preserved
- **Strict by default**: missing template variables fail the run (no silent fallbacks)
- **No noise**: skips copying `.git` and the template config `cutr.yaml`

## Install

- **With Go (recommended)**

```bash
go install github.com/yarlson/cutr@latest
```

- **From source**

```bash
git clone https://github.com/yarlson/cutr
cd cutr
go build -o cutr
# optional: move onto PATH
mv ./cutr /usr/local/bin/
```

- **Run without installing**

```bash
go run . <template> [output_dir]
```

Requires Go 1.24 or newer.

## Usage

```bash
cutr <template> [output_dir]
```

- **template**: local path, Git URL, `.git` URL, or `gh://owner/repo`
- **output_dir**: where to render the project (defaults to current directory `.`)

Examples:

```bash
# Local template folder → render into ./my-app
cutr ./path/to/template ./my-app

# Public Git repo (HTTPS)
cutr https://github.com/your-org/service-template.git ./svc

# SSH URL
cutr git@github.com:your-org/service-template.git ./svc

# GitHub shorthand
cutr gh://your-org/service-template ./svc
```

## Template layout

Your template is a normal folder whose root contains a `cutr.yaml` config. Everything except `.git` and `cutr.yaml` is processed.

- Files are treated as text and rendered with Go `text/template`
- Binary files are detected and copied as-is
- Directory and file names can also be templates; empty results are skipped
- Original file permissions are preserved

Example structure (template side):

```
my-template/
  cutr.yaml
  {{.project_name}}/
    README.md
    cmd/{{.project_name}}/main.go
    Makefile
```

## cutr.yaml reference

Minimal example:

```yaml
name: "go-service"
description: "Production-ready Go service template"
version: "1.0.0"

variables:
  project_name:
    type: string
    prompt: "Project name"
    default: "my-service"

  database:
    type: choice
    prompt: "Database"
    choices: [postgres, mysql]
    default: postgres

  enable_auth:
    type: boolean
    prompt: "Enable authentication?"
    default: false

  port:
    type: number
    prompt: "HTTP port"
    default: 8080
    min: 1024
    max: 65535
```

Supported variable types:

- **string**: optional `prompt`, `default`, `pattern`, `help`
- **choice**: `choices` (required), optional `prompt`, `default`, `help`
- **number**: optional `prompt`, `default`, `min`, `max`
- **boolean**: optional `prompt`, `default`

Notes:

- Prompts are asked in **alphabetical order** of variable names
- Rendering uses `missingkey=error`: referencing an undefined variable fails the run

Optional keys that are parsed but not executed yet (safe to include for future compatibility):

```yaml
hooks:
  pre_generation:
    - "echo 'Starting'"
  post_generation:
    - "echo 'Done'"

template:
  ignore_patterns: ["*.tmp", ".DS_Store"]
  keep_permissions: true
```

- **hooks** and **template.ignore_patterns** are recognized by the config parser but currently not run/enforced by the renderer
- **keep_permissions** matches current behavior (permissions are preserved)

## Built-in template functions

You can use these in file contents and path segments:

- **upper(s)**, **lower(s)**, **title(s)**
- **trim(s)**
- **snake(s)** → `my_project_name`
- **kebab(s)** → `my-project-name`
- **camel(s)** → `myProjectName`
- **pascal(s)** → `MyProjectName`
- **replace(s, old, new)**

Example:

```gotemplate
Service: {{.project_name | title}}
package {{.project_name | snake}}
```

## Prompts and non-TTY environments

- Interactive prompts use a modern TUI
- If a TTY is not available, cutr falls back to simple stdin prompts (press Enter to accept defaults). For booleans, common inputs like "y/yes/true/1" are accepted.

## Common errors

- "map has no entry for key": your template references a variable that wasn’t provided; add it to `cutr.yaml` or adjust the template
- Template parse errors (e.g., unmatched `{{`): fix the syntax in the template file
- "template path must be a directory": the source path must be a folder (not a single file)
- Git clone errors: ensure the repo exists and you have access (private repos require auth)

## End-to-end example

Given the `cutr.yaml` above, you can use variables in names and contents:

```
cmd/{{.project_name}}/main.go
```

```gotemplate
package main

import "fmt"

func main() {
  fmt.Println("{{.project_name | title}} is alive on port {{.port}}!")
}
```

Running:

```bash
cutr gh://your-org/go-service-template ./awesome
```

Produces:

```
./awesome/
  cmd/awesome/main.go
  README.md
  ...
```

## Development

- **Run tests**: `go test ./...`
- **Build**: `go build`
- **Run locally**: `go run . <template> [output_dir]`

## Why cutr?

- Minimal moving parts; built for speed and clarity
- Friendly defaults and helpful failures
- Uses familiar Go templates and adds practical string helpers

Related: Cookiecutter popularized this workflow; cutr keeps the spirit but pares it down for Go-first teams.
