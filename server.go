package kyugo

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	cfg "kyugo.dev/kyugo/v1/config"
	database "kyugo.dev/kyugo/v1/database"
	logger "kyugo.dev/kyugo/v1/logger"
)

// Options configures the created server.

type Options struct {
	// Config optionally carries the full application configuration. When set
	// `New` will prefer values from this config (such as server address and
	// database settings) unless explicitly overridden on Options.
	Config  *cfg.Config
	Handler http.Handler // optional; if nil, a default router will be used
	// DefaultMiddlewares are applied to the provided Handler (or default router)
	// inside the server during creation. They are applied in order.
	DefaultMiddlewares []func(http.Handler) http.Handler
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
}

// LoggerConfig represents structured logger configuration passed to the server.
type LoggerConfig struct {
	Type    string       // e.g. "color", "simple", "none"
	Level   logger.Level // numeric level
	Enabled bool         // enable or disable logging
	Color   bool         // prefer colorized output when true
}

type Server struct {
	srv    *http.Server
	logger *logger.Logger
	Config interface{}
	DB     *database.DB
	// services holds arbitrary service instances registered with the server.
	services map[string]interface{}
	svcMu    sync.RWMutex
}

func NewServer(opts Options) (*Server, error) {
	// prefer options.Config values when available
	var cfgSrc *cfg.Config
	if opts.Config != nil {
		cfgSrc = opts.Config
	} else {
		cfgSrc = &cfg.ConfigVar
	}
	// determine address
	addr := ":8080"
	if cfgSrc != nil && cfgSrc.Server.Host != "" {
		addr = fmt.Sprintf("%s:%d", cfgSrc.Server.Host, cfgSrc.Server.Port)
	}

	// create logger according to Config debug flag
	var std *logger.Logger
	lvl := logger.LevelInfo
	logType := "none"
	if cfgSrc != nil && cfgSrc.App.Debug {
		logType = "color"
		lvl = logger.LevelDebug
	}
	lc := LoggerConfig{Type: logType, Level: lvl, Enabled: true, Color: logType == "color"}

	if !lc.Enabled || lc.Type == "none" {
		std = logger.NewNop()
	} else if lc.Type == "simple" {
		std = logger.NewSimple(os.Stdout, lc.Level, lc.Color)
	} else if lc.Type == "color" {
		std = logger.NewConsole(os.Stdout, lc.Level, lc.Color)
	} else {
		std = logger.NewConsole(os.Stdout, lc.Level, lc.Color)
	}
	logger.SetStd(std)

	// prepare messages map (load resources once)
	var msgs map[string]string
	if cfgSrc != nil && cfgSrc.App.Language != "" {
		_ = LoadFromFS(os.DirFS("resources/langs"))
		msgs = GetAll(cfgSrc.App.Language)
	}

	// prepare base handler: prefer provided Handler, otherwise create router
	var h http.Handler
	var base http.Handler
	if opts.Handler != nil {
		base = opts.Handler
	} else {
		rt := NewRouter()
		base = rt.Handler()
	}

	// if messages were loaded, wrap base to inject them into the request context
	if msgs != nil {
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), MessagesKey, msgs)
			base.ServeHTTP(w, r.WithContext(ctx))
		})
	} else {
		h = base
	}

	if len(opts.DefaultMiddlewares) > 0 {
		for i := len(opts.DefaultMiddlewares) - 1; i >= 0; i-- {
			mw := opts.DefaultMiddlewares[i]
			h = mw(h)
		}
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      h,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
	}

	s := &Server{srv: srv, logger: std, services: make(map[string]interface{})}

	// connect database if present in config
	if cfgSrc != nil {
		/*if cfgSrc.Database.Type != "" {
			db, err := database.ConnectFromConfig(cfgSrc.Database)
			if err != nil {
				return nil, err
			}
			s.DB = db
			database.SetDefault(db)
		}*/
	}

	return s, nil
}

// ListenAndServe starts the HTTP server.
func (s *Server) Start() error {
	s.logger.Info(fmt.Sprintf("Server.Start %s=%s", logger.Colorize("addr", "36"), s.srv.Addr), nil)
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen failed: %w", err)
	}
	return nil
}

// RegisterService stores a service instance under the provided name.
func (s *Server) RegisterService(name string, svc interface{}) {
	if s == nil {
		return
	}
	s.svcMu.Lock()
	defer s.svcMu.Unlock()
	s.services[name] = svc
}

// Service returns a previously registered service by name or nil if not found.
func (s *Server) Service(name string) interface{} {
	if s == nil {
		return nil
	}
	s.svcMu.RLock()
	defer s.svcMu.RUnlock()
	return s.services[name]
}
