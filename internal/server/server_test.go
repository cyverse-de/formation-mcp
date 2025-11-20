package server

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/cyverse-de/formation-mcp/internal/client"
	"github.com/cyverse-de/formation-mcp/internal/workflows"
	"github.com/mark3labs/mcp-go/mcp"
)

// mockWorkflows implements workflows for testing
type mockWorkflows struct {
	launchAndWaitFunc      func(ctx context.Context, appID, systemID, name string, config client.LaunchConfig, maxWait time.Duration) (*workflows.LaunchResult, error)
	getRunningAnalysesFunc func(ctx context.Context) ([]client.Analysis, error)
	stopAnalysisFunc       func(ctx context.Context, analysisID string, saveOutputs bool) error
	openInBrowserFunc      func(url string) error
	browseDataWithFormatFunc func(ctx context.Context, path string, offset, limit int, includeMetadata bool) (interface{}, bool, error)
}

func (m *mockWorkflows) LaunchAndWait(ctx context.Context, appID, systemID, name string, config client.LaunchConfig, maxWait time.Duration) (*workflows.LaunchResult, error) {
	if m.launchAndWaitFunc != nil {
		return m.launchAndWaitFunc(ctx, appID, systemID, name, config, maxWait)
	}
	return &workflows.LaunchResult{}, nil
}

func (m *mockWorkflows) GetRunningAnalyses(ctx context.Context) ([]client.Analysis, error) {
	if m.getRunningAnalysesFunc != nil {
		return m.getRunningAnalysesFunc(ctx)
	}
	return []client.Analysis{}, nil
}

func (m *mockWorkflows) StopAnalysis(ctx context.Context, analysisID string, saveOutputs bool) error {
	if m.stopAnalysisFunc != nil {
		return m.stopAnalysisFunc(ctx, analysisID, saveOutputs)
	}
	return nil
}

func (m *mockWorkflows) OpenInBrowser(url string) error {
	if m.openInBrowserFunc != nil {
		return m.openInBrowserFunc(url)
	}
	return nil
}

func (m *mockWorkflows) BrowseDataWithFormat(ctx context.Context, path string, offset, limit int, includeMetadata bool) (interface{}, bool, error) {
	if m.browseDataWithFormatFunc != nil {
		return m.browseDataWithFormatFunc(ctx, path, offset, limit, includeMetadata)
	}
	return nil, false, nil
}

// mockClient implements FormationAPIClient for testing
type mockClient struct {
	listAppsFunc         func(ctx context.Context, name, integrator, description, jobType string, limit, offset int) ([]client.App, error)
	getAppParametersFunc func(ctx context.Context, systemID, appID string) (*client.AppParameters, error)
	getAnalysisStatusFunc func(ctx context.Context, analysisID string) (*client.AnalysisStatus, error)
	listAnalysesFunc     func(ctx context.Context, status string) ([]client.Analysis, error)
	createDirectoryFunc  func(ctx context.Context, path string, metadata map[string]interface{}) (*client.CreateDirectoryResponse, error)
	uploadFileFunc       func(ctx context.Context, path, content string, metadata map[string]interface{}) error
	setMetadataFunc      func(ctx context.Context, path string, metadata map[string]interface{}, replace bool) error
	deleteDataFunc       func(ctx context.Context, path string, recurse, dryRun bool) error
}

func (m *mockClient) Login(ctx context.Context) error { return nil }

func (m *mockClient) ListApps(ctx context.Context, name, integrator, description, jobType string, limit, offset int) ([]client.App, error) {
	if m.listAppsFunc != nil {
		return m.listAppsFunc(ctx, name, integrator, description, jobType, limit, offset)
	}
	return []client.App{}, nil
}

func (m *mockClient) GetAppParameters(ctx context.Context, systemID, appID string) (*client.AppParameters, error) {
	if m.getAppParametersFunc != nil {
		return m.getAppParametersFunc(ctx, systemID, appID)
	}
	return &client.AppParameters{}, nil
}

func (m *mockClient) LaunchApp(ctx context.Context, systemID, appID string, submission client.LaunchSubmission) (*client.LaunchResponse, error) {
	return &client.LaunchResponse{}, nil
}

