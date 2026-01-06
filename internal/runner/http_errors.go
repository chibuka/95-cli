package runner

import "errors"

func formatHTTPError(err error) string {
	switch {
	case errors.Is(err, ErrMissingContentLength):
		return "Your server response is missing the Content-Length header.\n→ Add 'Content-Length: 0'."

	case errors.Is(err, ErrServerTimeout):
		return "Server took too long to respond."

	case errors.Is(err, ErrConnectionFailed):
		return "Could not connect to server."

	default:
		return err.Error()
	}
}
