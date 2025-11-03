// Package client provides an HTTP client for the Formation API.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// metadataHeaderPrefix is the prefix for metadata headers
	metadataHeaderPrefix = "X-Datastore-"

	// tokenExpiryMargin is the safety margin before token expiry to trigger refresh
	tokenExpiryMargin = 60 * time.Second

	// defaultHTTPTimeout is the default timeout for HTTP requests
	defaultHTTPTimeout = 30 * time.Second
)

// FormationAPIClient defines the interface for interacting with the Formation API.
// This interface allows for mocking in tests and alternative implementations.
type FormationAPIClient interface {
	// Login authenticates with the Formation API and obtains an access token.
	Login(ctx context.Context) error

	// ListApps lists available applications with optional filtering.
	ListApps(ctx context.Context, name, integrator, description, jobType string, limit, offset int) ([]App, error)

	// GetAppParameters retrieves the parameter definitions for an app.
	GetAppParameters(ctx context.Context, systemID, appID string) (*AppParameters, error)

	// LaunchApp launches an application with the given configuration.
	LaunchApp(ctx context.Context, systemID, appID string, submission LaunchSubmission) (*LaunchResponse, error)

	// GetAnalysisStatus retrieves the status of an analysis.
	GetAnalysisStatus(ctx context.Context, analysisID string) (*AnalysisStatus, error)

	// ListAnalyses lists analyses filtered by status.
	ListAnalyses(ctx context.Context, status string) ([]Analysis, error)

	// ControlAnalysis controls an analysis (e.g., stop, pause).
	ControlAnalysis(ctx context.Context, analysisID, operation string, saveOutputs bool) error

	// BrowseData browses a directory or reads a file from iRODS.
	BrowseData(ctx context.Context, path string, offset, limit int, includeMetadata bool) (interface{}, error)

	// CreateDirectory creates a directory in iRODS.
	CreateDirectory(ctx context.Context, path string, metadata map[string]interface{}) (*CreateDirectoryResponse, error)

	// UploadFile uploads a file to iRODS.
	UploadFile(ctx context.Context, path, content string, metadata map[string]interface{}) error

	// SetMetadata sets metadata on a path in iRODS.
	SetMetadata(ctx context.Context, path string, metadata map[string]interface{}, replace bool) error

	// DeleteData deletes a file or directory from iRODS.
	DeleteData(ctx context.Context, path string, recurse, dryRun bool) error
}

// FormationClient is the HTTP client for the Formation API.
type FormationClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
	tokenExpiry time.Time
	username   string
	password   string
}

// NewFormationClient creates a new Formation API client.
func NewFormationClient(baseURL, token, username, password string) *FormationClient {
	return &FormationClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultHTTPTimeout,
		},
		token:    token,
		username: username,
		password: password,
	}
}

// Compile-time check to ensure FormationClient implements FormationAPIClient.
var _ FormationAPIClient = (*FormationClient)(nil)

// ensureToken ensures that the client has a valid token.
// If the token is expired or missing and credentials are provided, it will login.
func (c *FormationClient) ensureToken(ctx context.Context) error {
	// If we have a valid token, use it
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	// If no credentials, we can't refresh
	if c.username == "" || c.password == "" {
		if c.token == "" {
			return fmt.Errorf("no token or credentials available")
		}
		// We have a token but it might be expired - let's try using it anyway
		return nil
	}

	// Login to get a new token
	slog.Debug("token expired or missing, logging in", "username", c.username)
	return c.Login(ctx)
}

// Login authenticates with the Formation API and stores the token.
func (c *FormationClient) Login(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/login", nil)
	if err != nil {
		return fmt.Errorf("failed to create login request: %w", err)
	}

	// Use HTTP Basic Authentication
	req.SetBasicAuth(c.username, c.password)

	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	slog.Info("api_call", "method", "POST", "endpoint", "/login", "status", resp.StatusCode, "duration", duration)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return fmt.Errorf("failed to decode login response: %w", err)
	}

	c.token = loginResp.AccessToken
	// Calculate expiry time from expires_in (seconds) with safety margin
	expiresAt := time.Now().Add(time.Duration(loginResp.ExpiresIn) * time.Second)
	c.tokenExpiry = expiresAt.Add(-tokenExpiryMargin)

	slog.Info("login successful", "expires_in", loginResp.ExpiresIn, "expires_at", expiresAt, "effective_expiry", c.tokenExpiry)
	return nil
}

