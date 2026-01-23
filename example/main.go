package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"kyugo.dev/kyugo/v1"
	cfg "kyugo.dev/kyugo/v1/config"
	"kyugo.dev/kyugo/v1/example/http/controllers"
	"kyugo.dev/kyugo/v1/example/service/product"
	"kyugo.dev/kyugo/v1/example/service/user"
	logger "kyugo.dev/kyugo/v1/logger"
)

func main() {
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
}

func registerServices(server *kyugo.Server) {
	logger.Info("Registering services", nil)

	userService := user.NewService()
	productService := product.NewService()
	server.RegisterService("user", userService)
	server.RegisterService("product", productService)
}
