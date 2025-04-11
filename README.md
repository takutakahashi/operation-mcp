# Operations CLI Tool

A CLI tool for executing operations defined in a YAML configuration file.

## Features

- Dynamic command generation based on YAML configuration
- Hierarchical command structure with subcommands
- Parameter validation and templating
- Danger level management for sensitive operations
- Configurable action types (confirm, timeout, force)
- Remote execution via SSH

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

# Running commands on a remote host via SSH
operations --remote --host example.com --user admin kubectl_get_pod --namespace my-namespace

# Using a specific SSH key
operations --remote --host example.com --user admin --key ~/.ssh/custom_key kubectl_get_pod --namespace my-namespace
```

### Remote Execution Options

You can execute commands on a remote host using the following options:

```bash
--remote            Enable remote execution via SSH
--host string       SSH remote host (required in remote mode)
--user string       SSH username (default: current user)
--key string        Path to SSH private key (default: ~/.ssh/id_rsa)
--password string   SSH password (key authentication is preferred)
--port int          SSH port (default: 22)
--timeout duration  SSH connection timeout (default: 10s)
--verify-host       Verify host key (default: true)
```

You can also set SSH options in the configuration file:

```yaml
ssh:
  host: example.com
  user: username
  key: ~/.ssh/id_rsa
  port: 22
  verify_host: true
  timeout: 10
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