// doRequest performs an HTTP request with authentication and error handling.
func (c *FormationClient) doRequest(ctx context.Context, method, path string, body io.Reader, headers map[string]string) (*http.Response, error) {
	// Ensure we have a valid token
	if err := c.ensureToken(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authorization header
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	// Set additional headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	duration := time.Since(startTime)
	slog.Debug("api_call", "method", method, "path", path, "status", resp.StatusCode, "duration", duration)

	// Check for error status codes
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// buildDataPath constructs the full API path for data store operations.
// It ensures the path starts with /data/ and normalizes leading slashes.
func (c *FormationClient) buildDataPath(path string) string {
	return "/data/" + strings.TrimPrefix(path, "/")
}

// addMetadataHeaders adds metadata as X-Datastore-* headers to the headers map.
func (c *FormationClient) addMetadataHeaders(headers map[string]string, metadata map[string]interface{}) {
	for k, v := range metadata {
		headers[metadataHeaderPrefix+k] = fmt.Sprint(v)
	}
}

// doRequestAndDecode performs an HTTP request and decodes the JSON response.
// This helper reduces boilerplate for API calls that return JSON responses.
func (c *FormationClient) doRequestAndDecode(ctx context.Context, method, path string, body io.Reader, headers map[string]string, result interface{}) error {
	resp, err := c.doRequest(ctx, method, path, body, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

// ListApps lists available VICE applications.
func (c *FormationClient) ListApps(ctx context.Context, name, integrator, description, jobType string, limit, offset int) ([]App, error) {
	query := url.Values{}
	if name != "" {
		query.Set("name", name)
	}
	if integrator != "" {
		query.Set("integrator", integrator)
	}
	if description != "" {
		query.Set("description", description)
	}
	if jobType != "" {
		query.Set("job_type", jobType)
	}
	query.Set("limit", fmt.Sprintf("%d", limit))
	query.Set("offset", fmt.Sprintf("%d", offset))

	path := "/apps?" + query.Encode()
	var appResp AppListResponse
	if err := c.doRequestAndDecode(ctx, "GET", path, nil, nil, &appResp); err != nil {
		return nil, err
	}

	return appResp.Apps, nil
}

// GetAppParameters retrieves the parameters for an app.
func (c *FormationClient) GetAppParameters(ctx context.Context, systemID, appID string) (*AppParameters, error) {
	path := fmt.Sprintf("/apps/%s/%s/parameters", systemID, appID)
	var params AppParameters
	if err := c.doRequestAndDecode(ctx, "GET", path, nil, nil, &params); err != nil {
		return nil, err
	}

	return &params, nil
}

// LaunchApp launches an application.
// Submits a complete analysis submission to the Formation API.
// The Formation API will auto-generate name and output_dir if not provided.
// Email will be resolved from the JWT token if not provided.
func (c *FormationClient) LaunchApp(ctx context.Context, systemID, appID string, submission LaunchSubmission) (*LaunchResponse, error) {
	body, err := json.Marshal(submission)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal launch request: %w", err)
	}

	path := fmt.Sprintf("/app/launch/%s/%s", systemID, appID)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var launchResp LaunchResponse
	if err := c.doRequestAndDecode(ctx, "POST", path, bytes.NewReader(body), headers, &launchResp); err != nil {
		return nil, err
	}

	return &launchResp, nil
}

// GetAnalysisStatus retrieves the status of an analysis.
func (c *FormationClient) GetAnalysisStatus(ctx context.Context, analysisID string) (*AnalysisStatus, error) {
	path := fmt.Sprintf("/apps/analyses/%s/status", analysisID)
	var status AnalysisStatus
	if err := c.doRequestAndDecode(ctx, "GET", path, nil, nil, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// ListAnalyses lists analyses, optionally filtered by status.
func (c *FormationClient) ListAnalyses(ctx context.Context, status string) ([]Analysis, error) {
	query := url.Values{}
	if status != "" {
		query.Set("status", status)
	}

	path := "/apps/analyses/"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var analysisResp AnalysisListResponse
	if err := c.doRequestAndDecode(ctx, "GET", path, nil, nil, &analysisResp); err != nil {
		return nil, err
	}

	return analysisResp.Analyses, nil
}

// ControlAnalysis controls an analysis (save_and_exit, exit, extend_time).
// The operation parameter is sent as a query parameter, not in the body.
func (c *FormationClient) ControlAnalysis(ctx context.Context, analysisID, operation string, saveOutputs bool) error {
	query := url.Values{}
	query.Set("operation", operation)

	path := fmt.Sprintf("/apps/analyses/%s/control?%s", analysisID, query.Encode())

	resp, err := c.doRequest(ctx, "POST", path, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// BrowseData browses a directory or reads a file from iRODS.
func (c *FormationClient) BrowseData(ctx context.Context, path string, offset, limit int, includeMetadata bool) (interface{}, error) {
	query := url.Values{}
	if offset > 0 {
		query.Set("offset", fmt.Sprintf("%d", offset))
	}
	if limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", limit))
	}
	if includeMetadata {
		query.Set("include_metadata", "true")
	}

	fullPath := c.buildDataPath(path)
	if len(query) > 0 {
		fullPath += "?" + query.Encode()
	}

	resp, err := c.doRequest(ctx, "GET", fullPath, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	// If it's JSON, it's a directory listing
	if strings.Contains(contentType, "application/json") {
		var dirContents DirectoryContents
		if err := json.NewDecoder(resp.Body).Decode(&dirContents); err != nil {
			return nil, fmt.Errorf("failed to decode directory contents: %w", err)
		}

		// Note: Directory metadata is returned in HTTP headers (X-Datastore-*)
		// and should be extracted by the caller if needed
		return &dirContents, nil
	}

	// Otherwise, it's a file
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %w", err)
	}

	fileContent := &FileContent{
		Path:    path,
		Content: string(bodyBytes),
	}

	// Extract metadata from headers if present
	if includeMetadata {
		fileContent.Metadata = extractMetadataFromHeaders(resp.Header)
	}

	return fileContent, nil
}

// CreateDirectory creates a directory in iRODS.
// Uses resource_type=directory query parameter with no body, per Formation API.
func (c *FormationClient) CreateDirectory(ctx context.Context, path string, metadata map[string]interface{}) (*CreateDirectoryResponse, error) {
	fullPath := c.buildDataPath(path)

	// Add resource_type=directory query parameter
	query := url.Values{}
	query.Set("resource_type", "directory")
	fullPath += "?" + query.Encode()

	headers := map[string]string{}

	// Add metadata headers
	c.addMetadataHeaders(headers, metadata)

	// No body for directory creation
	resp, err := c.doRequest(ctx, "PUT", fullPath, nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var createResp CreateDirectoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return nil, fmt.Errorf("failed to decode create directory response: %w", err)
	}

	return &createResp, nil
}

// UploadFile uploads a file to iRODS.
func (c *FormationClient) UploadFile(ctx context.Context, path, content string, metadata map[string]interface{}) error {
	fullPath := c.buildDataPath(path)
	headers := map[string]string{
		"Content-Type": "application/octet-stream",
	}

	// Add metadata headers
	c.addMetadataHeaders(headers, metadata)

	resp, err := c.doRequest(ctx, "PUT", fullPath, strings.NewReader(content), headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// SetMetadata sets metadata on a path in iRODS.
func (c *FormationClient) SetMetadata(ctx context.Context, path string, metadata map[string]interface{}, replace bool) error {
	fullPath := c.buildDataPath(path)
	query := url.Values{}
	if replace {
		query.Set("replace_metadata", "true")
	}
	if len(query) > 0 {
		fullPath += "?" + query.Encode()
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// Add metadata headers
	c.addMetadataHeaders(headers, metadata)

	resp, err := c.doRequest(ctx, "PUT", fullPath, nil, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// DeleteData deletes a file or directory from iRODS.
func (c *FormationClient) DeleteData(ctx context.Context, path string, recurse, dryRun bool) error {
	fullPath := c.buildDataPath(path)
	query := url.Values{}
	if recurse {
		query.Set("recurse", "true")
	}
	if dryRun {
		query.Set("dry_run", "true")
	}
	if len(query) > 0 {
		fullPath += "?" + query.Encode()
	}

	resp, err := c.doRequest(ctx, "DELETE", fullPath, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// extractMetadataFromHeaders extracts metadata from HTTP headers with the metadata prefix.
func extractMetadataFromHeaders(headers http.Header) map[string]interface{} {
	metadata := make(map[string]interface{})
	for k, v := range headers {
		if strings.HasPrefix(k, metadataHeaderPrefix) {
			key := strings.TrimPrefix(k, metadataHeaderPrefix)
			if len(v) > 0 {
				metadata[key] = v[0]
			}
		}
	}
	return metadata
}
