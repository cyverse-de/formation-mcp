# Formation MCP

[![License](https://img.shields.io/github/license/cyverse-de/formation-mcp)](LICENSE)
[![Go](https://img.shields.io/badge/language-Go-00ADD8?logo=go)](https://golang.org/)
[![MCP Compatible](https://img.shields.io/badge/MCP-compatible-blueviolet)](https://modelcontextprotocol.io/)
[![CyVerse](https://img.shields.io/badge/CyVerse-Discovery%20Environment-brightgreen)](https://de.cyverse.org/)

**Formation MCP** connects AI coding assistants to the [CyVerse Discovery Environment](https://de.cyverse.org/) via the [Model Context Protocol (MCP)](https://modelcontextprotocol.io/). It lets you launch scientific applications, manage analyses, and interact with the CyVerse Data Store — all through natural language in your AI assistant of choice, without switching to a browser.

> Built on the [Formation API](https://github.com/cyverse-de/formation), Formation MCP exposes CyVerse's full job-launching and data-management surface as MCP tools compatible with Claude Code, Claude Desktop, Claude.ai (web), Cursor, VS Code Copilot, Windsurf, OpenClaw, Continue.dev, and Codex CLI.

---

## Table of Contents

- [How It Works](#how-it-works)
- [Prerequisites](#prerequisites)
- [Installation & Build](#installation--build)
  - [Download a Release](#download-a-release)
  - [Build from Source](#build-from-source)
  - [Cross-Platform Build Reference](#cross-platform-build-reference)
- [Platform Path Reference](#platform-path-reference)
- [Configuration](#configuration)
- [Transport Modes](#transport-modes)
- [Tool Reference](#tool-reference)
- [AI Environment Setup](#ai-environment-setup)
- [VM / Server Deployment](#vm--server-deployment)
  - [systemd Service](#systemd-service)
  - [nginx Reverse Proxy with TLS](#nginx-reverse-proxy-with-tls)
  - [Certbot (Let's Encrypt)](#certbot-lets-encrypt)
  - [Docker](#docker)
  - [Docker Compose](#docker-compose)
  - [Cloud Instance Quick-Start](#cloud-instance-quick-start)
- [Example Prompts](#example-prompts)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

---

## How It Works

```
AI Assistant (Claude, Cursor, etc.)
        │  MCP (stdio  OR  SSE/HTTPS)
        ▼
 formation-mcp-bin           ← this repo
        │  HTTPS REST
        ▼
 CyVerse Formation API
        │
        ▼
 CyVerse Discovery Environment
 (apps · analyses · Data Store)
```

The binary supports two transport modes selectable at runtime:

| Mode | Flag | Use case |
|---|---|---|
| **stdio** (default) | `--transport stdio` | Local AI clients (Claude Code, Cursor, etc.) |
| **SSE** | `--transport sse --port 8080` | Remote/web clients, Claude.ai, Docker, VM deployments |

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

Pre-built binaries for Linux, macOS, and Windows are on the [Releases page](https://github.com/cyverse-de/formation-mcp/releases).

```bash
# Linux / macOS — make executable
chmod +x formation-mcp-bin
sudo mv formation-mcp-bin /usr/local/bin/formation-mcp-bin
```

**macOS** — clear the quarantine flag after download:
```bash
xattr -d com.apple.quarantine formation-mcp-bin
```

**Windows** — download `formation-mcp-bin.exe` and place it on your `%PATH%` or note its full path.

---

### Build from Source

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

With [`just`](https://just.systems/):
```bash
just build          # Build binary
just test           # Run tests
just test-coverage  # Coverage report
just lint           # golangci-lint
just fmt            # gofmt
just check          # All checks
```

---

### Cross-Platform Build Reference

Cross-compile for any target from any OS:

| Target | `GOOS` | `GOARCH` | Output | Command |
|---|---|---|---|---|
| Linux x86_64 | `linux` | `amd64` | `formation-mcp-bin` | `GOOS=linux GOARCH=amd64 go build -o formation-mcp-bin ./cmd/formation-mcp` |
| Linux ARM64 | `linux` | `arm64` | `formation-mcp-bin` | `GOOS=linux GOARCH=arm64 go build -o formation-mcp-bin ./cmd/formation-mcp` |
| macOS Intel | `darwin` | `amd64` | `formation-mcp-bin` | `GOOS=darwin GOARCH=amd64 go build -o formation-mcp-bin ./cmd/formation-mcp` |
| macOS Apple Silicon | `darwin` | `arm64` | `formation-mcp-bin` | `GOOS=darwin GOARCH=arm64 go build -o formation-mcp-bin ./cmd/formation-mcp` |
| Windows x86_64 | `windows` | `amd64` | `formation-mcp-bin.exe` | `GOOS=windows GOARCH=amd64 go build -o formation-mcp-bin.exe ./cmd/formation-mcp` |

---

## Platform Path Reference

| Platform | Binary name | Key config locations |
|---|---|---|
| **Linux** | `formation-mcp-bin` | `~/.claude.json` · `~/.cursor/mcp.json` · `~/.continue/config.json` · `~/.formation-mcp.yaml` |
| **macOS** | `formation-mcp-bin` | `~/Library/Application Support/Claude/claude_desktop_config.json` · `~/.cursor/mcp.json` · `~/.formation-mcp.yaml` |
| **Windows** | `formation-mcp-bin.exe` | `%APPDATA%\Claude\claude_desktop_config.json` · `%USERPROFILE%\.cursor\mcp.json` · `%USERPROFILE%\.formation-mcp.yaml` |

---

## Configuration

Formation MCP reads credentials from three sources (highest to lowest priority):

### Option 1 — Standalone YAML (Recommended)

`~/.formation-mcp.yaml` (Linux/macOS) · `%USERPROFILE%\.formation-mcp.yaml` (Windows):

```yaml
base_url: https://de.cyverse.org/formation
username: your-cyverse-username
password: your-cyverse-password
log_level: info        # debug | info | warn | error
poll_interval: 5       # seconds between status-check polls
```

> Using this file keeps credentials out of every AI client config.

### Option 2 — Environment Variables

```bash
export FORMATION_BASE_URL="https://de.cyverse.org/formation"
export FORMATION_USERNAME="your-username"
export FORMATION_PASSWORD="your-password"
# OR use a JWT token instead:
export FORMATION_TOKEN="your-jwt-token"
export LOG_LEVEL="info"
```

### Option 3 — CLI Flags

```bash
formation-mcp-bin \
  --base-url https://de.cyverse.org/formation \
  --username your-username \
  --password your-password \
  --transport sse \
  --port 8080
```

All flags:

| Flag | Default | Description |
|---|---|---|
| `--config` | — | Path to a YAML config file |
| `--base-url` | — | Formation API base URL |
| `--username` | — | CyVerse username |
| `--password` | — | CyVerse password |
| `--token` | — | JWT token (instead of username/password) |
| `--log-level` | `info` | `debug` · `info` · `warn` · `error` |
| `--log-json` | `false` | Emit structured JSON logs |
| `--poll-interval` | `5` | Seconds between analysis status polls |
| `--transport` | `stdio` | `stdio` or `sse` |
| `--port` | `8080` | Listening port for SSE mode |
| `--version` | — | Print version and exit |

---

## Transport Modes

### stdio (local AI clients)

The default mode. The AI client spawns the binary as a subprocess and communicates over stdin/stdout. No network port required.

```bash
formation-mcp-bin --transport stdio
```

### SSE (remote / web clients)

SSE mode starts an HTTP server that speaks the MCP SSE protocol. Required for Claude.ai (web) and any remote deployment.

```bash
formation-mcp-bin --transport sse --port 8080
```

The SSE endpoint is available at `http://host:8080/sse`.

> For Claude.ai, the endpoint **must be served over HTTPS**. See [VM / Server Deployment](#vm--server-deployment) for nginx + TLS setup.

---

## Tool Reference

| Tool | Description | Key Parameters |
|---|---|---|
| `list_apps` | Search or list available DE applications | `search` (keyword, optional), `limit` |
| `get_app_parameters` | Get required/optional parameters for an app | `app_id`, `system_id` |
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

> **Prerequisite:** Build or download `formation-mcp-bin` and note its absolute path.
> If you use `~/.formation-mcp.yaml`, you can omit the `env` block from any config below.

---

<details>
<summary><strong>Claude Code (CLI)</strong></summary>

Edit `~/.claude.json` (Linux/macOS) or `%USERPROFILE%\.claude.json` (Windows):

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

Restart Claude Code, then verify: `claude mcp list`

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

Fully quit and relaunch Claude Desktop. A hammer 🔨 icon confirms MCP tools are active.

</details>

---

<details>
<summary><strong>Claude.ai (Web) — Remote MCP via HTTPS</strong></summary>

Claude.ai supports remote MCP servers over HTTPS/SSE. stdio does not work here — you must deploy the server with TLS (see [VM / Server Deployment](#vm--server-deployment)).

1. Start `formation-mcp-bin` in SSE mode behind nginx+TLS on your server
2. Go to [claude.ai](https://claude.ai) → **Settings → Integrations → Add MCP Server**
3. Paste your HTTPS SSE endpoint, e.g.:
   ```
   https://mcp.yourdomain.com/sse
   ```
4. Save. Claude.ai will connect and the Formation tools will appear in chat.

> The server must be reachable over the public internet with a valid TLS certificate. Self-signed certificates are not accepted. See [Certbot](#certbot-lets-encrypt) for free Let's Encrypt certs.

</details>

---

<details>
<summary><strong>VS Code with GitHub Copilot</strong></summary>

Create `.vscode/mcp.json` in your workspace root:

**Linux/macOS:**
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

Reload VS Code to activate.

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

Use `~/.formation-mcp.yaml` or env vars for credentials. Restart Cursor after saving.

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

Add to `~/.claude.json` under `mcpServers`:

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

Or via the `mcporter` CLI:
```bash
mcporter config add formation --command /usr/local/bin/formation-mcp-bin
```

</details>

---

<details>
<summary><strong>Continue.dev</strong></summary>

Edit `~/.continue/config.json`:

**Linux/macOS:**
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

</details>

---

<details>
<summary><strong>Codex CLI (OpenAI)</strong></summary>

Edit `~/.codex/config.toml` (Linux/macOS) or `%USERPROFILE%\.codex\config.toml` (Windows):

**Linux/macOS:**
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

**Windows:**
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

## VM / Server Deployment

Running Formation MCP in SSE mode on a VM lets Claude.ai and other web-based clients connect over HTTPS. The general stack is:

```
Claude.ai / remote client
       │ HTTPS (443)
       ▼
    nginx (TLS termination)
       │ HTTP (localhost:8080)
       ▼
  formation-mcp-bin --transport sse --port 8080
       │ HTTPS
       ▼
  CyVerse Formation API
```

---

### systemd Service

Create `/etc/systemd/system/formation-mcp.service`:

```ini
[Unit]
Description=Formation MCP Server (SSE)
After=network.target

[Service]
Type=simple
User=formation-mcp
EnvironmentFile=/etc/formation-mcp/env
ExecStart=/usr/local/bin/formation-mcp-bin --transport sse --port 8080
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
# Harden the service
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/formation-mcp

[Install]
WantedBy=multi-user.target
```

Create `/etc/formation-mcp/env` (mode `0600`, owned by the service user):

```bash
FORMATION_BASE_URL=https://de.cyverse.org/formation
FORMATION_USERNAME=your-username
FORMATION_PASSWORD=your-password
LOG_LEVEL=info
```

Enable and start:

```bash
sudo useradd --system --no-create-home formation-mcp
sudo mkdir -p /etc/formation-mcp /var/lib/formation-mcp
sudo install -m 0600 -o formation-mcp env /etc/formation-mcp/env
sudo install -m 0755 formation-mcp-bin /usr/local/bin/formation-mcp-bin
sudo systemctl daemon-reload
sudo systemctl enable --now formation-mcp
sudo systemctl status formation-mcp
```

---

### nginx Reverse Proxy with TLS

Install nginx, then create `/etc/nginx/sites-available/formation-mcp`:

```nginx
server {
    listen 80;
    server_name mcp.yourdomain.com;
    # Certbot will redirect this to HTTPS — leave it for now
}

server {
    listen 443 ssl;
    server_name mcp.yourdomain.com;

    # TLS — filled in by Certbot
    ssl_certificate     /etc/letsencrypt/live/mcp.yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/mcp.yourdomain.com/privkey.pem;
    include             /etc/letsencrypt/options-ssl-nginx.conf;
    ssl_dhparam         /etc/letsencrypt/ssl-dhparams.pem;

    location / {
        proxy_pass         http://127.0.0.1:8080;
        proxy_read_timeout 3600s;   # SSE connections stay open

        # SSE-required headers
        proxy_http_version 1.1;
        proxy_set_header   Connection '';
        proxy_set_header   Host $host;
        proxy_set_header   X-Real-IP $remote_addr;
        proxy_set_header   X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;

        # Disable buffering — critical for SSE
        proxy_buffering    off;
        add_header         Cache-Control no-cache;
        add_header         X-Accel-Buffering no;
    }
}
```

Enable the site:
```bash
sudo ln -s /etc/nginx/sites-available/formation-mcp /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

---

### Certbot (Let's Encrypt)

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d mcp.yourdomain.com
# Auto-renew is set up automatically; test with:
sudo certbot renew --dry-run
```

Your SSE endpoint will be: `https://mcp.yourdomain.com/sse`

Register it in Claude.ai: **Settings → Integrations → Add MCP Server → `https://mcp.yourdomain.com/sse`**

---

### Docker

Multi-stage `Dockerfile` (builder → minimal runtime):

```dockerfile
# Stage 1: build
FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /formation-mcp-bin ./cmd/formation-mcp

# Stage 2: runtime
FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /formation-mcp-bin /usr/local/bin/formation-mcp-bin
EXPOSE 8080
ENTRYPOINT ["formation-mcp-bin", "--transport", "sse", "--port", "8080"]
```

Build and run:
```bash
docker build -t formation-mcp:latest .
docker run -d \
  -p 8080:8080 \
  -e FORMATION_BASE_URL=https://de.cyverse.org/formation \
  -e FORMATION_USERNAME=your-username \
  -e FORMATION_PASSWORD=your-password \
  --name formation-mcp \
  formation-mcp:latest
```

---

### Docker Compose

`docker-compose.yml` with Formation MCP + nginx:

```yaml
version: "3.9"

services:
  formation-mcp:
    build: .
    restart: unless-stopped
    environment:
      FORMATION_BASE_URL: https://de.cyverse.org/formation
      FORMATION_USERNAME: ${FORMATION_USERNAME}
      FORMATION_PASSWORD: ${FORMATION_PASSWORD}
      LOG_LEVEL: info
    expose:
      - "8080"

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro
      - /etc/letsencrypt:/etc/letsencrypt:ro
    depends_on:
      - formation-mcp
```

Store credentials in a `.env` file (never commit it):
```bash
FORMATION_USERNAME=your-username
FORMATION_PASSWORD=your-password
```

Start the stack:
```bash
docker compose up -d
```

---

### Cloud Instance Quick-Start

Any Linux VM with a public IP works. Minimum spec for a lightly-used MCP server:

| Cloud | Instance type | vCPU | RAM | Notes |
|---|---|---|---|---|
| **Jetstream2** | `m3.small` | 2 | 4 GB | NSF allocation required; great for research |
| **AWS** | `t3.small` | 2 | 2 GB | Free tier eligible for first year |
| **GCP** | `e2-small` | 2 | 2 GB | Spot/preemptible cuts cost significantly |
| **Azure** | `B1s` | 1 | 1 GB | Adequate for light use |
| **DigitalOcean** | Basic 1 GB Droplet | 1 | 1 GB | Simplest setup; $6/month |

**Required firewall rules for all providers:**

| Port | Protocol | Purpose |
|---|---|---|
| 22 | TCP | SSH admin access |
| 80 | TCP | HTTP (Certbot challenge + redirect) |
| 443 | TCP | HTTPS / SSE endpoint |

After provisioning:
```bash
# Install Go, clone, build
sudo apt update && sudo apt install -y golang git nginx certbot python3-certbot-nginx
git clone https://github.com/cyverse-de/formation-mcp.git
cd formation-mcp && go build -o formation-mcp-bin ./cmd/formation-mcp
sudo mv formation-mcp-bin /usr/local/bin/

# Configure DNS: point mcp.yourdomain.com → your VM IP, then:
sudo certbot --nginx -d mcp.yourdomain.com

# Deploy systemd service (see above) and reload nginx
```

---

## Example Prompts

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
Show me what's in /iplant/home/myusername/projects.
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
Delete /iplant/home/myusername/scratch/old-test — do a dry run first so I can see what would be removed.
```

---

## Troubleshooting

### Authentication failed / Login failed

- Verify credentials at [user.cyverse.org](https://user.cyverse.org/)
- Check `FORMATION_BASE_URL`:
  - Production: `https://de.cyverse.org/formation`
  - QA: `https://qa.cyverse.org/formation`
- JWT token: ensure it hasn't expired

### Connection refused / Cannot connect

- Confirm internet access and that the Formation URL is reachable (`curl https://de.cyverse.org/formation`)
- URL must begin with `https://`

### AI client doesn't see Formation tools

- Confirm the binary path in config is correct and the file exists
- Verify it's executable: `ls -l /usr/local/bin/formation-mcp-bin`
- Test directly: `formation-mcp-bin --version`
- macOS quarantine: `xattr -d com.apple.quarantine formation-mcp-bin`
- Restart the AI client after any config change

### SSE / remote connection issues

- Confirm nginx `proxy_buffering off` and `X-Accel-Buffering no` headers are set
- Ensure `proxy_read_timeout` is long enough (SSE connections stay open)
- Check the service is running: `systemctl status formation-mcp`
- Verify the port isn't blocked: `curl http://localhost:8080/sse`
- Check nginx logs: `sudo journalctl -u nginx -f`

### Claude.ai won't connect to remote server

- The endpoint must use HTTPS with a valid (not self-signed) certificate
- Endpoint format: `https://mcp.yourdomain.com/sse`
- Test the cert: `curl -v https://mcp.yourdomain.com/sse`

### Debug logging

```bash
export LOG_LEVEL="debug"
# or in ~/.formation-mcp.yaml:
# log_level: debug
```

---

## Project Structure

```
formation-mcp/
├── cmd/formation-mcp/       # Main entry point (transport routing)
├── internal/
│   ├── client/              # Formation API HTTP client
│   ├── config/              # Config (YAML + env + flags)
│   ├── logging/             # Structured logging
│   ├── server/              # MCP server and tool definitions
│   └── workflows/           # Multi-step operations (launch + poll, etc.)
├── API_COVERAGE_ANALYSIS.md # Formation API surface coverage notes
├── Dockerfile               # Multi-stage Docker build
├── docker-compose.yml       # Compose stack (server + nginx)
├── justfile                 # Task runner
└── LICENSE
```

See [API_COVERAGE_ANALYSIS.md](API_COVERAGE_ANALYSIS.md) for Formation API coverage details.

---

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-tool`
3. Make changes, run `go test ./...` and `go vet ./...`
4. Submit a pull request

Open an issue first for significant changes.

---

## Support

- **Issues:** [GitHub Issues](https://github.com/cyverse-de/formation-mcp/issues)
- **CyVerse:** [cyverse.org](https://cyverse.org/) · [learning.cyverse.org](https://learning.cyverse.org/)
- **MCP Specification:** [modelcontextprotocol.io](https://modelcontextprotocol.io/)

---

## License

See [LICENSE](LICENSE) for details.
