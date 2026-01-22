package response

import (
	"encoding/json"
	"net/http"

	"kyugo.dev/kyugo/v1/validation"
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
}

type ErrorEnvelope struct {
	Status string    `json:"status"`
	Code   int       `json:"code"`
	Error  ErrorBody `json:"error"`
}

type SuccessEnvelope struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

func Success(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(SuccessEnvelope{Status: "success", Data: data})
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
	if vdd, ok := details.([]validation.FieldError); ok {
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
			case validation.FieldError:
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
func Error(w http.ResponseWriter, code int, _type, _code, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	eb := ErrorBody{Code: _code, Message: message, Type: _type}
	if d := convertDetails(details); len(d) > 0 {
		eb.Fields = d
	}

	env := ErrorEnvelope{Status: "error", Code: code, Error: eb}
	_ = json.NewEncoder(w).Encode(env)
}
