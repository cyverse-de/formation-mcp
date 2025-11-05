package client

// App represents an application in the Formation system.
type App struct {
	ID                 string `json:"id"`
	SystemID           string `json:"system_id"`
	Name               string `json:"name"`
	Description        string `json:"description"`
	IntegratorUsername string `json:"integrator_username,omitempty"`
}

// AppListResponse represents the response from listing apps.
type AppListResponse struct {
	Apps []App `json:"apps"`
}

// Parameter represents an app parameter.
type Parameter struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Label        string      `json:"label"`
	Description  string      `json:"description"`
	Required     bool        `json:"required"`
	Type         string      `json:"type"`
	DefaultValue interface{} `json:"default_value,omitempty"`
}

// ParameterGroup represents a group of parameters.
type ParameterGroup struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Label      string      `json:"label"`
	Parameters []Parameter `json:"parameters"`
}

// AppParameters represents the full parameter structure for an app.
type AppParameters struct {
	OverallJobType string           `json:"overall_job_type"`
	Groups         []ParameterGroup `json:"groups"`
}

// LaunchConfig represents the configuration parameters for an app.
type LaunchConfig map[string]interface{}

// LaunchSubmission represents the complete submission for launching an app.
// Matches the request body structure expected by POST /app/launch/{system_id}/{app_id}
type LaunchSubmission struct {
	Name         string                 `json:"name,omitempty"`
	Email        string                 `json:"email,omitempty"`
	Config       LaunchConfig           `json:"config"`
	SystemID     string                 `json:"system_id,omitempty"`
	Debug        bool                   `json:"debug,omitempty"`
	Notify       bool                   `json:"notify,omitempty"`
	OutputDir    string                 `json:"output_dir,omitempty"`
	Requirements map[string]interface{} `json:"requirements,omitempty"`
}

// LaunchResponse represents the response from launching an app.
// Matches the response from POST /app/launch/{system_id}/{app_id}
type LaunchResponse struct {
	AnalysisID string `json:"analysis_id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	URL        string `json:"url,omitempty"`
}

// AnalysisStatus represents the status of an analysis.
// Matches the response from GET /apps/analyses/{analysis_id}/status
type AnalysisStatus struct {
	AnalysisID      string                 `json:"analysis_id"`
	Status          string                 `json:"status"`
	URLReady        bool                   `json:"url_ready"`
	URL             string                 `json:"url,omitempty"`
	URLCheckDetails map[string]interface{} `json:"url_check_details,omitempty"`
}

// Analysis represents a running analysis.
// Matches the items in GET /apps/analyses/ response
type Analysis struct {
	AnalysisID string `json:"analysis_id"`
	AppID      string `json:"app_id"`
	SystemID   string `json:"system_id"`
	Status     string `json:"status"`
}

// AnalysisListResponse represents the response from listing analyses.
type AnalysisListResponse struct {
	Analyses []Analysis `json:"analyses"`
}

// ControlAnalysisRequest represents a request to control an analysis.
type ControlAnalysisRequest struct {
	Operation   string `json:"operation"`
	SaveOutputs bool   `json:"save_outputs,omitempty"`
}

// DirectoryEntry represents a file or directory in iRODS.
// Matches the items in the "contents" array from GET /data/{path}
type DirectoryEntry struct {
	Name string `json:"name"`
	Type string `json:"type"` // "data_object" or "collection"
}

// DirectoryContents represents the contents of a directory.
// Matches the response from GET /data/{path} for directories
type DirectoryContents struct {
	Path     string           `json:"path"`
	Type     string           `json:"type"` // "collection"
	Contents []DirectoryEntry `json:"contents"`
}

// FileContent represents the content of a file with metadata.
type FileContent struct {
	Path     string                 `json:"path"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CreateDirectoryRequest represents a request to create a directory.
type CreateDirectoryRequest struct {
	Path     string                 `json:"path"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CreateDirectoryResponse represents the response from creating a directory.
type CreateDirectoryResponse struct {
	Path string `json:"path"`
	Type string `json:"type"`
}

// SetMetadataRequest represents a request to set metadata on a path.
type SetMetadataRequest struct {
	Path     string                 `json:"path"`
	Metadata map[string]interface{} `json:"metadata"`
	Replace  bool                   `json:"replace"`
}

// DeleteRequest represents a request to delete a path.
type DeleteRequest struct {
	Path    string `json:"path"`
	Recurse bool   `json:"recurse"`
	DryRun  bool   `json:"dry_run"`
}

// LoginRequest represents a login request.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the response from a login request.
// This matches the Keycloak token response format.
type LoginResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
}
