package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestLogin tests the login functionality with a mock server
func TestLogin(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		serverResponse LoginResponse
		serverStatus   int
		wantErr        bool
		errContains    string
	}{
		{
			name:     "successful login",
			username: "testuser",
			password: "testpass",
			serverResponse: LoginResponse{
				AccessToken: "test-token-123",
				ExpiresIn:   3600,
				TokenType:   "Bearer",
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "invalid credentials",
			username:     "baduser",
			password:     "badpass",
			serverStatus: http.StatusUnauthorized,
			wantErr:      true,
			errContains:  "401",
		},
		{
			name:         "server error",
			username:     "testuser",
			password:     "testpass",
			serverStatus: http.StatusInternalServerError,
			wantErr:      true,
			errContains:  "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify basic auth
				username, password, ok := r.BasicAuth()
				if !ok || username != tt.username || password != tt.password {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				w.WriteHeader(tt.serverStatus)
				if tt.serverStatus == http.StatusOK {
					json.NewEncoder(w).Encode(tt.serverResponse)
				}
			}))
			defer server.Close()

			client := NewFormationClient(server.URL, "", tt.username, tt.password)
			err := client.Login(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Errorf("Login() expected error but got none")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Login() error = %v, should contain %v", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Login() unexpected error = %v", err)
				}
				if client.token != tt.serverResponse.AccessToken {
					t.Errorf("Login() token = %v, want %v", client.token, tt.serverResponse.AccessToken)
				}
			}
		})
	}
}

