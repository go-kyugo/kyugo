package request

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"kyugo.dev/kyugo/v1/router"
)

// Request is a small wrapper around *http.Request providing convenience
// methods used by handlers in the codebase.
type Request struct {
	R *http.Request
}

// New wraps an *http.Request.
func New(r *http.Request) *Request {
	return &Request{R: r}
}

// Param returns a URL parameter value by name.
func (r *Request) Param(name string) string {
	if r == nil || r.R == nil {
		return ""
	}
	return chi.URLParam(r.R, name)
}

// BodyAsRequest is a generic helper that attempts to retrieve the validated
// body previously stored by the router's validation step and assert it to T.
func BodyAsRequest[T any](r *Request) (T, bool) {
	var zero T
	if r == nil || r.R == nil {
		return zero, false
	}
	return router.BodyAs[T](r.R)
}

// Message looks up an application message (localization) for the request.
func (r *Request) Message(key string) (string, bool) {
	return router.Message(r.R, key)
}

// BindJSON decodes the request body into v. Returns an error if decoding
// fails or the request is nil.
func (r *Request) BindJSON(v any) error {
	if r == nil || r.R == nil {
		return errors.New("nil request")
	}
	dec := json.NewDecoder(r.R.Body)
	return dec.Decode(v)
}

// Method returns the HTTP method.
func (r *Request) Method() string {
	if r == nil || r.R == nil {
		return ""
	}
	return r.R.Method
}

// Path returns the request URL path.
func (r *Request) Path() string {
	if r == nil || r.R == nil {
		return ""
	}
	return r.R.URL.Path
}

// RemoteAddr returns the client's remote address.
func (r *Request) RemoteAddr() string {
	if r == nil || r.R == nil {
		return ""
	}
	return r.R.RemoteAddr
}
