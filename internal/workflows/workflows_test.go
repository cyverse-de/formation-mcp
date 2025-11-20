package workflows

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cyverse-de/formation-mcp/internal/client"
)

// mockFormationClient implements FormationAPIClient for testing
type mockFormationClient struct {
	loginFunc             func(ctx context.Context) error
	listAppsFunc          func(ctx context.Context, name, integrator, description, jobType string, limit, offset int) ([]client.App, error)
	getAppParametersFunc  func(ctx context.Context, systemID, appID string) (*client.AppParameters, error)
	launchAppFunc         func(ctx context.Context, systemID, appID string, submission client.LaunchSubmission) (*client.LaunchResponse, error)
	getAnalysisStatusFunc func(ctx context.Context, analysisID string) (*client.AnalysisStatus, error)
	listAnalysesFunc      func(ctx context.Context, status string) ([]client.Analysis, error)
	controlAnalysisFunc   func(ctx context.Context, analysisID, operation string, saveOutputs bool) error
	browseDataFunc        func(ctx context.Context, path string, offset, limit int, includeMetadata bool) (interface{}, error)
	createDirectoryFunc   func(ctx context.Context, path string, metadata map[string]interface{}) (*client.CreateDirectoryResponse, error)
	uploadFileFunc        func(ctx context.Context, path, content string, metadata map[string]interface{}) error
	setMetadataFunc       func(ctx context.Context, path string, metadata map[string]interface{}, replace bool) error
	deleteDataFunc        func(ctx context.Context, path string, recurse, dryRun bool) error
}

func (m *mockFormationClient) Login(ctx context.Context) error {
	if m.loginFunc != nil {
		return m.loginFunc(ctx)
	}
	return nil
}

func (m *mockFormationClient) ListApps(ctx context.Context, name, integrator, description, jobType string, limit, offset int) ([]client.App, error) {
	if m.listAppsFunc != nil {
		return m.listAppsFunc(ctx, name, integrator, description, jobType, limit, offset)
	}
	return []client.App{}, nil
}

func (m *mockFormationClient) GetAppParameters(ctx context.Context, systemID, appID string) (*client.AppParameters, error) {
	if m.getAppParametersFunc != nil {
		return m.getAppParametersFunc(ctx, systemID, appID)
	}
	return &client.AppParameters{}, nil
}

func (m *mockFormationClient) LaunchApp(ctx context.Context, systemID, appID string, submission client.LaunchSubmission) (*client.LaunchResponse, error) {
	if m.launchAppFunc != nil {
		return m.launchAppFunc(ctx, systemID, appID, submission)
	}
	return &client.LaunchResponse{}, nil
}

func (m *mockFormationClient) GetAnalysisStatus(ctx context.Context, analysisID string) (*client.AnalysisStatus, error) {
	if m.getAnalysisStatusFunc != nil {
		return m.getAnalysisStatusFunc(ctx, analysisID)
	}
	return &client.AnalysisStatus{}, nil
}

func (m *mockFormationClient) ListAnalyses(ctx context.Context, status string) ([]client.Analysis, error) {
	if m.listAnalysesFunc != nil {
		return m.listAnalysesFunc(ctx, status)
	}
	return []client.Analysis{}, nil
}

func (m *mockFormationClient) ControlAnalysis(ctx context.Context, analysisID, operation string, saveOutputs bool) error {
	if m.controlAnalysisFunc != nil {
		return m.controlAnalysisFunc(ctx, analysisID, operation, saveOutputs)
	}
	return nil
}

func (m *mockFormationClient) BrowseData(ctx context.Context, path string, offset, limit int, includeMetadata bool) (interface{}, error) {
	if m.browseDataFunc != nil {
		return m.browseDataFunc(ctx, path, offset, limit, includeMetadata)
	}
	return nil, nil
}

func (m *mockFormationClient) CreateDirectory(ctx context.Context, path string, metadata map[string]interface{}) (*client.CreateDirectoryResponse, error) {
	if m.createDirectoryFunc != nil {
		return m.createDirectoryFunc(ctx, path, metadata)
	}
	return &client.CreateDirectoryResponse{}, nil
}

func (m *mockFormationClient) UploadFile(ctx context.Context, path, content string, metadata map[string]interface{}) error {
	if m.uploadFileFunc != nil {
		return m.uploadFileFunc(ctx, path, content, metadata)
	}
	return nil
}

func (m *mockFormationClient) SetMetadata(ctx context.Context, path string, metadata map[string]interface{}, replace bool) error {
	if m.setMetadataFunc != nil {
		return m.setMetadataFunc(ctx, path, metadata, replace)
	}
	return nil
}

