package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/chibuka/95-cli/internal/config"
)

type TestConfig struct {
	StageName     string         `json:"stageName"`
	TestType      string         `json:"testType"` // "cli_interactive" or "http_server"
	ProgramConfig *ProgramConfig `json:"programConfig"`
	ServerConfig  *ServerConfig  `json:"serverConfig"`
	Tests         []Test         `json:"tests"`
}

// ProgramConfig defines how to run the user's program for HTTP tests
type ProgramConfig struct {
	Executable string            `json:"executable"`
	Args       []string          `json:"args"`
	Env        map[string]string `json:"env"`
}

// ServerConfig defines the server parameters for HTTP tests
type ServerConfig struct {
	Port          int `json:"port"`
	StartupWaitMs int `json:"startupWaitMs"`
}

// CascadedTestConfig represents test configurations for multiple stages
// When requesting stage X, this contains tests for stages 1..X
type CascadedTestConfig struct {
	TargetStageUuid   string          `json:"targetStageUuid"`
	TargetStageNumber int             `json:"targetStageNumber"`
	StagesToRun       []StageTestInfo `json:"stagesToRun"`
}

type StageTestInfo struct {
	StageUuid   string `json:"stageUuid"`
	StageNumber int    `json:"stageNumber"`
	StageName   string `json:"stageName"`
	TestConfig  string `json:"testConfig"` // JSON string
}