// TestListApps tests listing applications
func TestListApps(t *testing.T) {
	tests := []struct {
		name           string
		nameFilter     string
		limit          int
		offset         int
		serverResponse AppListResponse
		wantErr        bool
	}{
		{
			name:       "successful list",
			nameFilter: "",
			limit:      10,
			offset:     0,
			serverResponse: AppListResponse{
				Apps: []App{
					{
						ID:                 "app-1",
						SystemID:           "de",
						Name:               "Test App 1",
						Description:        "A test application",
						IntegratorUsername: "testuser",
					},
					{
						ID:                 "app-2",
						SystemID:           "de",
						Name:               "Test App 2",
						Description:        "Another test application",
						IntegratorUsername: "testuser",
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "with name filter",
			nameFilter: "Test",
			limit:      10,
			offset:     0,
			serverResponse: AppListResponse{
				Apps: []App{
					{
						ID:          "app-1",
						SystemID:    "de",
						Name:        "Test App 1",
						Description: "A test application",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify authorization header
				if r.Header.Get("Authorization") != "Bearer test-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// Verify query parameters
				query := r.URL.Query()
				if tt.nameFilter != "" && query.Get("name") != tt.nameFilter {
					t.Errorf("Expected name filter %v, got %v", tt.nameFilter, query.Get("name"))
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := NewFormationClient(server.URL, "test-token", "", "")
			apps, err := client.ListApps(context.Background(), tt.nameFilter, "", "", "", tt.limit, tt.offset)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ListApps() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ListApps() unexpected error = %v", err)
				}
				if len(apps) != len(tt.serverResponse.Apps) {
					t.Errorf("ListApps() got %d apps, want %d", len(apps), len(tt.serverResponse.Apps))
				}
			}
		})
	}
}

// TestGetAppParameters tests fetching app parameters
func TestGetAppParameters(t *testing.T) {
	expectedParams := &AppParameters{
		OverallJobType: "Interactive",
		Groups: []ParameterGroup{
			{
				ID:    "group-1",
				Name:  "Input",
				Label: "Input Parameters",
				Parameters: []Parameter{
					{
						ID:          "param-1",
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the path
		if !strings.HasPrefix(r.URL.Path, "/apps/") || !strings.HasSuffix(r.URL.Path, "/parameters") {
			t.Errorf("Unexpected path: %v", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedParams)
	}))
	defer server.Close()

	client := NewFormationClient(server.URL, "test-token", "", "")
	params, err := client.GetAppParameters(context.Background(), "de", "test-app-id")

	if err != nil {
		t.Errorf("GetAppParameters() unexpected error = %v", err)
	}
	if params.OverallJobType != expectedParams.OverallJobType {
		t.Errorf("GetAppParameters() job type = %v, want %v", params.OverallJobType, expectedParams.OverallJobType)
	}
	if len(params.Groups) != len(expectedParams.Groups) {
		t.Errorf("GetAppParameters() got %d groups, want %d", len(params.Groups), len(expectedParams.Groups))
	}
}

// TestLaunchApp tests launching an application
func TestLaunchApp(t *testing.T) {
	submission := LaunchSubmission{
		Name:   "test-analysis",
		Config: LaunchConfig{"param1": "value1"},
	}

	expectedResponse := &LaunchResponse{
		AnalysisID: "analysis-123",
		Name:       "test-analysis",
		Status:     "Submitted",
		URL:        "https://test.cyverse.run",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %v", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/app/launch/") {
			t.Errorf("Unexpected path: %v", r.URL.Path)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected application/json content type")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	client := NewFormationClient(server.URL, "test-token", "", "")
	response, err := client.LaunchApp(context.Background(), "de", "test-app-id", submission)

	if err != nil {
		t.Errorf("LaunchApp() unexpected error = %v", err)
	}
	if response.AnalysisID != expectedResponse.AnalysisID {
		t.Errorf("LaunchApp() analysis ID = %v, want %v", response.AnalysisID, expectedResponse.AnalysisID)
	}
}

// TestGetAnalysisStatus tests fetching analysis status
func TestGetAnalysisStatus(t *testing.T) {
	tests := []struct {
		name           string
		analysisID     string
		serverResponse AnalysisStatus
		wantErr        bool
	}{
		{
			name:       "running analysis",
			analysisID: "analysis-123",
			serverResponse: AnalysisStatus{
				AnalysisID: "analysis-123",
				Status:     "Running",
				URLReady:   false,
			},
			wantErr: false,
		},
		{
			name:       "ready analysis with URL",
			analysisID: "analysis-456",
			serverResponse: AnalysisStatus{
				AnalysisID: "analysis-456",
				Status:     "Running",
				URLReady:   true,
				URL:        "https://test.cyverse.run",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/apps/analyses/" + tt.analysisID + "/status"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %v, got %v", expectedPath, r.URL.Path)
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.serverResponse)
			}))
			defer server.Close()

			client := NewFormationClient(server.URL, "test-token", "", "")
			status, err := client.GetAnalysisStatus(context.Background(), tt.analysisID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetAnalysisStatus() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("GetAnalysisStatus() unexpected error = %v", err)
				}
				if status.AnalysisID != tt.serverResponse.AnalysisID {
					t.Errorf("GetAnalysisStatus() ID = %v, want %v", status.AnalysisID, tt.serverResponse.AnalysisID)
				}
				if status.URLReady != tt.serverResponse.URLReady {
					t.Errorf("GetAnalysisStatus() URLReady = %v, want %v", status.URLReady, tt.serverResponse.URLReady)
				}
			}
		})
	}
}

// TestListAnalyses tests listing analyses by status
func TestListAnalyses(t *testing.T) {
	expectedAnalyses := []Analysis{
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify status query parameter
		status := r.URL.Query().Get("status")
		if status != "Running" {
			t.Errorf("Expected status=Running, got %v", status)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(AnalysisListResponse{Analyses: expectedAnalyses})
	}))
	defer server.Close()

	client := NewFormationClient(server.URL, "test-token", "", "")
	analyses, err := client.ListAnalyses(context.Background(), "Running")

	if err != nil {
		t.Errorf("ListAnalyses() unexpected error = %v", err)
	}
	if len(analyses) != len(expectedAnalyses) {
		t.Errorf("ListAnalyses() got %d analyses, want %d", len(analyses), len(expectedAnalyses))
	}
}

// TestControlAnalysis tests controlling an analysis
func TestControlAnalysis(t *testing.T) {
	tests := []struct {
		name        string
		analysisID  string
		operation   string
		saveOutputs bool
		wantErr     bool
	}{
		{
			name:        "save and exit",
			analysisID:  "analysis-123",
			operation:   "save_and_exit",
			saveOutputs: true,
			wantErr:     false,
		},
		{
			name:        "exit without save",
			analysisID:  "analysis-456",
			operation:   "exit",
			saveOutputs: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST, got %v", r.Method)
				}

				// Verify operation query parameter
				operation := r.URL.Query().Get("operation")
				if operation != tt.operation {
					t.Errorf("Expected operation=%v, got %v", tt.operation, operation)
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewFormationClient(server.URL, "test-token", "", "")
			err := client.ControlAnalysis(context.Background(), tt.analysisID, tt.operation, tt.saveOutputs)

			if tt.wantErr && err == nil {
				t.Errorf("ControlAnalysis() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ControlAnalysis() unexpected error = %v", err)
			}
		})
	}
}

