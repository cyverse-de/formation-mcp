# Formation MCP

A Model Context Protocol (MCP) server for interacting with the CyVerse Discovery Environment Formation API. This server provides Claude Code and other MCP-compatible clients with tools to manage applications, analyses, and data in the Discovery Environment.

**Now available in Go!** This is the Go implementation, providing a single static binary with no external dependencies. The original Python implementation is available in the `python/` directory.

## Features

- **Application Management**: List, launch, and monitor interactive (VICE) and batch applications
- **Analysis Control**: Check status, list running analyses, and stop analyses
- **Data Operations**: Browse directories, upload/download files, manage metadata in iRODS
- **Browser Integration**: Automatically open interactive apps in your default browser
- **Single Binary**: Distributed as a static binary with no external dependencies
- **Human-Readable Logging**: Structured logs in both human-friendly and JSON formats

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/cyverse-de/formation-mcp/releases).

### From Source

Requirements:
- Go 1.25.3 or later
- `just` command runner (optional, for development tasks)

```bash
# Clone the repository
git clone https://github.com/cyverse-de/formation-mcp.git
cd formation-mcp

# Build
just build
# or
go build -o formation-mcp ./cmd/formation-mcp

# Install to GOPATH/bin
just install
# or
go install ./cmd/formation-mcp
```

## Configuration

Formation MCP can be configured using environment variables, a YAML configuration file, or command-line flags. Configuration precedence (highest to lowest):

1. Command-line flags
2. Environment variables
3. Configuration file
4. Default values

### Environment Variables

```bash
# Required
export FORMATION_BASE_URL="https://de.cyverse.org/formation"

# Authentication (choose one)
export FORMATION_TOKEN="your-jwt-token"
# OR
export FORMATION_USERNAME="your-username"
export FORMATION_PASSWORD="your-password"

# Optional
export LOG_LEVEL="info"              # debug, info, warn, error
export LOG_JSON="false"              # true for JSON output
```

### Configuration File

Create `~/.formation-mcp.yaml`:

```yaml
base_url: https://de.cyverse.org/formation
username: your-username
password: your-password
log_level: info
log_json: false
poll_interval: 5  # seconds
```

### Command-Line Flags

```bash
formation-mcp \
  --base-url https://de.cyverse.org/formation \
  --username your-username \
  --password your-password \
  --log-level info \
  --poll-interval 5
```

## Claude Code Integration

Add to your Claude Code MCP configuration:

### macOS/Linux

Edit `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "formation": {
      "command": "/path/to/formation-mcp",
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

### Windows

Edit `%APPDATA%\Claude\claude_desktop_config.json` with similar configuration.

## Available Tools

### Application Management

#### `list_apps`
List available interactive VICE applications.

**Parameters:**
- `name` (optional): Filter by app name
- `limit` (optional): Maximum number of apps to return (default: 10)
- `offset` (optional): Offset for pagination (default: 0)

#### `get_app_parameters`
Get configuration and required parameters for an application.

**Parameters:**
- `app_id` (required): The application ID
- `system_id` (optional): The system ID (default: "de")

#### `launch_app_and_wait`
Launch an application and wait for it to be ready (interactive apps only).

**Parameters:**
- `app_id` (required): The application ID
- `system_id` (optional): The system ID (default: "de")
- `name` (optional): Name for the analysis
- `config` (optional): Configuration parameters for the app
- `max_wait` (optional): Maximum time to wait in seconds (default: 300)

### Analysis Management

#### `get_analysis_status`
Check the status of a running analysis.

**Parameters:**
- `analysis_id` (required): The analysis ID

#### `list_running_analyses`
List all currently running analyses.

#### `stop_analysis`
Stop a running analysis.

**Parameters:**
- `analysis_id` (required): The analysis ID to stop
- `save_outputs` (optional): Whether to save outputs before stopping (default: true)

### Data Operations

#### `browse_data`
Browse a directory or read a file from iRODS.

**Parameters:**
- `path` (required): The path to browse or read
- `offset` (optional): Offset for pagination (default: 0)
- `limit` (optional): Limit for pagination (default: 100)
- `include_metadata` (optional): Include metadata in response (default: false)

#### `create_directory`
Create a new directory in iRODS.

**Parameters:**
- `path` (required): The path for the new directory
- `metadata` (optional): Optional metadata to attach

#### `upload_file`
Upload a file to iRODS.

**Parameters:**
- `path` (required): The destination path for the file
- `content` (required): The file content
- `metadata` (optional): Optional metadata to attach

#### `set_metadata`
Add or replace metadata on an existing path.

**Parameters:**
- `path` (required): The path to set metadata on
- `metadata` (required): Metadata to set
- `replace` (optional): Whether to replace existing metadata (default: false)

#### `delete_data`
Delete a file or directory from iRODS.

**Parameters:**
- `path` (required): The path to delete
- `recurse` (optional): Whether to recursively delete directories (default: false)
- `dry_run` (optional): Preview what would be deleted (default: false)

### Utility

#### `open_in_browser`
Open a URL in the default browser.

**Parameters:**
- `url` (required): The URL to open

## Development

### Prerequisites

- Go 1.25.3+
- `just` command runner
- `golangci-lint` (for linting)

### Common Tasks

```bash
# Build the binary
just build

# Run tests
just test

# Run tests with coverage
just test-coverage

# Run linter
just lint

# Format code
just fmt

# Run all checks (format, lint, test)
just check

# Clean build artifacts
just clean

# Build for all platforms
just build-all 1.0.0

# Create distribution archives
just dist 1.0.0
```

See `just --list` for all available commands.

### Project Structure

```
formation-mcp/
├── cmd/
│   └── formation-mcp/      # Main entry point
│       └── main.go
├── internal/
│   ├── client/             # Formation API client
│   │   ├── client.go
│   │   └── types.go
│   ├── config/             # Configuration management
│   │   ├── config.go
│   │   └── config_test.go
│   ├── logging/            # Structured logging
│   │   ├── logging.go
│   │   └── logging_test.go
│   ├── server/             # MCP server implementation
│   │   └── server.go
│   └── workflows/          # High-level workflows
│       └── workflows.go
├── Justfile                # Build recipes
├── go.mod
├── go.sum
└── README.md
```

## Logging

Formation MCP uses structured logging with two output formats:

### Human-Readable (Default)

```
2025-11-03T10:15:30.123 INFO  api_call method=GET endpoint=/apps duration=123ms
2025-11-03T10:15:31.456 DEBUG launching app app_id=abc123 system_id=de
```

### JSON Format

Enable with `--log-json` flag or `LOG_JSON=true`:

```json
{"time":"2025-11-03T10:15:30.123Z","level":"INFO","msg":"api_call","method":"GET","endpoint":"/apps","duration":"123ms"}
```

Log levels: `debug`, `info`, `warn`, `error`

## Troubleshooting

### Authentication Errors

- Verify your FORMATION_BASE_URL is correct
- Check that your username/password or token is valid
- Token may have expired - try username/password authentication

### Connection Issues

- Ensure the Formation service is accessible from your network
- Check firewall settings
- Verify FORMATION_BASE_URL uses https://

### Tool Failures

- Check logs with `--log-level debug` for detailed error messages
- Verify required parameters are provided
- Ensure you have permissions for the requested operation

## License

See [LICENSE](LICENSE) file.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `just check` to ensure tests and linting pass
5. Submit a pull request

## Support

- Report issues: [GitHub Issues](https://github.com/cyverse-de/formation-mcp/issues)
- CyVerse documentation: https://cyverse.org/
