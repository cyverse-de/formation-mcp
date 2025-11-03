// Package workflows provides high-level workflow operations for the Formation API.
package workflows

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime"
	"time"

	"github.com/cyverse-de/formation-mcp/internal/client"
)

// FormationWorkflows provides high-level workflow operations.
type FormationWorkflows struct {
	client       *client.FormationClient
	pollInterval time.Duration
}

// NewFormationWorkflows creates a new workflows instance.
func NewFormationWorkflows(c *client.FormationClient, pollInterval time.Duration) *FormationWorkflows {
	return &FormationWorkflows{
		client:       c,
		pollInterval: pollInterval,
	}
}

// LaunchResult represents the result of launching an app.
type LaunchResult struct {
	AnalysisID     string
	Name           string
	Status         string
	URL            string
	MissingParams  []string
	IsInteractive  bool
}

// LaunchAndWait launches an app and waits for it to be ready (if interactive).
// For batch jobs, it returns immediately after launch.
func (w *FormationWorkflows) LaunchAndWait(ctx context.Context, appID, systemID, name string, config client.LaunchConfig, maxWait time.Duration) (*LaunchResult, error) {
	// Get app parameters to determine if required params are missing
	params, err := w.client.GetAppParameters(ctx, systemID, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to get app parameters: %w", err)
	}

	// Check for missing required parameters
	missingParams := w.checkMissingParams(params, config)
	if len(missingParams) > 0 {
		return &LaunchResult{
			MissingParams: missingParams,
		}, nil
	}

	// Determine if this is an interactive app
	isInteractive := w.isInteractiveJobType(params.OverallJobType)

	slog.Info("launching app", "app_id", appID, "system_id", systemID, "job_type", params.OverallJobType, "interactive", isInteractive)

	// Build submission with proper defaults
	// Formation API will auto-generate name and output_dir if not provided
	submission := client.LaunchSubmission{
		Name:   name,
		Config: config,
		// Email will be resolved from JWT token by Formation API
		// Debug defaults to false in Formation API
		// Notify defaults to true in Formation API
	}

	// Launch the app
	launchResp, err := w.client.LaunchApp(ctx, systemID, appID, submission)
	if err != nil {
		return nil, fmt.Errorf("failed to launch app: %w", err)
	}

	result := &LaunchResult{
		AnalysisID:    launchResp.AnalysisID,
		Name:          launchResp.Name,
		Status:        launchResp.Status,
		IsInteractive: isInteractive,
	}

	// For batch jobs, return immediately
	if !isInteractive {
		slog.Info("batch job launched", "analysis_id", result.AnalysisID)
		return result, nil
	}

	// For interactive apps, poll until URL is ready or timeout
	slog.Info("waiting for interactive app to be ready", "analysis_id", result.AnalysisID, "max_wait", maxWait)

	deadline := time.Now().Add(maxWait)
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return result, fmt.Errorf("timeout waiting for app to be ready after %v", maxWait)
			}

			status, err := w.client.GetAnalysisStatus(ctx, result.AnalysisID)
			if err != nil {
				slog.Warn("failed to get analysis status", "analysis_id", result.AnalysisID, "error", err)
				continue
			}

			result.Status = status.Status
			result.URL = status.URL

			slog.Debug("analysis status", "analysis_id", result.AnalysisID, "status", status.Status, "url_ready", status.URLReady)

			if status.URLReady && status.URL != "" {
				slog.Info("interactive app ready", "analysis_id", result.AnalysisID, "url", status.URL)
				return result, nil
			}

			// Check for failed status
			if status.Status == "Failed" || status.Status == "Canceled" {
				return result, fmt.Errorf("analysis failed with status: %s", status.Status)
			}
		}
	}
}

// GetRunningAnalyses retrieves all running analyses.
func (w *FormationWorkflows) GetRunningAnalyses(ctx context.Context) ([]client.Analysis, error) {
	return w.client.ListAnalyses(ctx, "Running")
}

// StopAnalysis stops a running analysis.
func (w *FormationWorkflows) StopAnalysis(ctx context.Context, analysisID string, saveOutputs bool) error {
	operation := "exit"
	if saveOutputs {
		operation = "save_and_exit"
	}

	return w.client.ControlAnalysis(ctx, analysisID, operation, saveOutputs)
}

// OpenInBrowser opens a URL in the default browser.
func (w *FormationWorkflows) OpenInBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	slog.Info("opening url in browser", "url", url)
	return cmd.Start()
}

// checkMissingParams checks if any required parameters are missing from the config.
func (w *FormationWorkflows) checkMissingParams(params *client.AppParameters, config client.LaunchConfig) []string {
	var missing []string

	for _, group := range params.Groups {
		for _, param := range group.Parameters {
			if param.Required {
				// Check if the parameter is in the config
				if _, ok := config[param.ID]; !ok {
					missing = append(missing, param.Name)
				}
			}
		}
	}

	return missing
}

// isInteractiveJobType determines if a job type is interactive (VICE).
func (w *FormationWorkflows) isInteractiveJobType(jobType string) bool {
	// Interactive job types are typically "Interactive" or contain "VICE"
	return jobType == "Interactive" || jobType == "interactive" ||
		   jobType == "VICE" || jobType == "vice"
}

// BrowseDataWithFormat browses data and returns formatted output.
// For directories, it returns a structured listing.
// For files, it returns the content with optional metadata.
func (w *FormationWorkflows) BrowseDataWithFormat(ctx context.Context, path string, offset, limit int, includeMetadata bool) (interface{}, bool, error) {
	result, err := w.client.BrowseData(ctx, path, offset, limit, includeMetadata)
	if err != nil {
		return nil, false, err
	}

	// Check if it's a directory or file
	switch v := result.(type) {
	case *client.DirectoryContents:
		return v, true, nil // true = is directory
	case *client.FileContent:
		return v, false, nil // false = is file
	default:
		return result, false, nil
	}
}
