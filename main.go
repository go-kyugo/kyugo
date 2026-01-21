package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	cfg "kyugo.dev/kyugo/v1/config"
	controllers "kyugo.dev/kyugo/v1/examples/usage/http/controllers"
	product "kyugo.dev/kyugo/v1/examples/usage/services/product"
	user "kyugo.dev/kyugo/v1/examples/usage/services/user"
	"kyugo.dev/kyugo/v1/logger"
	"kyugo.dev/kyugo/v1/middleware"
	pr "kyugo.dev/kyugo/v1/router"
	"kyugo.dev/kyugo/v1/server"
)

func main() {

	r := pr.New()

	if err := cfg.LoadConfig("./config.json"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctrl := &controllers.Controller{}
	ctrl.RegisterRoutes(r)

	opts := server.Options{
		// pass full config to server.Options; server.New will prefer this
		// when populating address, database, etc.
		Config:  &cfg.ConfigVar,
		Handler: r.Handler(),
		DefaultMiddlewares: []func(http.Handler) http.Handler{
			middleware.CORS(cfg.ConfigVar.Server.Cors),
			middleware.Logger,
		},
		ReadTimeout:  time.Duration(cfg.ConfigVar.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.ConfigVar.Server.WriteTimeoutSeconds) * time.Second,
	}

	srv, err := server.New(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	registerServices(srv)
	ctrl.Init(srv)

	srv.Start()
}

func registerServices(server *server.Server) {
	logger.Info("Registering services", nil)

	userService := user.NewService()
	productService := product.NewService()
	server.RegisterService("user", userService)
	server.RegisterService("product", productService)

	// Services represent the Domain/Business layer.
	// This is where the core logic and value of your application resides.
	// This function is where you will register your services in the server's
	// service container to make them accessible to dependents.
	// https://goyave.dev/basics/services.html#service-container

	// TODO register services
}