func (m *mockClient) GetAnalysisStatus(ctx context.Context, analysisID string) (*client.AnalysisStatus, error) {
	if m.getAnalysisStatusFunc != nil {
		return m.getAnalysisStatusFunc(ctx, analysisID)
	}
	return &client.AnalysisStatus{}, nil
}

func (m *mockClient) ListAnalyses(ctx context.Context, status string) ([]client.Analysis, error) {
	if m.listAnalysesFunc != nil {
		return m.listAnalysesFunc(ctx, status)
	}
	return []client.Analysis{}, nil
}

func (m *mockClient) ControlAnalysis(ctx context.Context, analysisID, operation string, saveOutputs bool) error {
	return nil
}

func (m *mockClient) BrowseData(ctx context.Context, path string, offset, limit int, includeMetadata bool) (interface{}, error) {
	return nil, nil
}

func (m *mockClient) CreateDirectory(ctx context.Context, path string, metadata map[string]interface{}) (*client.CreateDirectoryResponse, error) {
	if m.createDirectoryFunc != nil {
		return m.createDirectoryFunc(ctx, path, metadata)
	}
	return &client.CreateDirectoryResponse{}, nil
}

func (m *mockClient) UploadFile(ctx context.Context, path, content string, metadata map[string]interface{}) error {
	if m.uploadFileFunc != nil {
		return m.uploadFileFunc(ctx, path, content, metadata)
	}
	return nil
}

func (m *mockClient) SetMetadata(ctx context.Context, path string, metadata map[string]interface{}, replace bool) error {
	if m.setMetadataFunc != nil {
		return m.setMetadataFunc(ctx, path, metadata, replace)
	}
	return nil
}

func (m *mockClient) DeleteData(ctx context.Context, path string, recurse, dryRun bool) error {
	if m.deleteDataFunc != nil {
		return m.deleteDataFunc(ctx, path, recurse, dryRun)
	}
	return nil
}

// TestHandleListApps tests the list_apps handler
func TestHandleListApps(t *testing.T) {
	mockApps := []client.App{
		{
			ID:                 "app-1",
			SystemID:           "de",
			Name:               "Test App 1",
			Description:        "Test description",
			IntegratorUsername: "testuser",
		},
		{
			ID:          "app-2",
			SystemID:    "de",
			Name:        "Test App 2",
			Description: "Another test",
		},
	}

	mockClientImpl := &mockClient{
		listAppsFunc: func(ctx context.Context, name, integrator, description, jobType string, limit, offset int) ([]client.App, error) {
			return mockApps, nil
		},
	}

	mockWorkflowsImpl := &mockWorkflows{}
	server := NewFormationMCPServer(mockWorkflowsImpl, mockClientImpl)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_apps",
			Arguments: map[string]interface{}{
				"limit":  10,
				"offset": 0,
			},
		},
	}

	result, err := server.handleListApps(context.Background(), request)
	if err != nil {
		t.Errorf("handleListApps() unexpected error = %v", err)
	}

	if result == nil {
		t.Fatal("handleListApps() returned nil result")
	}

	// Verify result contains expected app information
	if len(result.Content) == 0 {
		t.Error("handleListApps() returned empty content")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("handleListApps() result is not text content")
	}

	content := textContent.Text
	if !strings.Contains(content, "Test App 1") {
		t.Error("handleListApps() result doesn't contain app name")
	}
	if !strings.Contains(content, "app-1") {
		t.Error("handleListApps() result doesn't contain app ID")
	}
}

