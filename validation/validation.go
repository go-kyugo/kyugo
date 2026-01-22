package validation

import (
	"reflect"
	"strings"

	v10 "github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

var validate *v10.Validate

func init() {
	validate = v10.New()
}

// Validate runs struct validation using go-playground/validator.
func Validate(v interface{}) error {
	if validate == nil {
		validate = v10.New()
	}
	return validate.Struct(v)
}

// FormatValidationErrors converts validator.ValidationErrors into a slice of FieldError
// where Code follows the pattern "invalid_<rule>|<param>" when applicable.
func FormatValidationErrors(err error, v interface{}) []FieldError {
	if err == nil {
		return nil
	}
	ve, ok := err.(v10.ValidationErrors)
	if !ok {
		return []FieldError{{Field: "", Code: "invalid", Message: err.Error()}}
	}
	out := make([]FieldError, 0, len(ve))

	// reflect on provided example value to map struct field -> json tag
	var rt reflect.Type
	if v != nil {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if rv.IsValid() {
			rt = rv.Type()
		}
	}

	for _, f := range ve {
		fieldName := f.Field()
		// try to resolve json tag name
		if rt != nil {
			if sf, found := rt.FieldByName(fieldName); found {
				tag := sf.Tag.Get("json")
				if tag != "" {
					parts := strings.Split(tag, ",")
					if parts[0] != "" && parts[0] != "-" {
						fieldName = parts[0]
					}
				} else {
					fieldName = strings.ToLower(fieldName)
				}
			} else {
				fieldName = strings.ToLower(fieldName)
			}
		}

		rule := strings.ToLower(f.Tag())
		param := f.Param()
		// emit code in the form the router expects: "INVALID_<RULE>|<param>"
		code := "INVALID_" + strings.ToUpper(rule)
		if param != "" {
			code = code + "|" + param
		}

		out = append(out, FieldError{Field: strings.ToLower(fieldName), Code: code, Message: f.Error()})
	}
	return out
}
