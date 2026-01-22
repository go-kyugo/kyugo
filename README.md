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
rt := router.New()
ctrl.RegisterRoutes(rt)

opts := server.Options{
    Config: &cfg.ConfigVar,
    Handler: rt.Handler(),
    DefaultMiddlewares: []func(http.Handler) http.Handler{
        middleware.CORS(cfg.ConfigVar.Server.Cors),
        middleware.Logger,
    },
    ReadTimeout:  time.Duration(cfg.ConfigVar.Server.ReadTimeoutSeconds) * time.Second,
    WriteTimeout: time.Duration(cfg.ConfigVar.Server.WriteTimeoutSeconds) * time.Second,
}

srv, _ := server.New(opts)
srv.Start()
```

Where to look
--------------

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