// TestHandleGetAppParameters tests the get_app_parameters handler
func TestHandleGetAppParameters(t *testing.T) {
	mockParams := &client.AppParameters{
		OverallJobType: "Interactive",
		Groups: []client.ParameterGroup{
			{
				ID:    "group-1",
				Name:  "Input",
				Label: "Input Parameters",
				Parameters: []client.Parameter{
					{
						ID:          "param1",
						Name:        "input_file",
						Label:       "Input File",
						Description: "The input file",
						Required:    true,
						Type:        "FileInput",
					},
				},
			},
		},
	}

	mockClientImpl := &mockClient{
		getAppParametersFunc: func(ctx context.Context, systemID, appID string) (*client.AppParameters, error) {
			return mockParams, nil
		},
	}

	mockWorkflowsImpl := &mockWorkflows{}
	server := NewFormationMCPServer(mockWorkflowsImpl, mockClientImpl)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_app_parameters",
			Arguments: map[string]interface{}{
				"app_id":    "test-app",
				"system_id": "de",
			},
		},
	}

	result, err := server.handleGetAppParameters(context.Background(), request)
	if err != nil {
		t.Errorf("handleGetAppParameters() unexpected error = %v", err)
	}

	if result == nil {
		t.Fatal("handleGetAppParameters() returned nil result")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("handleGetAppParameters() result is not text content")
	}

	content := textContent.Text
	if !strings.Contains(content, "Interactive") {
		t.Error("handleGetAppParameters() result doesn't contain job type")
	}
	if !strings.Contains(content, "Input File") {
		t.Error("handleGetAppParameters() result doesn't contain parameter label")
	}
}

// TestHandleLaunchAppAndWait tests the launch_app_and_wait handler
func TestHandleLaunchAppAndWait(t *testing.T) {
	tests := []struct {
		name              string
		launchResult      *workflows.LaunchResult
		expectSuccess     bool
		expectMissingParams bool
	}{
		{
			name: "successful launch",
			launchResult: &workflows.LaunchResult{
				AnalysisID:    "analysis-123",
				Name:          "test-analysis",
				Status:        "Running",
				IsInteractive: true,
				URL:           "https://test.cyverse.run",
			},
			expectSuccess: true,
		},
		{
			name: "missing parameters",
			launchResult: &workflows.LaunchResult{
				MissingParams: []string{"param1", "param2"},
			},
			expectMissingParams: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWorkflowsImpl := &mockWorkflows{
				launchAndWaitFunc: func(ctx context.Context, appID, systemID, name string, config client.LaunchConfig, maxWait time.Duration) (*workflows.LaunchResult, error) {
					return tt.launchResult, nil
				},
			}

			server := NewFormationMCPServer(mockWorkflowsImpl, &mockClient{})

			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "launch_app_and_wait",
					Arguments: map[string]interface{}{
						"app_id":    "test-app",
						"system_id": "de",
						"name":      "test-analysis",
						"config":    map[string]interface{}{},
						"max_wait":  300,
					},
				},
			}

			result, err := server.handleLaunchAppAndWait(context.Background(), request)
			if err != nil {
				t.Errorf("handleLaunchAppAndWait() unexpected error = %v", err)
			}

			if result == nil {
				t.Fatal("handleLaunchAppAndWait() returned nil result")
			}

			textContent, ok := result.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatal("handleLaunchAppAndWait() result is not text content")
			}

			content := textContent.Text

			if tt.expectSuccess {
				if !strings.Contains(content, "analysis-123") {
					t.Error("handleLaunchAppAndWait() result doesn't contain analysis ID")
				}
				if !strings.Contains(content, "✅") {
					t.Error("handleLaunchAppAndWait() result doesn't contain success indicator")
				}
			}

			if tt.expectMissingParams {
				if !strings.Contains(content, "Missing Required Parameters") {
					t.Error("handleLaunchAppAndWait() result doesn't indicate missing params")
				}
			}
		})
	}
}

// TestHandleGetAnalysisStatus tests the get_analysis_status handler
func TestHandleGetAnalysisStatus(t *testing.T) {
	mockStatus := &client.AnalysisStatus{
		AnalysisID: "analysis-123",
		Status:     "Running",
		URLReady:   true,
		URL:        "https://test.cyverse.run",
	}

	mockClientImpl := &mockClient{
		getAnalysisStatusFunc: func(ctx context.Context, analysisID string) (*client.AnalysisStatus, error) {
			return mockStatus, nil
		},
	}

	server := NewFormationMCPServer(&mockWorkflows{}, mockClientImpl)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "get_analysis_status",
			Arguments: map[string]interface{}{
				"analysis_id": "analysis-123",
			},
		},
	}

	result, err := server.handleGetAnalysisStatus(context.Background(), request)
	if err != nil {
		t.Errorf("handleGetAnalysisStatus() unexpected error = %v", err)
	}

	if result == nil {
		t.Fatal("handleGetAnalysisStatus() returned nil result")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("handleGetAnalysisStatus() result is not text content")
	}

	content := textContent.Text
	if !strings.Contains(content, "analysis-123") {
		t.Error("handleGetAnalysisStatus() result doesn't contain analysis ID")
	}
	if !strings.Contains(content, "Running") {
		t.Error("handleGetAnalysisStatus() result doesn't contain status")
	}
}

