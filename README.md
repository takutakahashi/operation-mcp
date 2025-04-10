# Operations CLI Tool

A CLI tool for executing operations defined in a YAML configuration file.

## Features

- Dynamic command generation based on YAML configuration
- Hierarchical command structure with subcommands
- Parameter validation and templating
- Danger level management for sensitive operations
- Configurable action types (confirm, timeout, force)

## Installation

### Prerequisites

- Go 1.24 or later

### Building from source

```bash
# Clone the repository
git clone https://github.com/takutakahashi/operation-mcp.git
cd operation-mcp

# Build the binary
make build

# Install the binary (optional)
make install
```

## Usage

### Configuration

Create a YAML configuration file with your tools and actions. See `docs/examples/config.yaml` for an example.

### Running commands

```bash
# Using the default config file (./config.yaml or ~/.operations/config.yaml)
operations kubectl_get_pod --namespace my-namespace

# Using a specific config file
operations --config /path/to/config.yaml kubectl_get_pod --namespace my-namespace

# Running a subtool with parameters
operations kubectl_describe_pod --namespace my-namespace --pod my-pod

# Running a dangerous operation (will prompt for confirmation)
operations kubectl_delete_pod --namespace my-namespace --pod my-pod
```

## Configuration Format

See `docs/spec.md` for detailed configuration format documentation.

## Development

### Running tests

```bash
make test
```

### Running tests with coverage

```bash
make test-coverage
```

### Formatting code

```bash
make fmt
```

## CI/CD

This project uses GitHub Actions for continuous integration and continuous deployment.

### CI Workflows

- **Unit Tests**: Runs on every pull request.
  - Runs code formatting checks
  - Runs linting
  - Executes unit tests
  - Generates and uploads test coverage report

- **E2E Tests**: Runs on push to main branch and can be manually triggered.
  - Builds the application
  - Runs end-to-end tests using the test configuration
  - Uploads the built binary as an artifact

### CD Workflow

- **Release**: Triggered when a tag with format `v*` is pushed.
  - Runs unit tests
  - Uses GoReleaser to build binaries for multiple platforms:
    - Linux (x86_64, aarch64)
    - macOS (x86_64, aarch64)
  - Creates a GitHub Release with the built binaries
  - Uploads release artifacts

### Creating a Release

To create a new release:

```bash
# Tag the commit
git tag v1.0.0

# Push the tag
git push origin v1.0.0
```

This will automatically trigger the release workflow.
