# Formation MCP

Use Claude Code in your terminal to manage CyVerse Discovery Environment applications and data through natural conversation.

## What is this?

Formation MCP connects Claude Code (the AI assistant in your terminal) to the CyVerse Discovery Environment. Instead of using a web browser, you can ask Claude to:

- Launch and monitor scientific applications
- Browse and manage your data files
- Check the status of running analyses
- Upload and download files

**Example:** "Launch the Cloud Shell app with my analysis folder as input" or "What files are in my home directory?"

## Quick Start

### 1. Get Formation MCP

Download the latest version for your platform:
- **macOS/Linux/Windows**: [Download from releases](https://github.com/cyverse-de/formation-mcp/releases)

Or build from source if you have Go installed:
```bash
git clone https://github.com/cyverse-de/formation-mcp.git
cd formation-mcp
go build -o formation-mcp ./cmd/formation-mcp
```

### 2. Set Up Your Credentials

Create a configuration file at `~/.claude.json`:

```json
{
  "mcpServers": {
    "formation": {
      "command": "/path/to/formation-mcp",
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_USERNAME": "your-cyverse-username",
        "FORMATION_PASSWORD": "your-cyverse-password"
      }
    }
  }
}
```

**Important:** Replace `/path/to/formation-mcp` with the actual path where you saved the binary. On Windows, use `C:\\path\\to\\formation-mcp.exe`.

### 3. Start Using Claude Code

Open your terminal and start Claude Code. The Formation MCP server will automatically connect, giving Claude access to your CyVerse environment.

Try asking Claude:
- "What applications are available?"
- "List the contents of my home folder"
- "Launch the RStudio app"

## What Can You Do?

### Work with Applications

- **Find apps**: Search for available applications by name, integrator, or type
- **Get details**: See what parameters an app requires
- **Launch apps**: Start interactive or batch applications
- **Monitor progress**: Check the status of running analyses
- **Open in browser**: Automatically open interactive apps when ready
- **Stop analyses**: Cancel running jobs

### Manage Your Data

- **Browse directories**: List files and folders in your data store
- **Read files**: View file contents
- **Upload files**: Add new files to your data store
- **Create folders**: Organize your data with directories
- **Manage metadata**: Add or update metadata on files and folders
- **Delete items**: Remove files or directories (with dry-run preview)

## Configuration Options

Formation MCP can be configured three ways (in order of priority):

### Option 1: Claude Code Config File (Recommended)

Edit `~/.claude.json` and add the Formation MCP server as shown in Quick Start above.

### Option 2: Standalone Config File

Create `~/.formation-mcp.yaml`:

```yaml
base_url: https://de.cyverse.org/formation
username: your-username
password: your-password
log_level: info
poll_interval: 5  # seconds between status checks
```

### Option 3: Environment Variables

```bash
export FORMATION_BASE_URL="https://de.cyverse.org/formation"
export FORMATION_USERNAME="your-username"
export FORMATION_PASSWORD="your-password"
export LOG_LEVEL="info"  # debug, info, warn, error
```

## Troubleshooting

### "Authentication failed" or "Login failed"

- Check that your CyVerse username and password are correct
- Verify the FORMATION_BASE_URL matches your environment:
  - Production: `https://de.cyverse.org/formation`
  - QA: `https://qa.cyverse.org/formation`

### "Connection refused" or "Cannot connect"

- Make sure you have internet access
- Verify the Formation URL is accessible from your network
- Check that the URL starts with `https://`

### Claude doesn't see Formation tools

- Confirm the `command` path in `~/.claude.json` points to the correct binary location
- Try running the binary directly to ensure it's executable: `./formation-mcp --help`
- On macOS/Linux, you may need to make it executable: `chmod +x formation-mcp`
- Restart Claude Code after changing the configuration

### Apps or analyses aren't working

- Check you have permission to access the app or data
- For detailed error information, set `LOG_LEVEL="debug"` in your configuration
- Look for error messages in Claude's output

## Example Workflows

### Launch an Analysis

```
You: Find the Cloud Shell application
Claude: [Shows available Cloud Shell apps]

You: Launch the main Cloud Shell with my analysis folder as input
Claude: [Launches app and provides analysis ID]

You: Open it in my browser when it's ready
Claude: [Monitors status and opens browser when ready]
```

### Manage Data

```
You: What's in my home folder?
Claude: [Lists directories and files]

You: Show me the contents of the analyses folder
Claude: [Displays directory contents]

You: Create a new folder called "project-2025"
Claude: [Creates the directory]
```

## Advanced Configuration

### Using a Pre-obtained Token

If you have a JWT token instead of username/password:

```json
{
  "mcpServers": {
    "formation": {
      "command": "/path/to/formation-mcp",
      "env": {
        "FORMATION_BASE_URL": "https://de.cyverse.org/formation",
        "FORMATION_TOKEN": "your-jwt-token-here"
      }
    }
  }
}
```

### Custom Poll Interval

Control how often Formation MCP checks analysis status (default: 5 seconds):

```yaml
poll_interval: 10  # Check every 10 seconds
```

### Debug Logging

Enable detailed logging to troubleshoot issues:

```bash
export LOG_LEVEL="debug"
```

Or in your config file:
```yaml
log_level: debug
```

## Platform Support

Formation MCP runs on:
- macOS (Intel and Apple Silicon)
- Linux (x86_64, arm64)
- Windows (x86_64)

## For Developers

### Building from Source

Requirements:
- Go 1.25.3 or later
- Optional: `just` command runner

```bash
# Build
go build -o formation-mcp ./cmd/formation-mcp

# Run tests
go test ./...

# With just
just build
just test
```

### Project Structure

```
formation-mcp/
├── cmd/formation-mcp/      # Main application entry point
├── internal/
│   ├── client/             # Formation API HTTP client
│   ├── config/             # Configuration management
│   ├── logging/            # Structured logging
│   ├── server/             # MCP server implementation
│   └── workflows/          # High-level workflow operations
```

### Development Tasks

If you have `just` installed:

```bash
just build          # Build binary
just test           # Run tests
just test-coverage  # Run tests with coverage
just lint           # Run linter
just fmt            # Format code
just check          # Run all checks
```

### API Coverage

See [API_COVERAGE_ANALYSIS.md](API_COVERAGE_ANALYSIS.md) for details on which Formation API features are implemented.

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run `go test ./...` to verify tests pass
5. Submit a pull request

## Support

- **Issues**: [GitHub Issues](https://github.com/cyverse-de/formation-mcp/issues)
- **CyVerse**: https://cyverse.org/
- **Documentation**: https://learning.cyverse.org/

## License

See [LICENSE](LICENSE) file for details.