func (m *mockFormationClient) DeleteData(ctx context.Context, path string, recurse, dryRun bool) error {
	if m.deleteDataFunc != nil {
		return m.deleteDataFunc(ctx, path, recurse, dryRun)
	}
	return nil
}

// mockBrowserOpener implements BrowserOpener for testing
type mockBrowserOpener struct {
	openFunc func(url string) error
	lastURL  string
}

func (m *mockBrowserOpener) Open(url string) error {
	m.lastURL = url
	if m.openFunc != nil {
		return m.openFunc(url)
	}
	return nil
}

// TestLaunchAndWait tests the LaunchAndWait workflow
func TestLaunchAndWait(t *testing.T) {
	tests := []struct {
		name            string
		appID           string
		systemID        string
		analysisName    string
		config          client.LaunchConfig
		params          *client.AppParameters
		launchResp      *client.LaunchResponse
		statusSequence  []*client.AnalysisStatus
		wantErr         bool
		wantMissingParams bool
		errContains     string
	}{
		{
			name:         "successful batch job launch",
			appID:        "batch-app",
			systemID:     "de",
			analysisName: "test-batch",
			config:       client.LaunchConfig{},
			params: &client.AppParameters{
				OverallJobType: "DE",
				Groups:         []client.ParameterGroup{},
			},
			launchResp: &client.LaunchResponse{
				AnalysisID: "analysis-123",
				Name:       "test-batch",
				Status:     "Submitted",
			},
			wantErr: false,
		},
		{
			name:         "successful interactive app launch",
			appID:        "interactive-app",
			systemID:     "de",
			analysisName: "test-interactive",
			config:       client.LaunchConfig{},
			params: &client.AppParameters{
				OverallJobType: "Interactive",
				Groups:         []client.ParameterGroup{},
			},
			launchResp: &client.LaunchResponse{
				AnalysisID: "analysis-456",
				Name:       "test-interactive",
				Status:     "Submitted",
			},
			statusSequence: []*client.AnalysisStatus{
				{
					AnalysisID: "analysis-456",
					Status:     "Running",
					URLReady:   false,
				},
				{
					AnalysisID: "analysis-456",
					Status:     "Running",
					URLReady:   true,
					URL:        "https://test.cyverse.run",
				},
			},
			wantErr: false,
		},
		{
			name:         "missing required parameters",
			appID:        "app-with-params",
			systemID:     "de",
			analysisName: "test-missing-params",
			config:       client.LaunchConfig{},
			params: &client.AppParameters{
				OverallJobType: "Interactive",
				Groups: []client.ParameterGroup{
					{
						ID:   "group-1",
						Name: "Input",
						Parameters: []client.Parameter{
							{
								ID:       "param1",
								Name:     "Required Param",
								Required: true,
							},
						},
					},
				},
			},
			wantMissingParams: true,
			wantErr:           false,
		},
		{
			name:         "failed analysis",
			appID:        "failing-app",
			systemID:     "de",
			analysisName: "test-failed",
			config:       client.LaunchConfig{},
			params: &client.AppParameters{
				OverallJobType: "Interactive",
				Groups:         []client.ParameterGroup{},
			},
			launchResp: &client.LaunchResponse{
				AnalysisID: "analysis-789",
				Name:       "test-failed",
				Status:     "Submitted",
			},
			statusSequence: []*client.AnalysisStatus{
				{
					AnalysisID: "analysis-789",
					Status:     "Failed",
					URLReady:   false,
				},
			},
			wantErr:     true,
			errContains: "Failed",
		},
		{
			name:         "timeout waiting for URL",
			appID:        "slow-app",
			systemID:     "de",
			analysisName: "test-timeout",
			config:       client.LaunchConfig{},
			params: &client.AppParameters{
				OverallJobType: "Interactive",
				Groups:         []client.ParameterGroup{},
			},
			launchResp: &client.LaunchResponse{
				AnalysisID: "analysis-999",
				Name:       "test-timeout",
				Status:     "Submitted",
			},
			statusSequence: []*client.AnalysisStatus{
				{
					AnalysisID: "analysis-999",
					Status:     "Running",
					URLReady:   false,
				},
			},
			wantErr:     true,
			errContains: "timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusIndex := 0

			mockClient := &mockFormationClient{
				getAppParametersFunc: func(ctx context.Context, systemID, appID string) (*client.AppParameters, error) {
					return tt.params, nil
				},
				launchAppFunc: func(ctx context.Context, systemID, appID string, submission client.LaunchSubmission) (*client.LaunchResponse, error) {
					if tt.launchResp == nil {
						return nil, errors.New("launch failed")
					}
					return tt.launchResp, nil
				},
				getAnalysisStatusFunc: func(ctx context.Context, analysisID string) (*client.AnalysisStatus, error) {
					if statusIndex < len(tt.statusSequence) {
						status := tt.statusSequence[statusIndex]
						statusIndex++
						return status, nil
					}
					// Keep returning the last status
					if len(tt.statusSequence) > 0 {
						return tt.statusSequence[len(tt.statusSequence)-1], nil
					}
					return &client.AnalysisStatus{}, nil
				},
			}

			mockBrowser := &mockBrowserOpener{}
			workflows := NewFormationWorkflows(mockClient, mockBrowser, 10*time.Millisecond)

			maxWait := 100 * time.Millisecond
			if tt.errContains == "timeout" {
				maxWait = 50 * time.Millisecond
			}

			result, err := workflows.LaunchAndWait(
				context.Background(),
				tt.appID,
				tt.systemID,
				tt.analysisName,
				tt.config,
				maxWait,
			)

			if tt.wantErr {
				if err == nil {
					t.Errorf("LaunchAndWait() expected error but got none")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("LaunchAndWait() error = %v, should contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("LaunchAndWait() unexpected error = %v", err)
				}

				if tt.wantMissingParams {
					if len(result.MissingParams) == 0 {
						t.Errorf("LaunchAndWait() expected missing params but got none")
					}
				} else {
					if result == nil {
						t.Errorf("LaunchAndWait() returned nil result")
					} else {
						if result.AnalysisID == "" {
							t.Errorf("LaunchAndWait() missing analysis ID")
						}
					}
				}
			}
		})
	}
}

