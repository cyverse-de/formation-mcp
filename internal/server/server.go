// Package server provides the MCP server implementation for Formation.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/cyverse-de/formation-mcp/internal/client"
	"github.com/cyverse-de/formation-mcp/internal/workflows"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// FormationMCPServer wraps the MCP server with Formation-specific functionality.
type FormationMCPServer struct {
	server    *server.MCPServer
	workflows *workflows.FormationWorkflows
	client    *client.FormationClient
}

// NewFormationMCPServer creates a new Formation MCP server.
func NewFormationMCPServer(workflows *workflows.FormationWorkflows, c *client.FormationClient) *FormationMCPServer {
	mcpServer := server.NewMCPServer(
		"formation-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s := &FormationMCPServer{
		server:    mcpServer,
		workflows: workflows,
		client:    c,
	}

	// Register all 13 tools
	s.registerTools()

	return s
}

// Server returns the underlying MCP server.
func (s *FormationMCPServer) Server() *server.MCPServer {
	return s.server
}

// registerTools registers all Formation MCP tools.
func (s *FormationMCPServer) registerTools() {
	// App management tools
	s.server.AddTool(s.listAppsTool(), s.handleListApps)
	s.server.AddTool(s.getAppParametersTool(), s.handleGetAppParameters)
	s.server.AddTool(s.launchAppAndWaitTool(), s.handleLaunchAppAndWait)
	s.server.AddTool(s.getAnalysisStatusTool(), s.handleGetAnalysisStatus)
	s.server.AddTool(s.listRunningAnalysesTool(), s.handleListRunningAnalyses)
	s.server.AddTool(s.stopAnalysisTool(), s.handleStopAnalysis)
	s.server.AddTool(s.openInBrowserTool(), s.handleOpenInBrowser)

	// Data management tools
	s.server.AddTool(s.browseDataTool(), s.handleBrowseData)
	s.server.AddTool(s.createDirectoryTool(), s.handleCreateDirectory)
	s.server.AddTool(s.uploadFileTool(), s.handleUploadFile)
	s.server.AddTool(s.setMetadataTool(), s.handleSetMetadata)
	s.server.AddTool(s.deleteDataTool(), s.handleDeleteData)
}

// Tool definitions

func (s *FormationMCPServer) listAppsTool() mcp.Tool {
	return mcp.Tool{
		Name:        "list_apps",
		Description: "List available interactive VICE applications",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Optional filter by app name",
				},
				"integrator": map[string]interface{}{
					"type":        "string",
					"description": "Optional filter by integrator username (e.g., 'John Wregglesworth')",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Optional filter by app description",
				},
				"job_type": map[string]interface{}{
					"type":        "string",
					"description": "Optional filter by job type (Interactive, DE, OSG, Tapis)",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of apps to return (default 10)",
					"default":     10,
				},
				"offset": map[string]interface{}{
					"type":        "integer",
					"description": "Offset for pagination (default 0)",
					"default":     0,
				},
			},
		},
	}
}

func (s *FormationMCPServer) getAppParametersTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_app_parameters",
		Description: "Get the configuration and required parameters for an application",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"app_id": map[string]interface{}{
					"type":        "string",
					"description": "The application ID",
				},
				"system_id": map[string]interface{}{
					"type":        "string",
					"description": "The system ID (default: de)",
					"default":     "de",
				},
			},
			Required: []string{"app_id"},
		},
	}
}

func (s *FormationMCPServer) launchAppAndWaitTool() mcp.Tool {
	return mcp.Tool{
		Name:        "launch_app_and_wait",
		Description: "Launch an application and wait for it to be ready (interactive apps only). Returns immediately for batch jobs.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"app_id": map[string]interface{}{
					"type":        "string",
					"description": "The application ID",
				},
				"system_id": map[string]interface{}{
					"type":        "string",
					"description": "The system ID (default: de)",
					"default":     "de",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name for the analysis",
				},
				"config": map[string]interface{}{
					"type":        "object",
					"description": "Configuration parameters for the app",
				},
				"max_wait": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum time to wait in seconds (default 300)",
					"default":     300,
				},
			},
			Required: []string{"app_id"},
		},
	}
}

