package runner

import (
	"fmt"

	"github.com/chibuka/95-cli/client"
)

func RunHTTPTest(programConfig *client.ProgramConfig, serverConfig *client.ServerConfig,
	runCommand string, test client.Test) (*client.TestResult, error) {

	// Check if configs are provided
	if programConfig == nil {
		return nil, fmt.Errorf("program config is required for HTTP tests")
	}
	if serverConfig == nil {
		return nil, fmt.Errorf("server config is required for HTTP tests")
	}

	// Execute setup
	if err := ExecuteSetup(test.Setup); err != nil {
		return nil, fmt.Errorf("setup failed: %w", err)
	}

	// Ensure cleanup runs even if test fails
	defer func() {
		if test.Cleanup != nil {
			if err := ExecuteCleanup(test.Cleanup); err != nil {
				fmt.Printf("Warning: cleanup failed: %v\n", err)
			}
		}
	}()

	// Start server
	runner := &httpServerRunner{
		config: serverConfig,
	}

	if err := runner.startServer(programConfig, runCommand); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}
	defer runner.stopServer()

	// Send HTTP requests and collect responses
	var responses []client.HttpResponse
	for _, req := range test.HttpRequests {
		resp, err := runner.sendRequest(req)

		if err != nil {
			// Format user-friendly error message
			reqDesc := fmt.Sprintf("%s %s", req.Method, req.Path)
			return nil, fmt.Errorf("%s\n\n  → %s", reqDesc, formatHTTPError(err))
		}
		responses = append(responses, *resp)
	}

	return &client.TestResult{
		TestName:      test.TestName,
		HttpResponses: responses,
	}, nil
}
