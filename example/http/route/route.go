package route

import (
	"net/http"

	"github.com/go-kyugo/kyugo"
	"github.com/go-kyugo/kyugo/example/http/controller/products"
)

// Routing is an essential part of any Kyugo application.
// Defining routes is the action of associating a URI, sometimes having parameters,
// with a handler which will process the request and respond to it.
//
// This file contains your main route registering function that is passed to server.RegisterRoutes().

func Register(server *kyugo.Server, router *kyugo.Router) {

	router.Get("/docs", func(resp *kyugo.Response, req *kyugo.Request) {
		if err := resp.ServeFile("docs/index.html"); err != nil {
			resp.JSON(http.StatusNotFound, "File not found", nil)
		}
	})

	router.Get("/hello/{name}", func(resp *kyugo.Response, req *kyugo.Request) {
		name := req.Param("name")
		resp.JSON(http.StatusOK, "ok", map[string]string{"message": "Hello, " + name + "!"})
	})

	router.Controller(products.NewController())

	// 	name := request.PathParam("name")
}