func (s *FormationMCPServer) getAnalysisStatusTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_analysis_status",
		Description: "Check the status of a running analysis",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"analysis_id": map[string]interface{}{
					"type":        "string",
					"description": "The analysis ID",
				},
			},
			Required: []string{"analysis_id"},
		},
	}
}

func (s *FormationMCPServer) listRunningAnalysesTool() mcp.Tool {
	return mcp.Tool{
		Name:        "list_running_analyses",
		Description: "List analyses filtered by status",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"status": map[string]interface{}{
					"type":        "string",
					"description": "Status filter (default: Running). Common values: Running, Completed, Failed, Submitted, Canceled",
					"default":     "Running",
				},
			},
		},
	}
}

func (s *FormationMCPServer) stopAnalysisTool() mcp.Tool {
	return mcp.Tool{
		Name:        "stop_analysis",
		Description: "Stop a running analysis",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"analysis_id": map[string]interface{}{
					"type":        "string",
					"description": "The analysis ID to stop",
				},
				"save_outputs": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to save outputs before stopping (default true)",
					"default":     true,
				},
			},
			Required: []string{"analysis_id"},
		},
	}
}

func (s *FormationMCPServer) openInBrowserTool() mcp.Tool {
	return mcp.Tool{
		Name:        "open_in_browser",
		Description: "Open a URL in the default browser",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "The URL to open",
				},
			},
			Required: []string{"url"},
		},
	}
}

func (s *FormationMCPServer) browseDataTool() mcp.Tool {
	return mcp.Tool{
		Name:        "browse_data",
		Description: "Browse a directory or read a file from iRODS data store",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path to browse or read",
				},
				"offset": map[string]interface{}{
					"type":        "integer",
					"description": "Offset for pagination (default 0)",
					"default":     0,
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Limit for pagination (default 100)",
					"default":     100,
				},
				"include_metadata": map[string]interface{}{
					"type":        "boolean",
					"description": "Include metadata in the response (default false)",
					"default":     false,
				},
			},
			Required: []string{"path"},
		},
	}
}

func (s *FormationMCPServer) createDirectoryTool() mcp.Tool {
	return mcp.Tool{
		Name:        "create_directory",
		Description: "Create a new directory in iRODS",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path for the new directory",
				},
				"metadata": map[string]interface{}{
					"type":        "object",
					"description": "Optional metadata to attach to the directory",
				},
			},
			Required: []string{"path"},
		},
	}
}

func (s *FormationMCPServer) uploadFileTool() mcp.Tool {
	return mcp.Tool{
		Name:        "upload_file",
		Description: "Upload a file to iRODS",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The destination path for the file",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "The file content",
				},
				"metadata": map[string]interface{}{
					"type":        "object",
					"description": "Optional metadata to attach to the file",
				},
			},
			Required: []string{"path", "content"},
		},
	}
}

func (s *FormationMCPServer) setMetadataTool() mcp.Tool {
	return mcp.Tool{
		Name:        "set_metadata",
		Description: "Add or replace metadata on an existing path",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path to set metadata on",
				},
				"metadata": map[string]interface{}{
					"type":        "object",
					"description": "Metadata to set",
				},
				"replace": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to replace existing metadata (default false)",
					"default":     false,
				},
			},
			Required: []string{"path", "metadata"},
		},
	}
}

func (s *FormationMCPServer) deleteDataTool() mcp.Tool {
	return mcp.Tool{
		Name:        "delete_data",
		Description: "Delete a file or directory from iRODS",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"path": map[string]interface{}{
					"type":        "string",
					"description": "The path to delete",
				},
				"recurse": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to recursively delete directories (default false)",
					"default":     false,
				},
				"dry_run": map[string]interface{}{
					"type":        "boolean",
					"description": "Preview what would be deleted without actually deleting (default false)",
					"default":     false,
				},
			},
			Required: []string{"path"},
		},
	}
}

// Tool handlers

// unmarshalParams is a helper function to unmarshal tool request arguments.
func unmarshalParams(request mcp.CallToolRequest, params interface{}) error {
	argsBytes, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return fmt.Errorf("failed to marshal arguments: %w", err)
	}
	if err := json.Unmarshal(argsBytes, params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}
	return nil
}

