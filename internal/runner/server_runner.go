package runner

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/chibuka/95-cli/client"
)

type httpServerRunner struct {
	cmd    *exec.Cmd
	port   int
	config *client.ServerConfig
}

func (h *httpServerRunner) startServer(programConfig *client.ProgramConfig, runCommand string) error {
	// Parse the run command
	splitCmd := strings.Fields(runCommand)
	if len(splitCmd) == 0 {
		return fmt.Errorf("run command is empty")
	}

	// Build command with program config args
	args := splitCmd[1:]
	args = append(args, programConfig.Args...)

	h.cmd = exec.Command(splitCmd[0], args...)

	// Set environment variables
	env := os.Environ()
	for k, v := range programConfig.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	h.cmd.Env = env
	h.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Capture output for debugging
	var stdout, stderr bytes.Buffer
	h.cmd.Stdout = &stdout
	h.cmd.Stderr = &stderr

	if err := h.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Check if process is still running
	if h.cmd.Process == nil {
		return fmt.Errorf("server process exited immediately")
	}

	h.port = h.config.Port

	// Wait for server to be ready
	return h.waitForServer()
}

func (h *httpServerRunner) waitForServer() error {
	deadline := time.Now().Add(time.Duration(h.config.StartupWaitMs) * time.Millisecond)
	url := fmt.Sprintf("http://localhost:%d", h.port)

	attempt := 1
	for time.Now().Before(deadline) {

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			cancel()
			return fmt.Errorf("failed to create health check request: %w", err)
		}

		resp, err := http.DefaultClient.Do(req)
		cancel()

		if err == nil {
			resp.Body.Close()
			return nil
		}

		time.Sleep(200 * time.Millisecond)
		attempt++
	}

	return fmt.Errorf("server did not start within %dms", h.config.StartupWaitMs)
}

func (h *httpServerRunner) stopServer() {
	if h.cmd == nil || h.cmd.Process == nil {
		return
	}

	// Try graceful shutdown
	h.cmd.Process.Signal(os.Interrupt)

	// Wait a bit, then force kill if still running
	done := make(chan error)
	go func() {
		_, err := h.cmd.Process.Wait()
		done <- err
	}()

	select {
	case <-done:
		// Process exited
	case <-time.After(2 * time.Second):
		// Force kill
		syscall.Kill(-h.cmd.Process.Pid, syscall.SIGKILL)
	}
}
