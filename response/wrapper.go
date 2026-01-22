package response

import (
	"net/http"
	"strconv"
)

// Response wraps http.ResponseWriter and *http.Request (kept as raw on need).
type Response struct {
	W http.ResponseWriter
	R *http.Request
}

// New creates a Response wrapper.
func New(w http.ResponseWriter, r *http.Request) *Response {
	return &Response{W: w, R: r}
}

// JSON writes a JSON response with the given status.
func (resp *Response) JSON(status int, code int, message string, v interface{}) {
	resp.W.Header().Set("Content-Type", "application/json")
	if status >= 200 && status < 300 {
		Success(resp.W, code, message, v)
		return
	}
	// For non-2xx statuses use the Error envelope with a default HTTP code/message.

	Error(resp.W, status, "HTTP_ERROR", strconv.Itoa(code), message, nil)
}

// WriteDBError inspects an error and writes an internal server error
// response if err != nil. Returns true when an error was written.
func (resp *Response) WriteDBError(err error) bool {
	if err == nil {
		return false
	}
	Error(resp.W, http.StatusInternalServerError, "internal_error", "DB_ERROR", "Database error", nil)
	return true
}