func (s *FormationMCPServer) handleListApps(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Name        string `json:"name"`
		Integrator  string `json:"integrator"`
		Description string `json:"description"`
		JobType     string `json:"job_type"`
		Limit       int    `json:"limit"`
		Offset      int    `json:"offset"`
	}
	params.Limit = 10 // default

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("listing apps", "name", params.Name, "integrator", params.Integrator, "description", params.Description, "job_type", params.JobType, "limit", params.Limit, "offset", params.Offset)

	apps, err := s.client.ListApps(ctx, params.Name, params.Integrator, params.Description, params.JobType, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}

	// Format as markdown
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("## Available Applications (%d)\n\n", len(apps)))
	for _, app := range apps {
		builder.WriteString(fmt.Sprintf("### %s\n", app.Name))
		builder.WriteString(fmt.Sprintf("- **ID**: `%s`\n", app.ID))
		builder.WriteString(fmt.Sprintf("- **System**: `%s`\n", app.SystemID))
		if app.IntegratorUsername != "" {
			builder.WriteString(fmt.Sprintf("- **Integrator**: %s\n", app.IntegratorUsername))
		}
		builder.WriteString(fmt.Sprintf("- **Description**: %s\n\n", app.Description))
	}

	return mcp.NewToolResultText(builder.String()), nil
}

func (s *FormationMCPServer) handleGetAppParameters(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		AppID    string `json:"app_id"`
		SystemID string `json:"system_id"`
	}
	params.SystemID = "de" // default

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("getting app parameters", "app_id", params.AppID, "system_id", params.SystemID)

	appParams, err := s.client.GetAppParameters(ctx, params.SystemID, params.AppID)
	if err != nil {
		return nil, err
	}

	// Format as markdown
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("## App Parameters\n\n"))
	builder.WriteString(fmt.Sprintf("**Job Type**: %s\n\n", appParams.OverallJobType))

	for _, group := range appParams.Groups {
		builder.WriteString(fmt.Sprintf("### %s\n\n", group.Label))
		for _, param := range group.Parameters {
			required := ""
			if param.Required {
				required = " (required)"
			}
			builder.WriteString(fmt.Sprintf("- **%s**%s: %s\n", param.Label, required, param.Description))
			builder.WriteString(fmt.Sprintf("  - ID: `%s`\n", param.ID))
			builder.WriteString(fmt.Sprintf("  - Type: `%s`\n", param.Type))
			if param.DefaultValue != nil {
				builder.WriteString(fmt.Sprintf("  - Default: `%v`\n", param.DefaultValue))
			}
		}
		builder.WriteString("\n")
	}

	return mcp.NewToolResultText(builder.String()), nil
}

func (s *FormationMCPServer) handleLaunchAppAndWait(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		AppID    string                 `json:"app_id"`
		SystemID string                 `json:"system_id"`
		Name     string                 `json:"name"`
		Config   map[string]interface{} `json:"config"`
		MaxWait  int                    `json:"max_wait"`
	}
	params.SystemID = "de" // default
	params.MaxWait = 300   // default

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	if params.Config == nil {
		params.Config = make(map[string]interface{})
	}

	if params.Name == "" {
		params.Name = fmt.Sprintf("analysis-%d", time.Now().Unix())
	}

	slog.Info("launching app", "app_id", params.AppID, "system_id", params.SystemID, "name", params.Name)

	result, err := s.workflows.LaunchAndWait(
		ctx,
		params.AppID,
		params.SystemID,
		params.Name,
		params.Config,
		time.Duration(params.MaxWait)*time.Second,
	)
	if err != nil {
		return nil, err
	}

	// Check for missing parameters
	if len(result.MissingParams) > 0 {
		var builder strings.Builder
		builder.WriteString("⚠️  **Missing Required Parameters**\n\n")
		builder.WriteString("The following required parameters are missing:\n\n")
		for _, param := range result.MissingParams {
			builder.WriteString(fmt.Sprintf("- %s\n", param))
		}
		builder.WriteString("\nPlease provide these parameters in the config and try again.")
		return mcp.NewToolResultText(builder.String()), nil
	}

	// Format result
	var builder strings.Builder
	if result.IsInteractive {
		builder.WriteString("✅ **Interactive App Launched Successfully**\n\n")
		builder.WriteString(fmt.Sprintf("- **Analysis ID**: `%s`\n", result.AnalysisID))
		builder.WriteString(fmt.Sprintf("- **Name**: %s\n", result.Name))
		builder.WriteString(fmt.Sprintf("- **Status**: %s\n", result.Status))
		if result.URL != "" {
			builder.WriteString(fmt.Sprintf("- **URL**: %s\n", result.URL))
		}
	} else {
		builder.WriteString("✅ **Batch Job Launched Successfully**\n\n")
		builder.WriteString(fmt.Sprintf("- **Analysis ID**: `%s`\n", result.AnalysisID))
		builder.WriteString(fmt.Sprintf("- **Name**: %s\n", result.Name))
		builder.WriteString(fmt.Sprintf("- **Status**: %s\n", result.Status))
		builder.WriteString("\nThe batch job has been submitted and is running in the background.")
	}

	return mcp.NewToolResultText(builder.String()), nil
}

