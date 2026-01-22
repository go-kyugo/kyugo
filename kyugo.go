package kyugo

import (
	"kyugo.dev/kyugo/v1/config"
	"kyugo.dev/kyugo/v1/database"
	"kyugo.dev/kyugo/v1/logger"
	"kyugo.dev/kyugo/v1/request"
	"kyugo.dev/kyugo/v1/response"
	"kyugo.dev/kyugo/v1/router"
	"kyugo.dev/kyugo/v1/server"
)

// Re-export common types so callers can `import kyugo "kyugo.dev/kyugo/v1"`
// and use `kyugo.Response`, `kyugo.Request`, `kyugo.Server`, etc.
type Response = response.Response
type Request = request.Request
type Server = server.Server
type Logger = logger.Logger
type Fields = logger.Fields

// Re-export additional common packages/types
type Router = router.Router
type Config = config.Config
type DB = database.DB
