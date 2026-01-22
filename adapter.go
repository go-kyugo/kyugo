package kyugo

import (
	"net/http"
)

// Adapt converts a handler function that accepts our wrapper types into a
// standard http.HandlerFunc. The input `h` is expected to be a bound
// method value or function with signature func(*response.Response, *request.Request).
func Adapt(h func(*Response, *Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := NewResponse(w, r)
		req := NewRequest(r)
		h(resp, req)
	}
}