func (s *FormationMCPServer) handleGetAnalysisStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		AnalysisID string `json:"analysis_id"`
	}

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("getting analysis status", "analysis_id", params.AnalysisID)

	status, err := s.client.GetAnalysisStatus(ctx, params.AnalysisID)
	if err != nil {
		return nil, err
	}

	var builder strings.Builder
	builder.WriteString("## Analysis Status\n\n")
	builder.WriteString(fmt.Sprintf("- **Analysis ID**: `%s`\n", status.AnalysisID))
	builder.WriteString(fmt.Sprintf("- **Status**: %s\n", status.Status))
	if status.URLReady {
		builder.WriteString(fmt.Sprintf("- **URL Ready**: Yes\n"))
		if status.URL != "" {
			builder.WriteString(fmt.Sprintf("- **URL**: %s\n", status.URL))
		}
	} else {
		builder.WriteString(fmt.Sprintf("- **URL Ready**: No\n"))
	}

	return mcp.NewToolResultText(builder.String()), nil
}

func (s *FormationMCPServer) handleListRunningAnalyses(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Status string `json:"status"`
	}
	params.Status = "Running" // default

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("listing analyses", "status", params.Status)

	analyses, err := s.client.ListAnalyses(ctx, params.Status)
	if err != nil {
		return nil, err
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("## %s Analyses (%d)\n\n", params.Status, len(analyses)))
	if len(analyses) == 0 {
		builder.WriteString(fmt.Sprintf("No %s analyses found.", params.Status))
	} else {
		for _, analysis := range analyses {
			builder.WriteString(fmt.Sprintf("### Analysis `%s`\n", analysis.AnalysisID))
			builder.WriteString(fmt.Sprintf("- **Analysis ID**: `%s`\n", analysis.AnalysisID))
			builder.WriteString(fmt.Sprintf("- **App ID**: `%s`\n", analysis.AppID))
			builder.WriteString(fmt.Sprintf("- **System**: `%s`\n", analysis.SystemID))
			builder.WriteString(fmt.Sprintf("- **Status**: %s\n\n", analysis.Status))
		}
	}

	return mcp.NewToolResultText(builder.String()), nil
}

func (s *FormationMCPServer) handleStopAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		AnalysisID  string `json:"analysis_id"`
		SaveOutputs bool   `json:"save_outputs"`
	}
	params.SaveOutputs = true // default

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("stopping analysis", "analysis_id", params.AnalysisID, "save_outputs", params.SaveOutputs)

	if err := s.workflows.StopAnalysis(ctx, params.AnalysisID, params.SaveOutputs); err != nil {
		return nil, err
	}

	var builder strings.Builder
	builder.WriteString("✅ **Analysis Stopped**\n\n")
	builder.WriteString(fmt.Sprintf("- **Analysis ID**: `%s`\n", params.AnalysisID))
	if params.SaveOutputs {
		builder.WriteString("- **Outputs**: Saved")
	} else {
		builder.WriteString("- **Outputs**: Not saved")
	}

	return mcp.NewToolResultText(builder.String()), nil
}

func (s *FormationMCPServer) handleOpenInBrowser(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		URL string `json:"url"`
	}

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("opening url in browser", "url", params.URL)

	if err := s.workflows.OpenInBrowser(params.URL); err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("✅ Opened %s in browser", params.URL)), nil
}

