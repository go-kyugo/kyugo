package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"

	"kyugo.dev/kyugo/v1/registry"
	"kyugo.dev/kyugo/v1/response"
	"kyugo.dev/kyugo/v1/validation"
)

type routeEntry struct {
	method      string
	path        string
	handlerName string
}

var (
	mu     sync.RWMutex
	routes []routeEntry
	// validateBodyMap maps route keys to an optional DTO type to validate.
	validateBodyMu  sync.RWMutex
	validateBodyMap = make(map[string]reflect.Type)
	// middlewareMap maps route keys to a slice of middleware to apply.
	middlewareMu  sync.RWMutex
	middlewareMap = make(map[string][]func(http.Handler) http.Handler)
)

type ctxKey string

const validatedBodyKey ctxKey = "youu.validated_body"

// ContextKey is a key type exported for storing values in request context
// used by the router (for example messages injected by the server).
type ContextKey string

// MessagesKey is the context key where the server will store flattened
// language messages for the current application language.
var MessagesKey ContextKey = "youu.messages"

// RegisterHandlerName keeps compatibility with generated code that registers
// handlers by name and resolves them from the runtime registry.
func RegisterHandlerName(method, p, handlerName string) {
	mu.Lock()
	routes = append(routes, routeEntry{method: method, path: p, handlerName: handlerName})
	mu.Unlock()
}

// Router is a lightweight wrapper around an underlying chi router that
// exposes a small, fluent API similar to the example you provided.
type Router struct {
	r chi.Router
}

// New creates a new Router instance.
func New() *Router {
	return &Router{r: chi.NewRouter()}
}

// Handler returns the underlying http.Handler to be used with ListenAndServe.
func (rt *Router) Handler() http.Handler {
	// Register any routes previously added via RegisterHandlerName
	mu.RLock()
	defer mu.RUnlock()

	for _, re := range routes {
		method := strings.ToUpper(re.method)
		p := re.path
		name := re.handlerName

		if strings.HasPrefix(name, "missing:") {
			rt.r.Method(method, p, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				http.NotFound(w, &http.Request{})
			}))
			continue
		}

		rt.r.Method(method, p, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if h := registry.Get(name); h != nil {
				h(w, req)
				return
			}
			http.NotFound(w, req)
		}))
	}

	return rt.r
}

// Group creates a route group rooted at the provided prefix.
func (rt *Router) Group(prefix string) *Group {
	return &Group{parent: rt.r, prefix: prefix}
}

// Group represents a group of routes under a common prefix.
type Group struct {
	parent chi.Router
	prefix string
}

// With returns a new Group that applies the provided middleware to all
// routes registered through it. This mirrors chi's `With` behaviour and
// allows `router.Group("/x").With(mw).Get(...)` usage.
func (g *Group) With(mws ...func(http.Handler) http.Handler) *Group {
	return &Group{parent: g.parent.With(mws...), prefix: g.prefix}
}

// Use applies middleware to the group's parent router in-place and returns
// the same group for chaining. This mirrors chi's `Use` behaviour.
func (g *Group) Use(mws ...func(http.Handler) http.Handler) *Group {
	g.parent.Use(mws...)
	return g
}

// Middleware is an alias for Use to allow a consistent chainable name
// between groups and routes: `group.Middleware(mw...)`.
func (g *Group) Middleware(mws ...func(http.Handler) http.Handler) *Group {
	return g.Use(mws...)
}

func join(prefix, p string) string {
	if prefix == "" || prefix == "/" {
		return p
	}
	return path.Join(prefix, p)
}

// RouteChain provides chainable validation methods after registering a route.
// They are no-ops in this lightweight implementation but preserve the fluent API.
type RouteChain struct {
	key string
}

// ValidateQuery is a no-op that preserves the example fluent API.
func (rc *RouteChain) ValidateQuery(_ interface{}) *RouteChain { return rc }

// ValidateBody registers an optional DTO value for the previously registered
// route. If `dto` is nil we only check that the request body is valid JSON.
// If `dto` is a non-nil example value, the router will attempt to unmarshal
// the body into a fresh instance of that type and run `validation.Validate`.
func (rc *RouteChain) ValidateBody(dto interface{}) *RouteChain {
	if rc == nil || rc.key == "" {
		return rc
	}
	var t reflect.Type
	if dto != nil {
		t = reflect.TypeOf(dto)
	}
	validateBodyMu.Lock()
	validateBodyMap[rc.key] = t
	validateBodyMu.Unlock()
	return rc
}