// TestBrowseData tests browsing directories and reading files
func TestBrowseData(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		isDirectory    bool
		serverResponse interface{}
		contentType    string
		wantErr        bool
	}{
		{
			name:        "browse directory",
			path:        "/cyverse/home/testuser",
			isDirectory: true,
			serverResponse: DirectoryContents{
				Path: "/cyverse/home/testuser",
				Type: "collection",
				Contents: []DirectoryEntry{
					{Name: "file1.txt", Type: "data_object"},
					{Name: "subdir", Type: "collection"},
				},
			},
			contentType: "application/json",
			wantErr:     false,
		},
		{
			name:           "read file",
			path:           "/cyverse/home/testuser/file.txt",
			isDirectory:    false,
			serverResponse: "This is file content",
			contentType:    "text/plain",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify path construction
				expectedPath := "/data" + tt.path
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %v, got %v", expectedPath, r.URL.Path)
				}

				w.Header().Set("Content-Type", tt.contentType)
				w.WriteHeader(http.StatusOK)

				if tt.isDirectory {
					json.NewEncoder(w).Encode(tt.serverResponse)
				} else {
					w.Write([]byte(tt.serverResponse.(string)))
				}
			}))
			defer server.Close()

			client := NewFormationClient(server.URL, "test-token", "", "")
			result, err := client.BrowseData(context.Background(), tt.path, 0, 0, false)

			if tt.wantErr {
				if err == nil {
					t.Errorf("BrowseData() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("BrowseData() unexpected error = %v", err)
				}

				if tt.isDirectory {
					dirContents, ok := result.(*DirectoryContents)
					if !ok {
						t.Errorf("BrowseData() expected DirectoryContents")
					} else {
						expected := tt.serverResponse.(DirectoryContents)
						if len(dirContents.Contents) != len(expected.Contents) {
							t.Errorf("BrowseData() got %d entries, want %d", len(dirContents.Contents), len(expected.Contents))
						}
					}
				} else {
					fileContent, ok := result.(*FileContent)
					if !ok {
						t.Errorf("BrowseData() expected FileContent")
					} else {
						expected := tt.serverResponse.(string)
						if fileContent.Content != expected {
							t.Errorf("BrowseData() content = %v, want %v", fileContent.Content, expected)
						}
					}
				}
			}
		})
	}
}

// TestCreateDirectory tests directory creation
func TestCreateDirectory(t *testing.T) {
	metadata := map[string]interface{}{
		"project": "test-project",
		"owner":   "testuser",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT, got %v", r.Method)
		}

		// Verify query parameter
		resourceType := r.URL.Query().Get("resource_type")
		if resourceType != "directory" {
			t.Errorf("Expected resource_type=directory, got %v", resourceType)
		}

		// Verify metadata headers
		if r.Header.Get("X-Datastore-project") != "test-project" {
			t.Errorf("Expected X-Datastore-project header")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateDirectoryResponse{
			Path: "/cyverse/home/testuser/newdir",
			Type: "collection",
		})
	}))
	defer server.Close()

	client := NewFormationClient(server.URL, "test-token", "", "")
	resp, err := client.CreateDirectory(context.Background(), "/cyverse/home/testuser/newdir", metadata)

	if err != nil {
		t.Errorf("CreateDirectory() unexpected error = %v", err)
	}
	if resp.Type != "collection" {
		t.Errorf("CreateDirectory() type = %v, want collection", resp.Type)
	}
}

