package kyugo

import (
	database "github.com/go-kyugo/kyugo/database"
	logger "github.com/go-kyugo/kyugo/logger"
)

// Component is a base helper intended to be embedded in controller-like
// types. It provides accessors to common server resources.
type Component struct {
	server *Server
}

// Init initializes the component with the given server. It does NOT
// register routes; route registration must be performed by calling
// RegisterRoutes on the component (router.Controller will call both).
func (c *Component) Init(s *Server) {
	if c == nil {
		return
	}
	c.server = s
}

// Server returns the associated server instance (may be nil).
func (c *Component) Server() *Server {
	if c == nil {
		return nil
	}
	return c.server
}

// Service returns a previously-registered service by name or nil.
func (c *Component) Service(name string) interface{} {
	if c == nil || c.server == nil {
		return nil
	}
	return c.server.Service(name)
}

// LookupService returns the service and a boolean indicating presence.
func (c *Component) LookupService(name string) (interface{}, bool) {
	if c == nil || c.server == nil {
		return nil, false
	}
	v := c.server.Service(name)
	return v, v != nil
}

// Logger returns the server logger.
func (c *Component) Logger() *logger.Logger {
	if c == nil || c.server == nil {
		return nil
	}
	return c.server.logger
}

// DB returns the configured database instance (may be nil).
func (c *Component) DB() *database.DB {
	if c == nil || c.server == nil {
		return nil
	}
	return c.server.DB
}

// Config returns the server configuration (opaque type).
func (c *Component) Config() interface{} {
	if c == nil || c.server == nil {
		return nil
	}
	return c.server.Config
}