// Middleware registers middleware for the previously-registered route.
// Middleware added here will wrap the route handler and any validation
// performed by the router. Call it like:
//
//	group.Post(...).ValidateBody(...).Middleware(mw1, mw2)
func (rc *RouteChain) Middleware(mws ...func(http.Handler) http.Handler) *RouteChain {
	if rc == nil || rc.key == "" {
		return rc
	}
	middlewareMu.Lock()
	middlewareMap[rc.key] = append(middlewareMap[rc.key], mws...)
	middlewareMu.Unlock()
	return rc
}

// helper to register and return a RouteChain
func registerAndChain(parent chi.Router, method, p string, h http.HandlerFunc) *RouteChain {
	// convert {name:regex} -> {name} so chi matches the route
	re := regexp.MustCompile(`\{([a-zA-Z0-9_]+):[^}]+\}`)
	cleaned := re.ReplaceAllString(p, `{$1}`)

	key := strings.ToUpper(method) + " " + cleaned

	parent.Method(strings.ToUpper(method), cleaned, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// baseHandler performs body validation (if configured) and then
		// invokes the actual handler `h`.
		baseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// check if this route has a body validation rule
			validateBodyMu.RLock()
			t, ok := validateBodyMap[key]
			validateBodyMu.RUnlock()
			if ok {
				// read entire body and restore later so handler can read it too
				b, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Failed to read body", http.StatusInternalServerError)
					return
				}
				// syntax check JSON
				var tmp interface{}
				if len(b) == 0 {
					// empty body is invalid JSON for endpoints expecting a body
					msg, ok := Message(r, "locale.invalid_body")
					if !ok {
						msg = "Invalid JSON body"
					}
					response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "INVALID_BODY", msg, nil)
					return
				}
				if err := json.Unmarshal(b, &tmp); err != nil {
					msg, ok := Message(r, "locale.invalid_body")
					if !ok {
						msg = "Invalid JSON body"
					}
					response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "INVALID_BODY", msg, nil)
					return
				}

				// if a concrete DTO type was provided, unmarshal into it and run validation
				if t != nil {
					var v interface{}
					if t.Kind() == reflect.Ptr {
						v = reflect.New(t.Elem()).Interface()
					} else {
						v = reflect.New(t).Interface()
					}
					if err := json.Unmarshal(b, v); err != nil {
						msg, ok := Message(r, "locale.invalid_body")
						if !ok {
							msg = "Invalid JSON body"
						}
						response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "INVALID_BODY", msg, nil)
						return
					}
					if err := validation.Validate(v); err != nil {
						fields := validation.FormatValidationErrors(err, v)

						// localize each field message using resources with precedence:
						// 1. fields.<field>.<rule>
						// 2. rules.<rule>
						// 3. fields.<field>
						for i := range fields {
							fe := &fields[i]
							// parse rule and param from Code: "invalid_<rule>|<param>"
							// parse rule and param from Code. supported formats:
							// "INVALID_<RULE>|<param>", "invalid_<Rule>", "RULE", "RULE|param"
							code := fe.Code
							param := ""
							parts := strings.SplitN(code, "|", 2)
							main := parts[0]
							if len(parts) > 1 {
								param = parts[1]
							}
							// strip optional "INVALID_" prefix (case-insensitive)
							if strings.HasPrefix(strings.ToUpper(main), "INVALID_") {
								main = main[len("INVALID_"):]
							}
							rule := strings.ToLower(main)
							upper := strings.ToUpper(code)
							if strings.HasPrefix(upper, "INVALID_") {
								rest := code[len("INVALID_"):]
								parts := strings.SplitN(rest, "|", 2)
								rule = strings.ToLower(parts[0])
								if len(parts) > 1 {
									param = parts[1]
								}
							}
							var msgStr string
							keysToTry := []string{
								fmt.Sprintf("fields.%s.%s", strings.ToLower(fe.Field), rule),
								fmt.Sprintf("rules.%s", rule),
								fmt.Sprintf("fields.%s", strings.ToLower(fe.Field)),
							}
							for _, k := range keysToTry {
								if s, ok := Message(r, k); ok && s != "" {
									msgStr = s
									break
								}
							}
							if msgStr == "" {
								if fe.Message != "" {
									msgStr = fe.Message
								} else {
									msgStr = "Invalid value"
								}
							}
							fieldLabel := fe.Field
							if s, ok := Message(r, fmt.Sprintf("fields.%s", strings.ToLower(fe.Field))); ok && s != "" {
								fieldLabel = s
							}
							msgStr = strings.ReplaceAll(msgStr, "{field}", fieldLabel)
							msgStr = strings.ReplaceAll(msgStr, "{param}", param)
							fe.Message = msgStr
						}

						msg, ok := Message(r, "locale.validation_failed")
						if !ok || msg == "" {
							msg = "Validation failed"
						}
						response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "INVALID_ATTRIBUTES", msg, fields)
						return
					}
					// store validated value in request context for handler use
					ctx := context.WithValue(r.Context(), validatedBodyKey, v)
					r = r.WithContext(ctx)
				}

				// restore body for downstream handlers
				r.Body = io.NopCloser(bytes.NewReader(b))
			}

			h(w, r)
		})

		// fetch any middleware registered for this route and wrap the base
		// handler. Middleware should run before the validation/handler.
		middlewareMu.RLock()
		mws := middlewareMap[key]
		middlewareMu.RUnlock()

		final := http.Handler(baseHandler)
		for i := len(mws) - 1; i >= 0; i-- {
			final = mws[i](final)
		}

		final.ServeHTTP(w, r)
	}))
	return &RouteChain{key: key}
}