// TestIsInteractiveJobType tests job type detection
func TestIsInteractiveJobType(t *testing.T) {
	tests := []struct {
		jobType string
		want    bool
	}{
		{"Interactive", true},
		{"interactive", true},
		{"VICE", true},
		{"vice", true},
		{"DE", false},
		{"OSG", false},
		{"Tapis", false},
		{"", false},
	}

	workflows := &FormationWorkflows{}

	for _, tt := range tests {
		t.Run(tt.jobType, func(t *testing.T) {
			got := workflows.isInteractiveJobType(tt.jobType)
			if got != tt.want {
				t.Errorf("isInteractiveJobType(%v) = %v, want %v", tt.jobType, got, tt.want)
			}
		})
	}
}

// TestCheckMissingParams tests parameter validation
func TestCheckMissingParams(t *testing.T) {
	tests := []struct {
		name   string
		params *client.AppParameters
		config client.LaunchConfig
		want   []string
	}{
		{
			name: "no missing params",
			params: &client.AppParameters{
				Groups: []client.ParameterGroup{
					{
						Parameters: []client.Parameter{
							{ID: "param1", Name: "Param 1", Required: true},
						},
					},
				},
			},
			config: client.LaunchConfig{"param1": "value1"},
			want:   nil,
		},
		{
			name: "missing required param",
			params: &client.AppParameters{
				Groups: []client.ParameterGroup{
					{
						Parameters: []client.Parameter{
							{ID: "param1", Name: "Param 1", Required: true},
							{ID: "param2", Name: "Param 2", Required: true},
						},
					},
				},
			},
			config: client.LaunchConfig{"param1": "value1"},
			want:   []string{"Param 2"},
		},
		{
			name: "optional params ignored",
			params: &client.AppParameters{
				Groups: []client.ParameterGroup{
					{
						Parameters: []client.Parameter{
							{ID: "param1", Name: "Param 1", Required: false},
						},
					},
				},
			},
			config: client.LaunchConfig{},
			want:   nil,
		},
	}

	workflows := &FormationWorkflows{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := workflows.checkMissingParams(tt.params, tt.config)
			if len(got) != len(tt.want) {
				t.Errorf("checkMissingParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestStopAnalysis tests stopping an analysis
func TestStopAnalysis(t *testing.T) {
	tests := []struct {
		name        string
		analysisID  string
		saveOutputs bool
		wantErr     bool
	}{
		{
			name:        "stop with save",
			analysisID:  "analysis-123",
			saveOutputs: true,
			wantErr:     false,
		},
		{
			name:        "stop without save",
			analysisID:  "analysis-456",
			saveOutputs: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calledOperation string
			mockClient := &mockFormationClient{
				controlAnalysisFunc: func(ctx context.Context, analysisID, operation string, saveOutputs bool) error {
					calledOperation = operation
					return nil
				},
			}

			workflows := NewFormationWorkflows(mockClient, &mockBrowserOpener{}, 5*time.Second)
			err := workflows.StopAnalysis(context.Background(), tt.analysisID, tt.saveOutputs)

			if tt.wantErr && err == nil {
				t.Errorf("StopAnalysis() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("StopAnalysis() unexpected error = %v", err)
			}

			expectedOp := "exit"
			if tt.saveOutputs {
				expectedOp = "save_and_exit"
			}
			if calledOperation != expectedOp {
				t.Errorf("StopAnalysis() called with operation %v, want %v", calledOperation, expectedOp)
			}
		})
	}
}

// TestOpenInBrowser tests browser opening
func TestOpenInBrowser(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "successful open",
			url:     "https://test.cyverse.run",
			wantErr: false,
		},
		{
			name:    "browser open fails",
			url:     "https://fail.test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBrowser := &mockBrowserOpener{
				openFunc: func(url string) error {
					if url == "https://fail.test" {
						return errors.New("failed to open browser")
					}
					return nil
				},
			}

			workflows := NewFormationWorkflows(&mockFormationClient{}, mockBrowser, 5*time.Second)
			err := workflows.OpenInBrowser(tt.url)

			if tt.wantErr && err == nil {
				t.Errorf("OpenInBrowser() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("OpenInBrowser() unexpected error = %v", err)
			}

			if mockBrowser.lastURL != tt.url {
				t.Errorf("OpenInBrowser() opened %v, want %v", mockBrowser.lastURL, tt.url)
			}
		})
	}
}

// TestBrowseDataWithFormat tests data browsing with formatting
func TestBrowseDataWithFormat(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		isDirectory bool
		mockData    interface{}
		wantErr     bool
	}{
		{
			name:        "directory browse",
			path:        "/cyverse/home/test",
			isDirectory: true,
			mockData: &client.DirectoryContents{
				Path: "/cyverse/home/test",
				Type: "collection",
				Contents: []client.DirectoryEntry{
					{Name: "file1.txt", Type: "data_object"},
				},
			},
			wantErr: false,
		},
		{
			name:        "file read",
			path:        "/cyverse/home/test/file.txt",
			isDirectory: false,
			mockData: &client.FileContent{
				Path:    "/cyverse/home/test/file.txt",
				Content: "file content",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockFormationClient{
				browseDataFunc: func(ctx context.Context, path string, offset, limit int, includeMetadata bool) (interface{}, error) {
					return tt.mockData, nil
				},
			}

			workflows := NewFormationWorkflows(mockClient, &mockBrowserOpener{}, 5*time.Second)
			result, isDir, err := workflows.BrowseDataWithFormat(context.Background(), tt.path, 0, 0, false)

			if tt.wantErr {
				if err == nil {
					t.Errorf("BrowseDataWithFormat() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("BrowseDataWithFormat() unexpected error = %v", err)
				}
				if isDir != tt.isDirectory {
					t.Errorf("BrowseDataWithFormat() isDir = %v, want %v", isDir, tt.isDirectory)
				}
				if result == nil {
					t.Errorf("BrowseDataWithFormat() returned nil result")
				}
			}
		})
	}
}

// TestGetRunningAnalyses tests fetching running analyses
func TestGetRunningAnalyses(t *testing.T) {
	expectedAnalyses := []client.Analysis{
		{AnalysisID: "analysis-1", Status: "Running"},
		{AnalysisID: "analysis-2", Status: "Running"},
	}

	mockClient := &mockFormationClient{
		listAnalysesFunc: func(ctx context.Context, status string) ([]client.Analysis, error) {
			if status != "Running" {
				t.Errorf("Expected status=Running, got %v", status)
			}
			return expectedAnalyses, nil
		},
	}

	workflows := NewFormationWorkflows(mockClient, &mockBrowserOpener{}, 5*time.Second)
	analyses, err := workflows.GetRunningAnalyses(context.Background())

	if err != nil {
		t.Errorf("GetRunningAnalyses() unexpected error = %v", err)
	}
	if len(analyses) != len(expectedAnalyses) {
		t.Errorf("GetRunningAnalyses() got %d analyses, want %d", len(analyses), len(expectedAnalyses))
	}
}

// TestSystemBrowserOpener tests the real browser opener (smoke test)
func TestSystemBrowserOpener(t *testing.T) {
	// We can't really test opening a browser in CI, but we can test that the method exists
	// and doesn't panic with a valid URL. We won't actually open the browser.
	t.Skip("Skipping browser opener test - requires interactive environment")

	// The following code is not executed due to Skip, but verifies the interface exists
	_ = &SystemBrowserOpener{}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
