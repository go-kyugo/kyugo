package handler

import (
	"net/http"

	"kyugo.dev/kyugo/v1/request"
	"kyugo.dev/kyugo/v1/response"
)

// Adapt converts a handler function that accepts our wrapper types into a
// standard http.HandlerFunc. The input `h` is expected to be a bound
// method value or function with signature func(*response.Response, *request.Request).
func Adapt(h func(*response.Response, *request.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := response.New(w, r)
		req := request.New(r)
		h(resp, req)
	}
}