func (s *FormationMCPServer) handleBrowseData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Path            string `json:"path"`
		Offset          int    `json:"offset"`
		Limit           int    `json:"limit"`
		IncludeMetadata bool   `json:"include_metadata"`
	}
	// No default limit - 0 means unlimited for files, all entries for directories
	// Users can specify limit for pagination if needed

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("browsing data", "path", params.Path, "offset", params.Offset, "limit", params.Limit)

	result, isDir, err := s.workflows.BrowseDataWithFormat(ctx, params.Path, params.Offset, params.Limit, params.IncludeMetadata)
	if err != nil {
		return nil, err
	}

	var builder strings.Builder

	if isDir {
		dirContents := result.(*client.DirectoryContents)
		builder.WriteString(fmt.Sprintf("## Directory: %s\n\n", dirContents.Path))

		// Separate directories and files from contents
		var directories, files []client.DirectoryEntry
		for _, entry := range dirContents.Contents {
			if entry.Type == "collection" {
				directories = append(directories, entry)
			} else if entry.Type == "data_object" {
				files = append(files, entry)
			}
		}

		if len(directories) > 0 {
			builder.WriteString("### 📁 Directories\n\n")
			for _, dir := range directories {
				builder.WriteString(fmt.Sprintf("- %s\n", dir.Name))
			}
			builder.WriteString("\n")
		}

		if len(files) > 0 {
			builder.WriteString("### 📄 Files\n\n")
			for _, file := range files {
				builder.WriteString(fmt.Sprintf("- %s\n", file.Name))
			}
		}

		if len(dirContents.Contents) == 0 {
			builder.WriteString("*Empty directory*\n")
		}
	} else {
		fileContent := result.(*client.FileContent)
		builder.WriteString(fmt.Sprintf("## File: %s\n\n", fileContent.Path))
		if params.IncludeMetadata && len(fileContent.Metadata) > 0 {
			builder.WriteString("### Metadata\n\n")
			for k, v := range fileContent.Metadata {
				builder.WriteString(fmt.Sprintf("- **%s**: %v\n", k, v))
			}
			builder.WriteString("\n")
		}
		builder.WriteString("### Content\n\n```\n")
		builder.WriteString(fileContent.Content)
		builder.WriteString("\n```\n")
	}

	return mcp.NewToolResultText(builder.String()), nil
}

func (s *FormationMCPServer) handleCreateDirectory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Path     string                 `json:"path"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("creating directory", "path", params.Path)

	resp, err := s.client.CreateDirectory(ctx, params.Path, params.Metadata)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("✅ Created directory: %s", resp.Path)), nil
}

func (s *FormationMCPServer) handleUploadFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Path     string                 `json:"path"`
		Content  string                 `json:"content"`
		Metadata map[string]interface{} `json:"metadata"`
	}

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("uploading file", "path", params.Path, "size", len(params.Content))

	if err := s.client.UploadFile(ctx, params.Path, params.Content, params.Metadata); err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("✅ Uploaded file: %s (%d bytes)", params.Path, len(params.Content))), nil
}

func (s *FormationMCPServer) handleSetMetadata(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Path     string                 `json:"path"`
		Metadata map[string]interface{} `json:"metadata"`
		Replace  bool                   `json:"replace"`
	}

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("setting metadata", "path", params.Path, "replace", params.Replace)

	if err := s.client.SetMetadata(ctx, params.Path, params.Metadata, params.Replace); err != nil {
		return nil, err
	}

	action := "added to"
	if params.Replace {
		action = "replaced on"
	}

	return mcp.NewToolResultText(fmt.Sprintf("✅ Metadata %s: %s", action, params.Path)), nil
}

func (s *FormationMCPServer) handleDeleteData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		Path    string `json:"path"`
		Recurse bool   `json:"recurse"`
		DryRun  bool   `json:"dry_run"`
	}

	if err := unmarshalParams(request, &params); err != nil {
		return nil, err
	}

	slog.Info("deleting data", "path", params.Path, "recurse", params.Recurse, "dry_run", params.DryRun)

	if err := s.client.DeleteData(ctx, params.Path, params.Recurse, params.DryRun); err != nil {
		return nil, err
	}

	if params.DryRun {
		return mcp.NewToolResultText(fmt.Sprintf("✅ Dry run: Would delete %s", params.Path)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("✅ Deleted: %s", params.Path)), nil
}