// TestHandleListRunningAnalyses tests the list_running_analyses handler
func TestHandleListRunningAnalyses(t *testing.T) {
	mockAnalyses := []client.Analysis{
		{
			AnalysisID: "analysis-1",
			AppID:      "app-1",
			SystemID:   "de",
			Status:     "Running",
		},
		{
			AnalysisID: "analysis-2",
			AppID:      "app-2",
			SystemID:   "de",
			Status:     "Running",
		},
	}

	mockClientImpl := &mockClient{
		listAnalysesFunc: func(ctx context.Context, status string) ([]client.Analysis, error) {
			return mockAnalyses, nil
		},
	}

	server := NewFormationMCPServer(&mockWorkflows{}, mockClientImpl)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "list_running_analyses",
			Arguments: map[string]interface{}{
				"status": "Running",
			},
		},
	}

	result, err := server.handleListRunningAnalyses(context.Background(), request)
	if err != nil {
		t.Errorf("handleListRunningAnalyses() unexpected error = %v", err)
	}

	if result == nil {
		t.Fatal("handleListRunningAnalyses() returned nil result")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("handleListRunningAnalyses() result is not text content")
	}

	content := textContent.Text
	if !strings.Contains(content, "analysis-1") {
		t.Error("handleListRunningAnalyses() result doesn't contain first analysis")
	}
	if !strings.Contains(content, "analysis-2") {
		t.Error("handleListRunningAnalyses() result doesn't contain second analysis")
	}
}

// TestHandleStopAnalysis tests the stop_analysis handler
func TestHandleStopAnalysis(t *testing.T) {
	var capturedAnalysisID string
	var capturedSaveOutputs bool

	mockWorkflowsImpl := &mockWorkflows{
		stopAnalysisFunc: func(ctx context.Context, analysisID string, saveOutputs bool) error {
			capturedAnalysisID = analysisID
			capturedSaveOutputs = saveOutputs
			return nil
		},
	}

	server := NewFormationMCPServer(mockWorkflowsImpl, &mockClient{})

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "stop_analysis",
			Arguments: map[string]interface{}{
				"analysis_id":  "analysis-123",
				"save_outputs": true,
			},
		},
	}

	result, err := server.handleStopAnalysis(context.Background(), request)
	if err != nil {
		t.Errorf("handleStopAnalysis() unexpected error = %v", err)
	}

	if result == nil {
		t.Fatal("handleStopAnalysis() returned nil result")
	}

	if capturedAnalysisID != "analysis-123" {
		t.Errorf("handleStopAnalysis() called with analysis_id = %v, want analysis-123", capturedAnalysisID)
	}
	if !capturedSaveOutputs {
		t.Error("handleStopAnalysis() called with save_outputs = false, want true")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("handleStopAnalysis() result is not text content")
	}

	content := textContent.Text
	if !strings.Contains(content, "✅") {
		t.Error("handleStopAnalysis() result doesn't contain success indicator")
	}
}

// TestHandleOpenInBrowser tests the open_in_browser handler
func TestHandleOpenInBrowser(t *testing.T) {
	var capturedURL string

	mockWorkflowsImpl := &mockWorkflows{
		openInBrowserFunc: func(url string) error {
			capturedURL = url
			return nil
		},
	}

	server := NewFormationMCPServer(mockWorkflowsImpl, &mockClient{})

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "open_in_browser",
			Arguments: map[string]interface{}{
				"url": "https://test.cyverse.run",
			},
		},
	}

	result, err := server.handleOpenInBrowser(context.Background(), request)
	if err != nil {
		t.Errorf("handleOpenInBrowser() unexpected error = %v", err)
	}

	if result == nil {
		t.Fatal("handleOpenInBrowser() returned nil result")
	}

	if capturedURL != "https://test.cyverse.run" {
		t.Errorf("handleOpenInBrowser() called with url = %v, want https://test.cyverse.run", capturedURL)
	}
}

