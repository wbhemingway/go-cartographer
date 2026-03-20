package cartographer

import (
	"fmt"
	"io"
	"net/http"
)

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("cartographer api error: status %d: %s", e.StatusCode, e.Message)
}

func parseAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    string(body),
	}
}
