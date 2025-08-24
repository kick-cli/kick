# kick Examples

This directory contains example templates that demonstrate various features of kick.

## Available Templates

### go-cli-cobra

A comprehensive Go CLI application template using the Cobra framework.

**Features:**
- Modern Go CLI structure with Cobra
- Optional configuration file support with Viper
- Optional version command with build-time information
- Example subcommands and flags
- Build scripts and Makefile
- License templates (MIT, Apache-2.0, GPL-3.0, BSD-3-Clause)
- Comprehensive hooks for project setup
- Pattern validation for inputs
- Clean project structure

**Usage:**

```bash
# Use the template to create a new project
kick examples/go-cli-cobra ./my-awesome-cli

# Or from a remote repository
kick gh://kick-cli/kick/examples/go-cli-cobra ./my-awesome-cli
```

**What you'll get:**
- A complete Go CLI project with Cobra
- Interactive prompts for project configuration
- Automatic Go module initialization
- Dependency installation
- Executable build script
- Ready-to-use project structure

**Template Variables:**
- `project_name` - Name of the CLI application (lowercase, hyphens allowed)
- `module_name` - Go module path (e.g., github.com/user/project)
- `author_name` - Your name
- `author_email` - Your email address
- `description` - Project description
- `license` - License type (MIT, Apache-2.0, GPL-3.0, BSD-3-Clause)
- `include_config` - Include configuration file support (boolean)
- `include_version` - Include version command (boolean)
- `go_version` - Go version requirement

**Hooks Demonstrated:**
- Pre-generation: Project information display
- Post-generation: Go module init, dependency installation, build setup

## Using the Examples

1. **Local usage:**
   ```bash
   kick examples/go-cli-cobra ./my-project
   ```

2. **Remote usage (when examples are in a Git repository):**
   ```bash
   kick gh://kick-cli/kick/examples/go-cli-cobra ./my-project
   ```

3. **Test the template without committing:**
   ```bash
   kick examples/go-cli-cobra /tmp/test-project
   cd /tmp/test-project
   go run . --help
   ```

## Creating Your Own Templates

Use these examples as inspiration for creating your own templates:

1. Study the `kick.yaml` configuration
2. Look at how template variables are used throughout the files
3. See how hooks automate setup tasks
4. Notice how conditional content works with `{{- if .variable}}`

## Contributing Examples

We welcome additional examples! Please ensure your template:

1. Has a complete `kick.yaml` with good variable validation
2. Demonstrates useful kick features (hooks, template settings, etc.)
3. Includes a README explaining what the template does
4. Works end-to-end without manual intervention
5. Follows Go and general best practices

Submit a pull request with your template in a new subdirectory.