type Test struct {
	TestName       string `json:"testName"`
	Stdin          string `json:"stdin"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
	// in case testType is "http_server"
	HttpRequests []HttpRequest `json:"httpRequests"`
	// Setup and cleanup operations
	Setup   *TestSetup   `json:"setup"`
	Cleanup *TestCleanup `json:"cleanup"`
	// Note: assertions are stripped by backend
}

// Define a new struct for the file items
type FileCreation struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// TestSetup defines operations to perform before running a test
type TestSetup struct {
	CreateDirs []string `json:"createDirs"`
	// Change this from map[string]string to []FileCreation
	CreateFiles []FileCreation `json:"createFiles"`
	DeleteFiles []string       `json:"deleteFiles"`
	DeleteDirs  []string       `json:"deleteDirs"`
}

// TestCleanup defines operations to perform after running a test
type TestCleanup struct {
	DeleteDirs  []string `json:"deleteDirs"`
	DeleteFiles []string `json:"deleteFiles"`
}

type TestResult struct {
	TestName      string         `json:"testName"`
	ExitCode      int            `json:"exitCode"`
	Stdout        string         `json:"stdout"`
	Stderr        string         `json:"stderr"`
	HttpResponses []HttpResponse `json:"httpResponses"`
}

type HttpResponse struct {
	StatusCode int               `json:"statusCode"`
	Body       string            `json:"body"`
	Headers    map[string]string `json:"headers"`
}

type HttpRequest struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type SubmissionRequest struct {
	StageUuid         string       `json:"stageUuid"`
	Language          string       `json:"language"`
	TestResults       []TestResult `json:"testResults"`
	TargetStageNumber *int         `json:"targetStageNumber"`
}

type SubmissionResult struct {
	Passed       bool          `json:"passed"`
	TotalTests   int           `json:"totalTests"`
	PassedTests  int           `json:"passedTests"`
	FailedTests  int           `json:"failedTests"`
	Feedback     string        `json:"feedback"`
	TestFailures []TestFailure `json:"testFailures"`
}

type TestFailure struct {
	TestName string `json:"testName"`
	Reason   string `json:"reason"`
}

// FetchCascadedTests fetches test configurations for all prerequisite stages
// When requesting stage X, this returns tests for stages 1..X to ensure backward compatibility
func FetchCascadedTests(stageUuid string, cfg *config.Config) (*CascadedTestConfig, error) {
	apiURL := cfg.GetAPIURL()
	endpoint := fmt.Sprintf("%s/api/stages/%s/tests", apiURL, stageUuid)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	body, err := sendRequest(req, cfg)
	if err != nil {
		if httpErr, ok := err.(*HttpError); ok {
			switch httpErr.StatusCode {
			case http.StatusUnauthorized:
				return nil, fmt.Errorf("authentication failed - your session has expired\n\n→ Run '95 login' to sign in again")
			case http.StatusNotFound:
				return nil, fmt.Errorf("stage '%s' not found\n\n→ Check the UUID and try again", stageUuid)
			case http.StatusForbidden:
				return nil, fmt.Errorf("access denied - you don't have permission to access this stage")
			default:
				return nil, fmt.Errorf("HTTP %d - %s", httpErr.StatusCode, httpErr.Body)
			}
		}
		return nil, err
	}

	var cascadedConfig CascadedTestConfig
	if err := json.Unmarshal(body, &cascadedConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cascaded test config: %w", err)
	}

	return &cascadedConfig, nil
}

// ParseStageTests parses the test config JSON string for a single stage
func ParseStageTests(stageInfo StageTestInfo) (*TestConfig, error) {
	var testConfig TestConfig
	if err := json.Unmarshal([]byte(stageInfo.TestConfig), &testConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal stage %d test config: %w", stageInfo.StageNumber, err)
	}
	return &testConfig, nil
}

// FetchTests is kept for backward compatibility but now uses the cascaded endpoint
// Returns only the tests for the requested stage
func FetchTests(stageUuid string, cfg *config.Config) (*TestConfig, error) {
	cascaded, err := FetchCascadedTests(stageUuid, cfg)
	if err != nil {
		return nil, err
	}

	// Return tests for the target stage only (last in the list)
	if len(cascaded.StagesToRun) == 0 {
		return nil, fmt.Errorf("no stages found in cascaded config")
	}

	targetStage := cascaded.StagesToRun[len(cascaded.StagesToRun)-1]
	return ParseStageTests(targetStage)
}

func SubmitResults(stageUuid string, language string, cfg *config.Config, results []TestResult, targetStageNumber *int) (*SubmissionResult, error) {
	apiURL := cfg.GetAPIURL()
	submissionReq := SubmissionRequest{
		StageUuid:         stageUuid,
		Language:          language,
		TestResults:       results,
		TargetStageNumber: targetStageNumber,
	}
	submissionData, err := json.Marshal(submissionReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal submission request: %w", err)
	}

	validateURL := fmt.Sprintf("%s/api/stages/validate", apiURL)

	req, err := http.NewRequest("POST", validateURL, bytes.NewReader(submissionData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Set GetBody to allow retries
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(submissionData)), nil
	}

	body, err := sendRequest(req, cfg)
	if err != nil {
		if httpErr, ok := err.(*HttpError); ok {
			return nil, fmt.Errorf("HTTP %d - %s", httpErr.StatusCode, httpErr.Body)
		}
		return nil, err
	}

	var submissionResult SubmissionResult
	if err := json.Unmarshal(body, &submissionResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal submission result: %w", err)
	}

	return &submissionResult, nil
}

type HttpError struct {
	StatusCode int
	Body       string
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
}

func sendRequest(req *http.Request, cfg *config.Config) ([]byte, error) {
	// Attempt the request
	body, statusCode, err := doRequest(req, cfg.AccessToken, cfg.UserId)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusUnauthorized {
		if statusCode != http.StatusOK {
			return nil, &HttpError{StatusCode: statusCode, Body: string(body)}
		}
		return body, nil
	}

	fmt.Println("Access token expired. Refreshing...")

	if err := performTokenRefresh(cfg); err != nil {
		return nil, err
	}

	// Retry the request with new token
	body, statusCode, err = doRequest(req, cfg.AccessToken, cfg.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to send retry request: %w", err)
	}

	if statusCode != http.StatusOK {
		return nil, &HttpError{StatusCode: statusCode, Body: string(body)}
	}

	return body, nil
}

// Handles the mechanics of checking, calling API, and saving config
func performTokenRefresh(cfg *config.Config) error {
	if cfg.RefreshToken == "" {
		return fmt.Errorf("authentication failed - no refresh token available\n\n→ Run '95 login' to sign in again")
	}

	authResponse, err := RefreshToken(cfg.RefreshToken, cfg.GetAPIURL())
	if err != nil {
		return fmt.Errorf("token refresh failed - %w\n\n→ Run '95 login' to re-authenticate", err)
	}

	cfg.AccessToken = authResponse.AccessToken
	cfg.RefreshToken = authResponse.RefreshToken

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save refreshed tokens: %w", err)
	}

	fmt.Println("✓ Token refreshed successfully!")
	return nil
}

func doRequest(req *http.Request, token string, userId int) ([]byte, int, error) {
	// Rewind body if it was read previously (for retries!)
	if req.GetBody != nil {
		bodyCopy, err := req.GetBody()
		if err != nil {
			return nil, 0, err
		}
		req.Body = bodyCopy
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("X-User-Id", strconv.Itoa(userId))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	return body, res.StatusCode, err
}
