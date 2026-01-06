// cmd/helpers.go
package cmd

import (
	"fmt"
	"strings"

	"github.com/chibuka/95-cli/client"
)

// getTestInput returns the input to display for a test
func getTestInput(test client.Test) string {
	if len(test.HttpRequests) > 0 {
		// For HTTP tests, show the first request
		req := test.HttpRequests[0]
		return fmt.Sprintf("%s %s", req.Method, req.Path)
	}
	return test.Stdin
}

// formatTestOutput formats the test result for display
func formatTestOutput(testType string, result *client.TestResult) string {
	if testType == "http_server" && len(result.HttpResponses) > 0 {
		var output strings.Builder
		for i, resp := range result.HttpResponses {
			if i > 0 {
				output.WriteString("\n---\n")
			}
			output.WriteString(fmt.Sprintf("HTTP %d\n", resp.StatusCode))
			for k, v := range resp.Headers {
				output.WriteString(fmt.Sprintf("%s: %s\n", k, v))
			}
			if resp.Body != "" {
				output.WriteString("\n")
				output.WriteString(resp.Body)
			}
		}
		return output.String()
	}
	return result.Stdout
}
