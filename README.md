# Kyugo — Lightweight Go API Toolkit

Kyugo is a small, opinionated toolkit to build HTTP APIs in Go. It provides:

- configuration loading utilities
- a configurable HTTP server with pluggable middlewares and services container
- a lightweight router wrapper with a fluent API and request body validation
- JSON response helpers with a standard envelope for success and error responses
- thin validation helpers built on top of go-playground/validator
- a simple runtime registry for named handlers
- logging adapters (console, color, simple, or no-op)
- optional database connection helper and support for running SQL migrations

This repository includes a working example at [examples/usage](examples/usage).

Quick start
-----------

1. Copy the example config to the example folder and edit values:

```bash
cp examples/usage/config.example.json examples/usage/config.json
```

2. Run the usage example:

```bash
go run ./examples/usage
```

Core concepts
-------------

- Configuration: load JSON config into `config.ConfigVar` using the helpers in `config`.
- Server: create a `server.Options` and call `server.New(opts)`; options accept a `Handler`, default middlewares, timeouts and an optional `Config` pointer.
- Router: use `router.New()` and controller `RegisterRoutes(router)` functions. Routes support `Group`, `Get`, `Post`, `Patch`, `Delete`, and the chainable `.ValidateBody()` / `.ValidateQuery()` helpers.
- Validation: the router can validate JSON bodies against a provided DTO type and produce localized, field-aware validation errors.
- Responses: use the response helpers to send consistent success/error envelopes across the API.
- Registry: register handlers by name with `registry.Register(name, handler)` and the router can resolve them at runtime.
- Services: register service instances in the server via `Server.RegisterService(name, instance)` and retrieve them with `Server.Service(name)`.

New helpers and wrappers
------------------------

- `request.Request`: small wrapper with helpers like `Param`, `Message` and the package-level generic helper `BodyAsRequest[T]` to fetch validated request bodies.
- `response.Response`: wrapper around `http.ResponseWriter` with helpers to standardize responses:
    - `JSON(status int, message string, v interface{}, extras ...kyugo.ErrorExtras)` — writes a success envelope for 2xx statuses or an error envelope for non-2xx. For success responses the `code` in the JSON equals the HTTP status passed. For error responses pass an optional `kyugo.ErrorExtras` to control the `error.code` and `error.type` fields.
    - `WriteDBError(err error)` — helper that writes an internal server error using the standard error envelope when `err != nil`.
- `handler.Adapt`: adapter to convert controller methods with signature `func(*response.Response, *request.Request)` into `http.HandlerFunc` for router registration.
- Logger: zerolog-backed console writer with short level codes and ANSI color support; request logger middleware emits a single console line with colored keys for fast scanning.

Configuration structure
-----------------------

Configuration is a plain JSON object matching `config.Config`. Important sections:

- `app`: `name`, `environment`, `debug`, `language`
- `server`: `host`, `port`, timeouts and a nested `cors` configuration
- `database`: `type`, `host`, `port`, `user`, `password`, `dbname`, `sslmode`

See [examples/usage/config.example.json](examples/usage/config.example.json) for a complete example.

Example usage (snippet)
-----------------------

```go
r := kyugo.NewRouter()

if err := cfg.LoadConfig("./config.json"); err != nil {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

ctrl := &controllers.Controller{}
ctrl.RegisterRoutes(r)

opts := kyugo.Options{
	Config:  &cfg.ConfigVar,
	Handler: r.Handler(),
	DefaultMiddlewares: []func(http.Handler) http.Handler{
		kyugo.CORS(cfg.ConfigVar.Server.Cors),
		kyugo.LoggerMiddleware,
	},
	ReadTimeout:  time.Duration(cfg.ConfigVar.Server.ReadTimeoutSeconds) * time.Second,
	WriteTimeout: time.Duration(cfg.ConfigVar.Server.WriteTimeoutSeconds) * time.Second,
}

srv, err := kyugo.NewServer(opts)
if err != nil {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

registerServices(srv)
ctrl.Init(srv)

srv.Start()
```

Controller signature (new style)
--------------------------------

Handlers can receive typed request/response wrappers. Example controller method:

```go
func (ctrl *Controller) Create(resp *kyugo.Response, req *kyugo.Request) {
    // fetch validated body registered with router.ValidateBody
    body, ok := kyugo.BodyAsRequest[*dto.CreateProductRequest](req)
	if !ok {
		msg, ok := req.Message("locale.bad_request")
		if !ok || msg == "" {
			msg = "Bad Request"
		}
		resp.JSON(http.StatusBadRequest, msg, nil, kyugo.ErrorExtras{
			Code: "BAD_REQUEST",
			Type: "VALIDATION_ERROR",
		})
	}
    // localized message
    msg, _ := req.Message("locale.product_created")
    resp.JSON(http.StatusOK, msg, body)
}
```

Notes
-----
- Ensure you registered the DTO in the route with `.ValidateBody(&dto.CreateProductRequest{})` so the router validates and populates the typed body.
- To use the shorthand re-exports, import the top-level package:

```go
import kyugo "github.com/go-kyugo/kyugo"
```

Then you can reference `kyugo.Request`, `kyugo.Response`, `kyugo.Adapt`, etc.


Where to look
--------------

Error envelopes
---------------

Error responses follow this structure:

```json
{
    "status": "error",
    "code": 400,
    "error": {
        "type": "INVALID_BODY",
        "code": "INVALID_REQUEST",
        "message": "Invalid JSON body",
        "fields": [ /* optional field errors */ ],
        "meta": { /* optional meta object */ }
    }
}
```

Use `ErrorExtras` to set the `error.code` and `error.type` values when calling `resp.JSON` or `ErrorResponse` directly. Example:

```go
// via resp.JSON
resp.JSON(http.StatusBadRequest, "Invalid JSON body", nil, kyugo.ErrorExtras{Code: "INVALID_REQUEST", Type: "INVALID_BODY"})

// or directly
kyugo.ErrorResponse(w, http.StatusBadRequest, "Invalid JSON body", nil, kyugo.ErrorExtras{Code: "INVALID_REQUEST", Type: "INVALID_BODY"})
```

- Example application: [examples/usage](examples/usage)
- Server implementation: [server/server.go](server/server.go)
- Router and validation logic: [router/router.go](router/router.go)
- Configuration helpers: [config/config.go](config/config.go)
- Simple runtime registry: [registry/registry.go](registry/registry.go)

Next steps
----------

- Run the example and adapt the configuration to your environment.
- Add migrations under `database/migrations` if you use the database helper.
- Implement your application controllers under `examples/usage/http/controllers` as a reference.

If you want, I can run a quick `go vet` / `go build` against the example and update this README with any build notes.