// TestHandleBrowseData tests the browse_data handler
func TestHandleBrowseData(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		mockData    interface{}
		isDirectory bool
	}{
		{
			name: "browse directory",
			path: "/cyverse/home/test",
			mockData: &client.DirectoryContents{
				Path: "/cyverse/home/test",
				Type: "collection",
				Contents: []client.DirectoryEntry{
					{Name: "file1.txt", Type: "data_object"},
					{Name: "subdir", Type: "collection"},
				},
			},
			isDirectory: true,
		},
		{
			name: "read file",
			path: "/cyverse/home/test/file.txt",
			mockData: &client.FileContent{
				Path:    "/cyverse/home/test/file.txt",
				Content: "file content here",
			},
			isDirectory: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWorkflowsImpl := &mockWorkflows{
				browseDataWithFormatFunc: func(ctx context.Context, path string, offset, limit int, includeMetadata bool) (interface{}, bool, error) {
					return tt.mockData, tt.isDirectory, nil
				},
			}

			server := NewFormationMCPServer(mockWorkflowsImpl, &mockClient{})

			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "browse_data",
					Arguments: map[string]interface{}{
						"path":             tt.path,
						"offset":           0,
						"limit":            100,
						"include_metadata": false,
					},
				},
			}

			result, err := server.handleBrowseData(context.Background(), request)
			if err != nil {
				t.Errorf("handleBrowseData() unexpected error = %v", err)
			}

			if result == nil {
				t.Fatal("handleBrowseData() returned nil result")
			}

			textContent, ok := result.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatal("handleBrowseData() result is not text content")
			}

			content := textContent.Text
			if !strings.Contains(content, tt.path) {
				t.Errorf("handleBrowseData() result doesn't contain path %v", tt.path)
			}

			if tt.isDirectory {
				if !strings.Contains(content, "file1.txt") {
					t.Error("handleBrowseData() directory result doesn't contain file entry")
				}
			} else {
				if !strings.Contains(content, "file content here") {
					t.Error("handleBrowseData() file result doesn't contain content")
				}
			}
		})
	}
}

// TestHandleCreateDirectory tests the create_directory handler
func TestHandleCreateDirectory(t *testing.T) {
	mockClientImpl := &mockClient{
		createDirectoryFunc: func(ctx context.Context, path string, metadata map[string]interface{}) (*client.CreateDirectoryResponse, error) {
			return &client.CreateDirectoryResponse{
				Path: path,
				Type: "collection",
			}, nil
		},
	}

	server := NewFormationMCPServer(&mockWorkflows{}, mockClientImpl)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "create_directory",
			Arguments: map[string]interface{}{
				"path":     "/cyverse/home/test/newdir",
				"metadata": map[string]interface{}{"project": "test"},
			},
		},
	}

	result, err := server.handleCreateDirectory(context.Background(), request)
	if err != nil {
		t.Errorf("handleCreateDirectory() unexpected error = %v", err)
	}

	if result == nil {
		t.Fatal("handleCreateDirectory() returned nil result")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("handleCreateDirectory() result is not text content")
	}

	content := textContent.Text
	if !strings.Contains(content, "✅") {
		t.Error("handleCreateDirectory() result doesn't contain success indicator")
	}
	if !strings.Contains(content, "/cyverse/home/test/newdir") {
		t.Error("handleCreateDirectory() result doesn't contain path")
	}
}

