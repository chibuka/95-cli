package runner

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/chibuka/95-cli/client"
)

func RunTest(runCommand string, stdin string, timeoutSeconds int) (*client.TestResult, error) {
	splitRunCmd := strings.Fields(runCommand)
	if len(splitRunCmd) == 0 {
		return nil, fmt.Errorf("run command is empty")
	}
	cmd, args := splitRunCmd[0], splitRunCmd[1:]

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	execCmd := exec.CommandContext(ctx, cmd, args...)

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	execCmd.Stdout = &stdoutBuffer
	execCmd.Stderr = &stderrBuffer

	stdinPipe, err := execCmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// this is non-blocking
	err = execCmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// sending test input
	_, err = stdinPipe.Write([]byte(stdin))
	if err != nil {
		return nil, fmt.Errorf("failed to write stdin: %w", err)
	}

	err = stdinPipe.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close stdin pipe: %w", err)
	}

	err = execCmd.Wait()

	exitCode := 0
	if err != nil {
		// check if it's a non-zero exit code (expected)
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Some other error (timeout, command not found, etc.)
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}

	// TestName is empty because the caller will fill it in
	return &client.TestResult{
		ExitCode: exitCode,
		Stdout:   stdoutBuffer.String(),
		Stderr:   stderrBuffer.String(),
	}, nil
}
