# Formation MCP Server

MCP (Model Context Protocol) server for the CyVerse Formation API. Enables AI assistants like Claude to discover, launch, and manage interactive VICE applications.

## Features

- **List Apps**: Browse available interactive applications with filtering
- **Launch & Wait**: Launch apps and automatically wait for them to become ready
- **Status Monitoring**: Check analysis status and readiness
- **Browser Integration**: Open running analyses in browser
- **Analysis Management**: List running analyses and control their lifecycle

## Installation

```bash
# Install dependencies
uv sync

# Run the MCP server (for testing)
uv run formation-mcp
```

## Configuration

The server requires the following environment variables:

**Required:**
- `FORMATION_BASE_URL`: Base URL of the Formation service (e.g., `https://de.cyverse.org/formation`)

**Authentication (choose one):**

*Option 1 - Username/Password (Recommended):*
- `FORMATION_USERNAME`: Your CyVerse username
- `FORMATION_PASSWORD`: Your CyVerse password

The server will automatically obtain and refresh JWT tokens as needed.

*Option 2 - Pre-obtained Token:*
- `FORMATION_TOKEN`: JWT authentication token from Formation's `/login` endpoint

## Usage with Claude Code

Add to `~/.claude.json`:

**Using username/password (recommended):**
```json
{
  "mcpServers": {
    "formation": {
      "command": "uv",
      "args": [
        "run",
        "--directory",
        "/home/johnw/work/src/github.com/cyverse-de/formation-mcp",
        "formation-mcp"
      ],
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

**Or using a pre-obtained token:**
```json
{
  "mcpServers": {
    "formation": {
      "command": "uv",
      "args": [
        "run",
        "--directory",
        "/home/johnw/work/src/github.com/cyverse-de/formation-mcp",
        "formation-mcp"
      ],
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_TOKEN": "your-jwt-token"
      }
    }
  }
}
```

## Available Tools

### `list_apps`
List available interactive VICE applications with optional filtering.

**Parameters:**
- `name` (optional): Filter by app name
- `limit` (optional): Maximum number of results (default: 10)
- `offset` (optional): Pagination offset (default: 0)

### `launch_app_and_wait`
Launch an interactive application and wait for it to become ready.

**Parameters:**
- `app_id` (required): UUID of the app to launch
- `name` (optional): Custom name for the analysis
- `max_wait` (optional): Maximum seconds to wait (default: 300)

**Returns:**
- `analysis_id`: UUID of the created analysis
- `url`: URL of the running application
- `status`: Current status

### `get_analysis_status`
Check the status of a running analysis.

**Parameters:**
- `analysis_id` (required): UUID of the analysis

### `list_running_analyses`
List all currently running analyses.

### `open_in_browser`
Open an analysis URL in the default browser.

**Parameters:**
- `url` (required): URL to open

## Development

```bash
# Install dev dependencies
uv sync

# Run tests
uv run pytest

# Run linter
uv run ruff check --fix

# Format code
uv run ruff format
```

## Architecture

- `server.py`: MCP server implementation with tool definitions
- `client.py`: HTTP client for Formation REST API
- `workflows.py`: High-level workflows combining multiple API calls
- `config.py`: Configuration from environment variables

## License

See LICENSE file in the repository root.