// TestHandleUploadFile tests the upload_file handler
func TestHandleUploadFile(t *testing.T) {
	var capturedPath string
	var capturedContent string

	mockClientImpl := &mockClient{
		uploadFileFunc: func(ctx context.Context, path, content string, metadata map[string]interface{}) error {
			capturedPath = path
			capturedContent = content
			return nil
		},
	}

	server := NewFormationMCPServer(&mockWorkflows{}, mockClientImpl)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "upload_file",
			Arguments: map[string]interface{}{
				"path":    "/cyverse/home/test/file.txt",
				"content": "test file content",
			},
		},
	}

	result, err := server.handleUploadFile(context.Background(), request)
	if err != nil {
		t.Errorf("handleUploadFile() unexpected error = %v", err)
	}

	if result == nil {
		t.Fatal("handleUploadFile() returned nil result")
	}

	if capturedPath != "/cyverse/home/test/file.txt" {
		t.Errorf("handleUploadFile() called with path = %v", capturedPath)
	}
	if capturedContent != "test file content" {
		t.Errorf("handleUploadFile() called with content = %v", capturedContent)
	}
}

// TestHandleSetMetadata tests the set_metadata handler
func TestHandleSetMetadata(t *testing.T) {
	mockClientImpl := &mockClient{
		setMetadataFunc: func(ctx context.Context, path string, metadata map[string]interface{}, replace bool) error {
			return nil
		},
	}

	server := NewFormationMCPServer(&mockWorkflows{}, mockClientImpl)

	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name: "set_metadata",
			Arguments: map[string]interface{}{
				"path":     "/cyverse/home/test/file.txt",
				"metadata": map[string]interface{}{"key": "value"},
				"replace":  true,
			},
		},
	}

	result, err := server.handleSetMetadata(context.Background(), request)
	if err != nil {
		t.Errorf("handleSetMetadata() unexpected error = %v", err)
	}

	if result == nil {
		t.Fatal("handleSetMetadata() returned nil result")
	}

	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatal("handleSetMetadata() result is not text content")
	}

	content := textContent.Text
	if !strings.Contains(content, "✅") {
		t.Error("handleSetMetadata() result doesn't contain success indicator")
	}
}

// TestHandleDeleteData tests the delete_data handler
func TestHandleDeleteData(t *testing.T) {
	tests := []struct {
		name    string
		dryRun  bool
		recurse bool
	}{
		{
			name:    "actual delete",
			dryRun:  false,
			recurse: false,
		},
		{
			name:    "dry run",
			dryRun:  true,
			recurse: false,
		},
		{
			name:    "recursive delete",
			dryRun:  false,
			recurse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedDryRun, capturedRecurse bool

			mockClientImpl := &mockClient{
				deleteDataFunc: func(ctx context.Context, path string, recurse, dryRun bool) error {
					capturedDryRun = dryRun
					capturedRecurse = recurse
					return nil
				},
			}

			server := NewFormationMCPServer(&mockWorkflows{}, mockClientImpl)

			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name: "delete_data",
					Arguments: map[string]interface{}{
						"path":    "/cyverse/home/test/file.txt",
						"dry_run": tt.dryRun,
						"recurse": tt.recurse,
					},
				},
			}

			result, err := server.handleDeleteData(context.Background(), request)
			if err != nil {
				t.Errorf("handleDeleteData() unexpected error = %v", err)
			}

			if result == nil {
				t.Fatal("handleDeleteData() returned nil result")
			}

			if capturedDryRun != tt.dryRun {
				t.Errorf("handleDeleteData() dry_run = %v, want %v", capturedDryRun, tt.dryRun)
			}
			if capturedRecurse != tt.recurse {
				t.Errorf("handleDeleteData() recurse = %v, want %v", capturedRecurse, tt.recurse)
			}

			textContent, ok := result.Content[0].(mcp.TextContent)
			if !ok {
				t.Fatal("handleDeleteData() result is not text content")
			}

			content := textContent.Text
			if !strings.Contains(content, "✅") {
				t.Error("handleDeleteData() result doesn't contain success indicator")
			}

			if tt.dryRun && !strings.Contains(content, "Dry run") {
				t.Error("handleDeleteData() dry run result doesn't indicate dry run")
			}
		})
	}
}

