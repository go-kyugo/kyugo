package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-kyugo/kyugo"
	cfg "github.com/go-kyugo/kyugo/config"
	"github.com/go-kyugo/kyugo/example/http/controller"
	"github.com/go-kyugo/kyugo/example/service/product"
	"github.com/go-kyugo/kyugo/example/service/user"
	logger "github.com/go-kyugo/kyugo/logger"
)

func main() {
	r := kyugo.NewRouter()

	if err := cfg.LoadConfig("./config.json"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctrl := &controller.Controller{}
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
