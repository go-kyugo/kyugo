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
}