// TestUnmarshalParams tests parameter unmarshaling
func TestUnmarshalParams(t *testing.T) {
	type testParams struct {
		Name  string `json:"name"`
		Limit int    `json:"limit"`
	}

	tests := []struct {
		name      string
		arguments map[string]interface{}
		wantErr   bool
	}{
		{
			name: "valid params",
			arguments: map[string]interface{}{
				"name":  "test",
				"limit": 10,
			},
			wantErr: false,
		},
		{
			name: "empty params",
			arguments: map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.arguments,
				},
			}

			var params testParams
			err := unmarshalParams(request, &params)

			if tt.wantErr && err == nil {
				t.Error("unmarshalParams() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unmarshalParams() unexpected error = %v", err)
			}

			if !tt.wantErr && tt.arguments["name"] != nil {
				if params.Name != tt.arguments["name"].(string) {
					t.Errorf("unmarshalParams() name = %v, want %v", params.Name, tt.arguments["name"])
				}
			}
		})
	}
}

// TestNewFormationMCPServer tests server creation
func TestNewFormationMCPServer(t *testing.T) {
	mockWorkflowsImpl := &mockWorkflows{}
	mockClientImpl := &mockClient{}

	server := NewFormationMCPServer(mockWorkflowsImpl, mockClientImpl)

	if server == nil {
		t.Fatal("NewFormationMCPServer() returned nil")
	}

	if server.server == nil {
		t.Error("NewFormationMCPServer() MCP server is nil")
	}

	if server.workflows == nil {
		t.Error("NewFormationMCPServer() workflows is nil")
	}

	if server.client == nil {
		t.Error("NewFormationMCPServer() client is nil")
	}
}

// TestToolRegistration verifies all tools are registered
func TestToolRegistration(t *testing.T) {
	mockWorkflowsImpl := &mockWorkflows{}
	mockClientImpl := &mockClient{}

	server := NewFormationMCPServer(mockWorkflowsImpl, mockClientImpl)

	expectedTools := []string{
		"list_apps",
		"get_app_parameters",
		"launch_app_and_wait",
		"get_analysis_status",
		"list_running_analyses",
		"stop_analysis",
		"open_in_browser",
		"browse_data",
		"create_directory",
		"upload_file",
		"set_metadata",
		"delete_data",
	}

	// We can't directly access the tools from the MCP server,
	// but we can verify the tool definitions are created correctly
	for _, toolName := range expectedTools {
		var tool mcp.Tool

		switch toolName {
		case "list_apps":
			tool = server.listAppsTool()
		case "get_app_parameters":
			tool = server.getAppParametersTool()
		case "launch_app_and_wait":
			tool = server.launchAppAndWaitTool()
		case "get_analysis_status":
			tool = server.getAnalysisStatusTool()
		case "list_running_analyses":
			tool = server.listRunningAnalysesTool()
		case "stop_analysis":
			tool = server.stopAnalysisTool()
		case "open_in_browser":
			tool = server.openInBrowserTool()
		case "browse_data":
			tool = server.browseDataTool()
		case "create_directory":
			tool = server.createDirectoryTool()
		case "upload_file":
			tool = server.uploadFileTool()
		case "set_metadata":
			tool = server.setMetadataTool()
		case "delete_data":
			tool = server.deleteDataTool()
		}

		if tool.Name != toolName {
			t.Errorf("Tool %v not properly defined", toolName)
		}

		if tool.Description == "" {
			t.Errorf("Tool %v has no description", toolName)
		}
	}
}

// TestToolSchemaValidation verifies tool input schemas
func TestToolSchemaValidation(t *testing.T) {
	server := NewFormationMCPServer(&mockWorkflows{}, &mockClient{})

	// Test list_apps schema
	listAppsTool := server.listAppsTool()
	props := listAppsTool.InputSchema.Properties
	if props == nil {
		t.Error("list_apps has no properties defined")
	}

	// Test launch_app_and_wait schema
	launchTool := server.launchAppAndWaitTool()
	if launchTool.InputSchema.Required == nil || len(launchTool.InputSchema.Required) == 0 {
		t.Error("launch_app_and_wait has no required parameters")
	}

	appIDFound := false
	for _, req := range launchTool.InputSchema.Required {
		if req == "app_id" {
			appIDFound = true
			break
		}
	}
	if !appIDFound {
		t.Error("launch_app_and_wait doesn't require app_id parameter")
	}
}

// Helper to convert interface to JSON and back
func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
