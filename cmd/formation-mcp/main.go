// Formation MCP server provides MCP tools for interacting with the CyVerse Discovery Environment.
// It enables Claude Code and other MCP-compatible clients to manage applications, analyses, and data.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cyverse-de/formation-mcp/internal/client"
	"github.com/cyverse-de/formation-mcp/internal/config"
	"github.com/cyverse-de/formation-mcp/internal/logging"
	formationServer "github.com/cyverse-de/formation-mcp/internal/server"
	"github.com/cyverse-de/formation-mcp/internal/workflows"
	"github.com/mark3labs/mcp-go/server"
)

const version = "1.0.0"

func main() {
	// Define CLI flags
	var (
		configFile   = flag.String("config", "", "Path to configuration file")
		baseURL      = flag.String("base-url", "", "Formation base URL (overrides config file and env var)")
		token        = flag.String("token", "", "Formation JWT token (overrides config file and env var)")
		username     = flag.String("username", "", "Formation username (overrides config file and env var)")
		password     = flag.String("password", "", "Formation password (overrides config file and env var)")
		logLevel     = flag.String("log-level", "", "Log level: debug, info, warn, error (default: info)")
		logJSON      = flag.Bool("log-json", false, "Output logs in JSON format")
		showVersion  = flag.Bool("version", false, "Show version and exit")
		pollInterval = flag.Int("poll-interval", 0, "Analysis status poll interval in seconds (default: 5)")
	)

	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("formation-mcp version %s\n", version)
		os.Exit(0)
	}

	// Build configuration from all sources
	cliConfig := &config.Config{
		ConfigFile:   *configFile,
		BaseURL:      *baseURL,
		Token:        *token,
		Username:     *username,
		Password:     *password,
		LogLevel:     *logLevel,
		LogJSON:      *logJSON,
		PollInterval: *pollInterval,
	}

	// Load configuration with proper precedence
	cfg, err := config.Load(cliConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
		os.Exit(1)
	}

	// Setup logging
	logger := logging.Setup(os.Stderr, cfg.LogLevel, cfg.LogJSON)
	logger.Info("formation-mcp starting", "version", version)

	// Create Formation API client
	formationClient := client.NewFormationClient(
		cfg.BaseURL,
		cfg.Token,
		cfg.Username,
		cfg.Password,
	)

	// Create workflows
	pollDuration := time.Duration(cfg.PollInterval) * time.Second
	browserOpener := &workflows.SystemBrowserOpener{}
	formationWorkflows := workflows.NewFormationWorkflows(formationClient, browserOpener, pollDuration)

	// Create MCP server
	formationMCPServer := formationServer.NewFormationMCPServer(formationWorkflows, formationClient)

	// Start stdio server
	logger.Info("starting MCP stdio server")
	if err := server.ServeStdio(formationMCPServer.Server()); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	logger.Info("formation-mcp shutting down")
}
