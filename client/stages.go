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
	StageName string `json:"stageName"`
	TestType  string `json:"testType"` // "cli_interactive" or "http_server"
	Tests     []Test `json:"tests"`
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
	HttpRequests []HttpRequest `json:"httpRequests,omitempty"`
	// Setup and cleanup operations
	Setup   *TestSetup   `json:"setup,omitempty"`
	Cleanup *TestCleanup `json:"cleanup,omitempty"`
	// Note: assertions are stripped by backend
}

// TestSetup defines operations to perform before running a test
type TestSetup struct {
	CreateDirs  []string          `json:"createDirs,omitempty"`
	CreateFiles map[string]string `json:"createFiles,omitempty"` // map of path -> content
	DeleteFiles []string          `json:"deleteFiles,omitempty"` // files to delete before test (ensure clean slate)
}

// TestCleanup defines operations to perform after running a test
type TestCleanup struct {
	DeleteDirs  []string `json:"deleteDirs,omitempty"`
	DeleteFiles []string `json:"deleteFiles,omitempty"`
}

type TestResult struct {
	TestName      string         `json:"testName"`
	ExitCode      int            `json:"exitCode"`
	Stdout        string         `json:"stdout"`
	Stderr        string         `json:"stderr"`
	HttpResponses []HttpResponse `json:"httpResponses,omitempty"`
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
	TargetStageNumber *int         `json:"targetStageNumber,omitempty"`
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

	// First attempt with current access token
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AccessToken))
	req.Header.Set("X-User-Id", strconv.Itoa(cfg.UserId))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tests: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// If 401, try to refresh token and retry
	if res.StatusCode == 401 {
		fmt.Println("Access token expired. Refreshing...")

		if cfg.RefreshToken == "" {
			return nil, fmt.Errorf("authentication failed - no refresh token available\n\n→ Run '95 login' to sign in again")
		}

		// Attempt to refresh the token
		authResponse, err := RefreshToken(cfg.RefreshToken, apiURL)
		if err != nil {
			return nil, fmt.Errorf("token refresh failed - %w\n\n→ Run '95 login' to re-authenticate", err)
		}

		// Update config with new tokens
		cfg.AccessToken = authResponse.AccessToken
		cfg.RefreshToken = authResponse.RefreshToken

		// Save updated config
		if err := cfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to save refreshed tokens: %w", err)
		}

		fmt.Println("✓ Token refreshed successfully!")

		// Retry the request with new token
		retryReq, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create retry request: %w", err)
		}
		retryReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AccessToken))
		retryReq.Header.Set("X-User-Id", strconv.Itoa(cfg.UserId))

		retryRes, err := http.DefaultClient.Do(retryReq)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch tests after refresh: %w", err)
		}
		defer func() { _ = retryRes.Body.Close() }()

		retryBody, err := io.ReadAll(retryRes.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read retry response body: %w", err)
		}

		if retryRes.StatusCode != 200 {
			return nil, fmt.Errorf("HTTP %d - %s", retryRes.StatusCode, retryBody)
		}

		// Use retry response for final result
		body = retryBody
		res = retryRes
	}

	if res.StatusCode != http.StatusOK {
		switch res.StatusCode {
		case 401:
			return nil, fmt.Errorf("authentication failed - your session has expired\n\n→ Run '95 login' to sign in again")
		case 404:
			return nil, fmt.Errorf("stage '%s' not found\n\n→ Check the UUID and try again", stageUuid)
		case 403:
			return nil, fmt.Errorf("access denied - you don't have permission to access this stage")
		default:
			errMsg := string(body)
			if errMsg == "" {
				errMsg = "unknown error"
			}
			return nil, fmt.Errorf("HTTP %d - %s", res.StatusCode, errMsg)
		}
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
// Returns only the tests for the requested stage (not prerequisites)
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

	// First attempt with current access token
	req, err := http.NewRequest("POST", validateURL, bytes.NewReader(submissionData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AccessToken))
	req.Header.Set("X-User-Id", strconv.Itoa(cfg.UserId))
	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to submit results: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// If 401, try to refresh token and retry
	if res.StatusCode == 401 {
		fmt.Println("Access token expired. Refreshing...")

		if cfg.RefreshToken == "" {
			return nil, fmt.Errorf("authentication failed - no refresh token available\n\n→ Run '95 login' to sign in again")
		}

		// Attempt to refresh the token
		authResponse, err := RefreshToken(cfg.RefreshToken, apiURL)
		if err != nil {
			return nil, fmt.Errorf("token refresh failed - %w\n\n→ Run '95 login' to re-authenticate", err)
		}

		// Update config with new tokens
		cfg.AccessToken = authResponse.AccessToken
		cfg.RefreshToken = authResponse.RefreshToken

		// Save updated config
		if err := cfg.Save(); err != nil {
			return nil, fmt.Errorf("failed to save refreshed tokens: %w", err)
		}

		fmt.Println("✓ Token refreshed successfully!")

		// Retry the submission with new token
		retryReq, err := http.NewRequest("POST", validateURL, bytes.NewReader(submissionData))
		if err != nil {
			return nil, fmt.Errorf("failed to create retry request: %w", err)
		}
		retryReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.AccessToken))
		retryReq.Header.Set("X-User-Id", strconv.Itoa(cfg.UserId))
		retryReq.Header.Set("Content-Type", "application/json")

		retryRes, err := http.DefaultClient.Do(retryReq)
		if err != nil {
			return nil, fmt.Errorf("failed to submit results after refresh: %w", err)
		}
		defer func() { _ = retryRes.Body.Close() }()

		retryBody, err := io.ReadAll(retryRes.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read retry response body: %w", err)
		}

		if retryRes.StatusCode != 200 {
			return nil, fmt.Errorf("HTTP %d - %s", retryRes.StatusCode, retryBody)
		}

		// Use retry response for final result
		body = retryBody
	} else if res.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d - %s", res.StatusCode, body)
	}

	var submissionResult SubmissionResult
	if err := json.Unmarshal(body, &submissionResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal submission result: %w", err)
	}

	return &submissionResult, nil
}
