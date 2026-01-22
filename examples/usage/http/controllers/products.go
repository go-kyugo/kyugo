package controllers

import (
	"net/http"

	"kyugo.dev/kyugo/v1/examples/usage/dto"
	exmw "kyugo.dev/kyugo/v1/examples/usage/http/middleware"
	service "kyugo.dev/kyugo/v1/examples/usage/services"
	"kyugo.dev/kyugo/v1/handler"
	"kyugo.dev/kyugo/v1/request"
	"kyugo.dev/kyugo/v1/response"
	pr "kyugo.dev/kyugo/v1/router"
	srv "kyugo.dev/kyugo/v1/server"
)

type ProductService interface {
	GetByID(id int) (map[string]interface{}, error)
}

type Controller struct {
	ProductService ProductService
}

func (ctrl *Controller) Init(s *srv.Server) {
	ctrl.ProductService = s.Service(service.Product).(ProductService)
}

func (c *Controller) Index(resp *response.Response, req *request.Request) {
	//resp.JSON(http.StatusOK, map[string]interface{}{"list": []int{1, 2, 3}})
}

func (c *Controller) Create(resp *response.Response, req *request.Request) {
	product, ok := request.BodyAsRequest[*dto.CreateProductRequest](req)
	if !ok {
		response.Error(resp.W, http.StatusBadRequest, "invalid_request", "missing_parameter", "request body is required", nil)
		return
	}

	/*if c.ProductService != nil {
		if data, err := c.ProductService.GetByID(bodyPtr.ID); err == nil && data != nil {
			if v, ok := data["id"].(int); ok {
				product.ID = v
			}
			if v, ok := data["name"].(string); ok {
				product.Name = v
			}
			if v, ok := data["price"].(float64); ok {
				product.Price = v
			}
			if v, ok := data["created_at"].(time.Time); ok {
				product.CreatedAt = v
			}
		}
	}*/

	msg, ok := req.Message("locale.product_created")
	if !ok || msg == "" {
		msg = "Product created"
	}

	resp.JSON(http.StatusOK, 200, msg, product)
}

func (c *Controller) Show(resp *response.Response, req *request.Request) {

	//resp.JSON(http.StatusOK, map[string]interface{}{"id": id})
}

func (c *Controller) Update(resp *response.Response, req *request.Request) {

	//resp.JSON(http.StatusOK, map[string]interface{}{"updated": true, "id": id})
}

func (c *Controller) Delete(resp *response.Response, req *request.Request) {

	//resp.JSON(http.StatusOK, map[string]interface{}{"deleted": true, "id": id})
}

func (ctrl *Controller) RegisterRoutes(router *pr.Router) {
	group := router.Group("/products")

	group.Get("/", handler.Adapt(ctrl.Index)).ValidateQuery(nil)
	group.Post("/", handler.Adapt(ctrl.Create)).ValidateBody(&dto.CreateProductRequest{}).Middleware(exmw.Example)
	group.Get("/{productID:[0-9]+}", handler.Adapt(ctrl.Show))
	group.Patch("/{productID:[0-9]+}", handler.Adapt(ctrl.Update)).ValidateBody(&dto.CreateProductRequest{})
	group.Delete("/{productID:[0-9]+}", handler.Adapt(ctrl.Delete))
}