// TestUploadFile tests file upload
func TestUploadFile(t *testing.T) {
	path := "/cyverse/home/testuser/test.txt"
	content := "test file content"
	metadata := map[string]interface{}{
		"description": "test file",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT, got %v", r.Method)
		}

		// Verify content type
		if r.Header.Get("Content-Type") != "application/octet-stream" {
			t.Errorf("Expected application/octet-stream content type")
		}

		// Verify metadata header
		if r.Header.Get("X-Datastore-description") != "test file" {
			t.Errorf("Expected X-Datastore-description header")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewFormationClient(server.URL, "test-token", "", "")
	err := client.UploadFile(context.Background(), path, content, metadata)

	if err != nil {
		t.Errorf("UploadFile() unexpected error = %v", err)
	}
}

// TestSetMetadata tests metadata setting
func TestSetMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT, got %v", r.Method)
		}

		// Verify replace_metadata query parameter
		replace := r.URL.Query().Get("replace_metadata")
		if replace != "true" {
			t.Errorf("Expected replace_metadata=true, got %v", replace)
		}

		// Verify metadata headers
		if r.Header.Get("X-Datastore-key1") != "value1" {
			t.Errorf("Expected X-Datastore-key1 header")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewFormationClient(server.URL, "test-token", "", "")
	err := client.SetMetadata(context.Background(), "/cyverse/home/testuser/file.txt", metadata, true)

	if err != nil {
		t.Errorf("SetMetadata() unexpected error = %v", err)
	}
}

// TestDeleteData tests data deletion
func TestDeleteData(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		recurse bool
		dryRun  bool
		wantErr bool
	}{
		{
			name:    "delete file",
			path:    "/cyverse/home/testuser/file.txt",
			recurse: false,
			dryRun:  false,
			wantErr: false,
		},
		{
			name:    "recursive delete",
			path:    "/cyverse/home/testuser/dir",
			recurse: true,
			dryRun:  false,
			wantErr: false,
		},
		{
			name:    "dry run",
			path:    "/cyverse/home/testuser/file.txt",
			recurse: false,
			dryRun:  true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE, got %v", r.Method)
				}

				// Verify query parameters
				query := r.URL.Query()
				if tt.recurse && query.Get("recurse") != "true" {
					t.Errorf("Expected recurse=true")
				}
				if tt.dryRun && query.Get("dry_run") != "true" {
					t.Errorf("Expected dry_run=true")
				}

				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewFormationClient(server.URL, "test-token", "", "")
			err := client.DeleteData(context.Background(), tt.path, tt.recurse, tt.dryRun)

			if tt.wantErr && err == nil {
				t.Errorf("DeleteData() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("DeleteData() unexpected error = %v", err)
			}
		})
	}
}

// TestTokenRefresh tests automatic token refresh
func TestTokenRefresh(t *testing.T) {
	loginCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" {
			loginCount++
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(LoginResponse{
				AccessToken: "refreshed-token",
				ExpiresIn:   1, // Very short expiry for testing
			})
			return
		}

		// Any other endpoint
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(AppListResponse{Apps: []App{}})
	}))
	defer server.Close()

	client := NewFormationClient(server.URL, "", "testuser", "testpass")

	// First call should login
	_, err := client.ListApps(context.Background(), "", "", "", "", 10, 0)
	if err != nil {
		t.Errorf("First ListApps() unexpected error = %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Second call should refresh token
	_, err = client.ListApps(context.Background(), "", "", "", "", 10, 0)
	if err != nil {
		t.Errorf("Second ListApps() unexpected error = %v", err)
	}

	if loginCount < 2 {
		t.Errorf("Expected at least 2 login calls, got %d", loginCount)
	}
}
