package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chibuka/95-cli/client"
)

var (
	ErrMissingContentLength = errors.New("missing content-length")
	ErrServerTimeout        = errors.New("server timeout")
	ErrConnectionFailed     = errors.New("connection failed")
)

func (h *httpServerRunner) sendRequest(req client.HttpRequest) (*client.HttpResponse, error) {
	url := fmt.Sprintf("http://localhost:%d%s", h.port, req.Path)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequestWithContext(ctx, req.Method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range req.Headers {
		httpReq.Header.Set(k, v)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}
	resp, err := httpClient.Do(httpReq)

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, ErrServerTimeout
		}
		return nil, fmt.Errorf("could not connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.ContentLength == -1 {
		// -1 indicates the header was missing or invalid (chunked)
		return nil, ErrMissingContentLength
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf(
				"server took too long to send response body (timeout after 30s)",
			)
		}
		return nil, fmt.Errorf("could not read server response: %w", err)
	}

	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &client.HttpResponse{
		StatusCode: resp.StatusCode,
		Body:       string(body),
		Headers:    headers,
	}, nil
}
