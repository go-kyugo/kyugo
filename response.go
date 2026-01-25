package kyugo

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
)

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// ErrorBody defines the structure inside the top-level `error` key. Field
// ordering here matters for JSON output: Code, Fields (optional), Message, Type.
type ErrorBody struct {
	Type    string        `json:"type"`
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Fields  []ErrorDetail `json:"fields,omitempty"`
	Meta    interface{}   `json:"meta,omitempty"`
}

type ErrorEnvelope struct {
	Status string    `json:"status"`
	Code   int       `json:"code"`
	Error  ErrorBody `json:"error"`
}

type SuccessEnvelope struct {
	Status  string      `json:"status"`
	Code    int         `json:"code,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"`
}

type ErrorExtras struct {
	Code string
	Type string
}

type Response struct {
	W http.ResponseWriter
	R *http.Request
}

// New creates a Response wrapper.
func NewResponse(w http.ResponseWriter, r *http.Request) *Response {
	return &Response{W: w, R: r}
}

// JSON writes a JSON response with the given status.
func (resp *Response) JSON(status int, message string, v interface{}, extras ...ErrorExtras) {
	resp.W.Header().Set("Content-Type", "application/json")
	if status >= 200 && status < 300 {
		SuccessResponse(resp.W, status, message, v)
		return
	}
	// For non-2xx statuses use the Error envelope with a default HTTP code/message.

	ErrorResponse(resp.W, status, message, nil, extras...)
}

// ServeFile serves a file from disk using http.ServeFile. Helpful for
// serving static HTML or other files directly.
func (resp *Response) ServeFile(filePath string) error {
	if resp == nil || resp.W == nil {
		return fmt.Errorf("nil response")
	}

	// prefer in-memory resource when available
	if b, ok := GetResource(filePath); ok {
		ct := mime.TypeByExtension(filepath.Ext(filePath))
		if ct == "" {
			ct = http.DetectContentType(b)
		}
		resp.W.Header().Set("Content-Type", ct)
		resp.W.WriteHeader(http.StatusOK)
		_, _ = resp.W.Write(b)
		return nil
	}

	// fallback to disk; require request for ServeFile
	if resp.R == nil {
		return fmt.Errorf("nil request for ServeFile fallback")
	}
	http.ServeFile(resp.W, resp.R, filePath)
	return nil
}

// Attachment streams a file as an attachment (forces download) with the
// provided filename. It sets an appropriate Content-Type when possible.
func (resp *Response) Attachment(filePath, filename string) error {
	if resp == nil || resp.W == nil || resp.R == nil {
		return fmt.Errorf("nil response/request")
	}
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	ct := mime.TypeByExtension(filepath.Ext(filePath))
	if ct == "" {
		ct = "application/octet-stream"
	}
	resp.W.Header().Set("Content-Type", ct)
	resp.W.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	http.ServeContent(resp.W, resp.R, filename, fi.ModTime(), f)
	return nil
}

// WriteDBError inspects an error and writes an internal server error
// response if err != nil. Returns true when an error was written.
func (resp *Response) WriteDBError(err error) bool {
	if err == nil {
		return false
	}
	ErrorResponse(resp.W, http.StatusInternalServerError, "internal_error", nil, ErrorExtras{
		Code: "DB_ERROR",
		Type: "DATABASE_ERROR",
	})
	return true
}

func SuccessResponse(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(SuccessEnvelope{Status: "success", Code: code, Message: message, Data: data})
}

// convertDetails attempts to turn various detail types into []ErrorDetail.
func convertDetails(details interface{}) []ErrorDetail {
	if details == nil {
		return nil
	}
	// already in our type
	if dd, ok := details.([]ErrorDetail); ok {
		return dd
	}
	// from validation package
	if vdd, ok := details.([]FieldError); ok {
		out := make([]ErrorDetail, 0, len(vdd))
		for _, fe := range vdd {
			out = append(out, ErrorDetail{Field: fe.Field, Code: fe.Code, Message: fe.Message})
		}
		return out
	}
	// generic slice of interfaces
	if s, ok := details.([]interface{}); ok {
		out := make([]ErrorDetail, 0, len(s))
		for _, it := range s {
			switch v := it.(type) {
			case ErrorDetail:
				out = append(out, v)
			case FieldError:
				out = append(out, ErrorDetail{Field: v.Field, Code: v.Code, Message: v.Message})
			case map[string]interface{}:
				ed := ErrorDetail{}
				if f, ok := v["field"].(string); ok {
					ed.Field = f
				}
				if c, ok := v["code"].(string); ok {
					ed.Code = c
				}
				if m, ok := v["message"].(string); ok {
					ed.Message = m
				}
				out = append(out, ed)
			}
		}
		return out
	}
	return nil
}

// Error builds a consistent error envelope. When `details` are provided they
// are included under `error.fields`; otherwise that key is omitted.
func ErrorResponse(w http.ResponseWriter, code int, message string, details interface{}, extras ...ErrorExtras) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	eb := ErrorBody{Code: "", Message: message, Type: ""}
	// If at least one extras entry is provided, use its values as overrides.
	// Map ErrorExtras.Code -> ErrorBody.Type and ErrorExtras.Type -> ErrorBody.Code
	if len(extras) > 0 {
		if s := extras[0].Code; s != "" {
			eb.Type = s
		}
		if s := extras[0].Type; s != "" {
			eb.Code = s
		}
	}
	// If a second extras entry is provided, it can override the first.
	if len(extras) > 1 {
		if s := extras[1].Code; s != "" {
			eb.Type = s
		}
		if s := extras[1].Type; s != "" {
			eb.Code = s
		}
	}
	if d := convertDetails(details); len(d) > 0 {
		eb.Fields = d
	}
	// If extras are provided, interpret them as override values for
	// error `code` and `type` in that order: extras[0] -> code, extras[1] -> type.

	env := ErrorEnvelope{Status: "error", Code: code, Error: eb}
	_ = json.NewEncoder(w).Encode(env)
}
