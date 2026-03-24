# Formation MCP

[![License](https://img.shields.io/github/license/cyverse-de/formation-mcp)](LICENSE)
[![Go](https://img.shields.io/badge/language-Go-00ADD8?logo=go)](https://golang.org/)
[![MCP Compatible](https://img.shields.io/badge/MCP-compatible-blueviolet)](https://modelcontextprotocol.io/)
[![CyVerse](https://img.shields.io/badge/CyVerse-Discovery%20Environment-brightgreen)](https://de.cyverse.org/)

**Formation MCP** connects AI coding assistants to the [CyVerse Discovery Environment](https://de.cyverse.org/) via the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/). It lets you launch scientific applications, manage analyses, and interact with the CyVerse Data Store — all through natural language in your AI assistant of choice, without switching to a browser.

> Built on the [Formation API](https://github.com/cyverse-de/formation), Formation MCP exposes CyVerse's full job-launching and data-management surface as MCP tools usable from Claude Code, Claude Desktop, Cursor, VS Code Copilot, Windsurf, OpenClaw, Continue.dev, and Codex CLI.

---

## Table of Contents

- [What It Does](#what-it-does)
- [Prerequisites](#prerequisites)
- [Installation & Build](#installation--build)
  - [Download a Release](#download-a-release)
  - [Build from Source](#build-from-source)
  - [Cross-Platform Build Reference](#cross-platform-build-reference)
- [Platform Path Reference](#platform-path-reference)
- [Configuration](#configuration)
- [Tool Reference](#tool-reference)
- [AI Environment Setup](#ai-environment-setup)
- [Example Prompts](#example-prompts)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

---

## What It Does

Formation MCP bridges AI assistants and the CyVerse Discovery Environment by translating MCP tool calls into Formation API requests. Once registered, your AI assistant can:

- **Search and launch** any DE application with typed parameters
- **Monitor analyses** — polling status until complete, then optionally opening the result in a browser
- **Browse, upload, and organize** files in the CyVerse Data Store
- **Apply metadata** to files and folders
- **Cancel runaway jobs** without touching the web UI

---

## Prerequisites

| Requirement | Notes |
|---|---|
| CyVerse account | Free at [cyverse.org](https://cyverse.org/) |
| Network access to CyVerse DE | Production: `https://de.cyverse.org` |
| Go ≥ 1.21 | Only needed if building from source |
| One supported AI client | See [AI Environment Setup](#ai-environment-setup) |

---

## Installation & Build

### Download a Release

Pre-built binaries for Linux, macOS, and Windows are available on the [Releases page](https://github.com/cyverse-de/formation-mcp/releases). Download the binary for your platform and make it executable:

```bash
# Linux / macOS
chmod +x formation-mcp-bin
# Optionally move to PATH
sudo mv formation-mcp-bin /usr/local/bin/formation-mcp-bin
```

On **macOS** you may need to clear the quarantine flag after download:
```bash
xattr -d com.apple.quarantine formation-mcp-bin
```

On **Windows**, download `formation-mcp-bin.exe` and place it somewhere on your `%PATH%` or note the full path for config.

---

### Build from Source

Clone and build with Go (1.21+):

```bash
git clone https://github.com/cyverse-de/formation-mcp.git
cd formation-mcp
```

**Linux:**
```bash
go build -o formation-mcp-bin ./cmd/formation-mcp
```

**macOS:**
```bash
GOOS=darwin go build -o formation-mcp-bin ./cmd/formation-mcp
```

**Windows (PowerShell):**
```powershell
go build -o formation-mcp-bin.exe ./cmd/formation-mcp
```

If you have [`just`](https://just.systems/) installed:
```bash
just build        # Build binary
just test         # Run tests
just lint         # Run linter
just fmt          # Format code
just check        # All checks
just test-coverage
```

---

### Cross-Platform Build Reference

You can cross-compile for any target from any OS using `GOOS` and `GOARCH`:

| Target OS | `GOOS` | `GOARCH` | Output binary | Example command |
|---|---|---|---|---|
| Linux x86_64 | `linux` | `amd64` | `formation-mcp-bin` | `GOOS=linux GOARCH=amd64 go build -o formation-mcp-bin ./cmd/formation-mcp` |
| Linux ARM64 | `linux` | `arm64` | `formation-mcp-bin` | `GOOS=linux GOARCH=arm64 go build -o formation-mcp-bin ./cmd/formation-mcp` |
| macOS Intel | `darwin` | `amd64` | `formation-mcp-bin` | `GOOS=darwin GOARCH=amd64 go build -o formation-mcp-bin ./cmd/formation-mcp` |
| macOS Apple Silicon | `darwin` | `arm64` | `formation-mcp-bin` | `GOOS=darwin GOARCH=arm64 go build -o formation-mcp-bin ./cmd/formation-mcp` |
| Windows x86_64 | `windows` | `amd64` | `formation-mcp-bin.exe` | `GOOS=windows GOARCH=amd64 go build -o formation-mcp-bin.exe ./cmd/formation-mcp` |

---

## Platform Path Reference

| Platform | Binary name | MCP client config location |
|---|---|---|
| **Linux** | `formation-mcp-bin` | `~/.claude.json` · `~/.cursor/mcp.json` · `~/.continue/config.json` |
| **macOS** | `formation-mcp-bin` | `~/Library/Application Support/Claude/claude_desktop_config.json` · `~/.cursor/mcp.json` |
| **Windows** | `formation-mcp-bin.exe` | `%APPDATA%\Claude\claude_desktop_config.json` · `%USERPROFILE%\.cursor\mcp.json` |

Standalone Formation config (all platforms):

| Platform | Config file |
|---|---|
| Linux / macOS | `~/.formation-mcp.yaml` |
| Windows | `%USERPROFILE%\.formation-mcp.yaml` |

---

## Configuration

Formation MCP reads credentials from three sources (highest to lowest priority):

### Option 1 — Standalone YAML (Recommended)

Create `~/.formation-mcp.yaml` (Linux/macOS) or `%USERPROFILE%\.formation-mcp.yaml` (Windows):

```yaml
base_url: https://de.cyverse.org/formation
username: your-cyverse-username
password: your-cyverse-password
log_level: info        # debug | info | warn | error
poll_interval: 5       # seconds between status-check polls
```

> **Tip:** Using this file keeps credentials out of your AI client config entirely.

### Option 2 — Environment Variables

```bash
export FORMATION_BASE_URL="https://de.cyverse.org/formation"
export FORMATION_USERNAME="your-username"
export FORMATION_PASSWORD="your-password"
# OR use a JWT token instead of username/password:
export FORMATION_TOKEN="your-jwt-token"
export LOG_LEVEL="info"
```

### Option 3 — Inline in AI Client Config

Credentials can be passed directly in the `env` block of each client's config (shown in each platform section below). Useful for quick setups, but less secure than a separate YAML file.

---

## Tool Reference

| Tool | Description | Key Parameters |
|---|---|---|
| `list_apps` | Search or list available DE applications | `search` (optional keyword), `limit` |
| `get_app_parameters` | Get required/optional parameters for a specific app | `app_id`, `system_id` |
| `launch_app_and_wait` | Launch a DE app and poll until ready/complete | `app_id`, `system_id`, `job_name`, `inputs`, `parameters` |
| `get_analysis_status` | Check status of a running analysis | `analysis_id` |
| `list_running_analyses` | List all currently running analyses | — |
| `stop_analysis` | Cancel/stop a running analysis | `analysis_id` |
| `open_in_browser` | Open an interactive app URL in the browser | `analysis_id` |
| `browse_data` | List files and folders in CyVerse Data Store | `path` |
| `create_directory` | Create a new directory in Data Store | `path` |
| `upload_file` | Upload a local file to Data Store | `local_path`, `remote_path` |
| `set_metadata` | Add or update metadata on a file or folder | `path`, `attribute`, `value`, `unit` |
| `delete_data` | Delete a file or directory (supports dry-run) | `path`, `dry_run` |

---

## AI Environment Setup

> **Prerequisite for all clients:** Build or download `formation-mcp-bin` and note its absolute path.
> If you use `~/.formation-mcp.yaml`, you can omit the `env` block from any config below.

---

<details>
<summary><strong>Claude Code (CLI)</strong></summary>

Claude Code reads MCP servers from `~/.claude.json`. Add a `mcpServers` entry:

**Linux/macOS** (`~/.claude.json`):
```json
{
  "mcpServers": {
    "formation": {
      "command": "/usr/local/bin/formation-mcp-bin",
      "args": [],
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

**Windows** (`%USERPROFILE%\.claude.json`):
```json
{
  "mcpServers": {
    "formation": {
      "command": "C:\\tools\\formation-mcp-bin.exe",
      "args": [],
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

Restart Claude Code after editing. Verify with: `claude mcp list`

</details>

---

<details>
<summary><strong>Claude Desktop</strong></summary>

Open Claude Desktop → **Settings → Developer → Edit Config**.

**macOS** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "formation": {
      "command": "/usr/local/bin/formation-mcp-bin",
      "args": [],
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

**Windows** (`%APPDATA%\Claude\claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "formation": {
      "command": "C:\\tools\\formation-mcp-bin.exe",
      "args": [],
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

Fully quit and relaunch Claude Desktop. You should see a hammer 🔨 icon indicating MCP tools are loaded.

> If you use `~/.formation-mcp.yaml`, omit the `env` block entirely.

</details>

---

<details>
<summary><strong>VS Code with GitHub Copilot</strong></summary>

Create `.vscode/mcp.json` in your workspace root (or add to user `settings.json` under `mcp.servers`):

```json
{
  "servers": {
    "formation": {
      "type": "stdio",
      "command": "/usr/local/bin/formation-mcp-bin",
      "args": [],
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

**Windows:**
```json
{
  "servers": {
    "formation": {
      "type": "stdio",
      "command": "C:\\tools\\formation-mcp-bin.exe",
      "args": [],
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

Reload VS Code. The Formation tools will appear in Copilot Chat when MCP is enabled.

</details>

---

<details>
<summary><strong>Cursor</strong></summary>

Edit `~/.cursor/mcp.json` (Linux/macOS) or `%USERPROFILE%\.cursor\mcp.json` (Windows):

**Linux/macOS:**
```json
{
  "mcpServers": {
    "formation": {
      "command": "/usr/local/bin/formation-mcp-bin",
      "args": []
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "formation": {
      "command": "C:\\tools\\formation-mcp-bin.exe",
      "args": []
    }
  }
}
```

> Use `~/.formation-mcp.yaml` or environment variables to supply credentials; Cursor's MCP config doesn't have an `env` field in all versions.

Restart Cursor after saving.

</details>

---

<details>
<summary><strong>Windsurf / Codeium</strong></summary>

Edit `~/.codeium/windsurf/mcp_config.json` (Linux/macOS) or `%USERPROFILE%\.codeium\windsurf\mcp_config.json` (Windows):

**Linux/macOS:**
```json
{
  "mcpServers": {
    "formation": {
      "command": "/usr/local/bin/formation-mcp-bin",
      "args": [],
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "formation": {
      "command": "C:\\tools\\formation-mcp-bin.exe",
      "args": [],
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

Restart Windsurf to apply.

</details>

---

<details>
<summary><strong>OpenClaw</strong></summary>

OpenClaw uses the same `~/.claude.json` format as Claude Code. Add to the `mcpServers` object:

```json
{
  "mcpServers": {
    "formation": {
      "type": "stdio",
      "command": "/usr/local/bin/formation-mcp-bin",
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-username",
        "FORMATION_PASSWORD": "your-password"
      }
    }
  }
}
```

Alternatively, use the `mcporter` CLI:
```bash
mcporter config add formation --command /usr/local/bin/formation-mcp-bin
```

</details>

---

<details>
<summary><strong>Continue.dev</strong></summary>

Edit `~/.continue/config.json` and add under `experimental.modelContextProtocolServers`:

```json
{
  "experimental": {
    "modelContextProtocolServers": [
      {
        "transport": {
          "type": "stdio",
          "command": "/usr/local/bin/formation-mcp-bin",
          "args": [],
          "env": {
            "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
            "FORMATION_USERNAME": "your-username",
            "FORMATION_PASSWORD": "your-password"
          }
        }
      }
    ]
  }
}
```

**Windows:**
```json
{
  "experimental": {
    "modelContextProtocolServers": [
      {
        "transport": {
          "type": "stdio",
          "command": "C:\\tools\\formation-mcp-bin.exe",
          "args": []
        }
      }
    ]
  }
}
```

Reload VS Code / your editor after saving.

</details>

---

<details>
<summary><strong>Codex CLI (OpenAI)</strong></summary>

[Codex CLI](https://github.com/openai/codex) supports MCP via `~/.codex/config.toml` (Linux/macOS) or `%USERPROFILE%\.codex\config.toml` (Windows).

Add an `[[mcp_servers]]` entry:

**Linux/macOS** (`~/.codex/config.toml`):
```toml
[[mcp_servers]]
name = "formation"
command = "/usr/local/bin/formation-mcp-bin"
args = []

[mcp_servers.env]
FORMATION_BASE_URL = "https://de.cyverse.org/formation"
FORMATION_USERNAME = "your-username"
FORMATION_PASSWORD = "your-password"
```

**Windows** (`%USERPROFILE%\.codex\config.toml`):
```toml
[[mcp_servers]]
name = "formation"
command = "C:\\tools\\formation-mcp-bin.exe"
args = []

[mcp_servers.env]
FORMATION_BASE_URL = "https://de.cyverse.org/formation"
FORMATION_USERNAME = "your-username"
FORMATION_PASSWORD = "your-password"
```

> **Note:** The OpenAI ChatGPT desktop app does not currently support MCP natively. Codex CLI does.

</details>

---

## Example Prompts

These natural-language prompts work once Formation MCP is connected to your assistant:

```
What QGIS or RStudio applications are available in the Discovery Environment?
```

```
Launch the Cloud Shell app and wait until it's ready, then open it in my browser.
```

```
What analyses do I have running right now?
```

```
Check the status of analysis 8a3f2b10-abc1-4def-9012-abcdef012345.
```

```
Show me what's in my /iplant/home/myusername/projects folder.
```

```
Create a directory at /iplant/home/myusername/projects/new-experiment.
```

```
Upload /home/me/data/sample.csv to /iplant/home/myusername/projects/new-experiment/.
```

```
Add metadata to /iplant/home/myusername/data/results.csv — attribute "experiment", value "run-42".
```

```
Stop the analysis named "overnight-blast-job" if it's still running.
```

```
Delete /iplant/home/myusername/scratch/old-test — but do a dry run first so I can see what would be removed.
```

---

## Troubleshooting

### Authentication failed / Login failed

- Verify your CyVerse username and password at [user.cyverse.org](https://user.cyverse.org/)
- Check `FORMATION_BASE_URL` matches your target environment:
  - Production: `https://de.cyverse.org/formation`
  - QA: `https://qa.cyverse.org/formation`
- If using a token, ensure it hasn't expired

### Connection refused / Cannot connect

- Confirm you have internet access and the Formation URL is reachable
- Check that the URL begins with `https://`

### Claude / AI client doesn't see Formation tools

- Confirm the `command` path in config points to the correct binary and it exists
- Verify the binary is executable: `ls -l /usr/local/bin/formation-mcp-bin`
- Test the binary directly: `formation-mcp-bin --help`
- macOS: check quarantine: `xattr -l formation-mcp-bin` (clear with `xattr -d com.apple.quarantine formation-mcp-bin`)
- Restart the AI client fully after any config change

### Debug logging

Enable verbose output to diagnose API errors:

```bash
export LOG_LEVEL="debug"
```

Or in `~/.formation-mcp.yaml`:
```yaml
log_level: debug
```

### Analysis stuck / never completes

- Increase `poll_interval` in your YAML if the DE is under load
- Use `get_analysis_status` directly to check state
- Use `stop_analysis` to cancel a hung job

---

## Project Structure

```
formation-mcp/
├── cmd/formation-mcp/       # Main application entry point
├── internal/
│   ├── client/              # Formation API HTTP client
│   ├── config/              # Configuration (YAML + env parsing)
│   ├── logging/             # Structured logging
│   ├── server/              # MCP server and tool definitions
│   └── workflows/           # High-level multi-step operations
├── API_COVERAGE_ANALYSIS.md # Formation API surface coverage notes
├── justfile                 # Task runner (optional)
└── LICENSE
```

See [API_COVERAGE_ANALYSIS.md](API_COVERAGE_ANALYSIS.md) for details on which Formation API endpoints are implemented.

---

## Contributing

Contributions are welcome!

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-tool`
3. Make changes and run `go test ./...`
4. Run `go vet ./...` and `golangci-lint run` (if available)
5. Submit a pull request

Please open an issue first for significant changes.

---

## Support

- **Issues:** [GitHub Issues](https://github.com/cyverse-de/formation-mcp/issues)
- **CyVerse:** [cyverse.org](https://cyverse.org/)
- **CyVerse Learning:** [learning.cyverse.org](https://learning.cyverse.org/)
- **MCP Specification:** [modelcontextprotocol.io](https://modelcontextprotocol.io/)

---

## License

See [LICENSE](LICENSE) for details.