// Get registers a GET handler under the group's prefix.
func (g *Group) Get(p string, h http.HandlerFunc, mws ...func(http.Handler) http.Handler) *RouteChain {
	full := join(g.prefix, p)
	return registerAndChain(g.parent.With(mws...), "GET", full, h)
}

// Post registers a POST handler under the group's prefix.
func (g *Group) Post(p string, h http.HandlerFunc, mws ...func(http.Handler) http.Handler) *RouteChain {
	full := join(g.prefix, p)
	return registerAndChain(g.parent.With(mws...), "POST", full, h)
}

// Patch registers a PATCH handler under the group's prefix.
func (g *Group) Patch(p string, h http.HandlerFunc, mws ...func(http.Handler) http.Handler) *RouteChain {
	full := join(g.prefix, p)
	return registerAndChain(g.parent.With(mws...), "PATCH", full, h)
}

// Delete registers a DELETE handler under the group's prefix.
func (g *Group) Delete(p string, h http.HandlerFunc, mws ...func(http.Handler) http.Handler) *RouteChain {
	full := join(g.prefix, p)
	return registerAndChain(g.parent.With(mws...), "DELETE", full, h)
}

// Param returns a URL parameter value by name. It delegates to chi.URLParam
// but keeps handlers free from importing chi directly.
func Param(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}

// BodyAs retrieves a previously-validated request body (set by ValidateBody)
// and attempts to return it as type T. The returned bool is false when the
// validated body is not present or cannot be asserted to T.
func BodyAs[T any](r *http.Request) (T, bool) {
	var zero T
	if r == nil {
		return zero, false
	}
	v := r.Context().Value(validatedBodyKey)
	if v == nil {
		return zero, false
	}
	if vv, ok := v.(T); ok {
		return vv, true
	}
	// try pointer -> value conversion
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.Elem().IsValid() {
			if rv.Elem().CanInterface() {
				if val, ok := rv.Elem().Interface().(T); ok {
					return val, true
				}
			}
		}
	}
	return zero, false
}

// Message returns a message previously injected into the request context
// by the server. It looks up `key` in the flattened messages map.
func Message(r *http.Request, key string) (string, bool) {
	var zero string
	if r == nil {
		return zero, false
	}
	v := r.Context().Value(MessagesKey)
	if v == nil {
		return zero, false
	}
	if m, ok := v.(map[string]string); ok {
		if s, ok2 := m[key]; ok2 {
			return s, true
		}
	}
	return zero, false
}
