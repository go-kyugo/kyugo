package controllers

import (
	"net/http"

	"kyugo.dev/kyugo/v1"
	"kyugo.dev/kyugo/v1/example/dto"
	"kyugo.dev/kyugo/v1/example/http/middleware"
	"kyugo.dev/kyugo/v1/example/services"
)

type ProductService interface {
	GetByID(id int) (map[string]interface{}, error)
}

type Controller struct {
	ProductService ProductService
}

func (ctrl *Controller) Init(s *kyugo.Server) {
	ctrl.ProductService = s.Service(services.Product).(ProductService)
}

func (c *Controller) Index(resp *kyugo.Response, req *kyugo.Request) {
	//resp.JSON(http.StatusOK, map[string]interface{}{"list": []int{1, 2, 3}})
}

func (c *Controller) Create(resp *kyugo.Response, req *kyugo.Request) {
	product, ok := kyugo.BodyAsRequest[*dto.CreateProductRequest](req)
	if !ok {
		resp.JSON(http.StatusBadRequest, "request body is required", nil, kyugo.ErrorExtras{
			Code: "BAD_REQUEST",
			Type: "VALIDATION_ERROR",
		})
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

	resp.JSON(http.StatusOK, msg, product)
}

func (c *Controller) Show(resp *kyugo.Response, req *kyugo.Request) {

	//resp.JSON(http.StatusOK, map[string]interface{}{"id": id})
}

func (c *Controller) Update(resp *kyugo.Response, req *kyugo.Request) {

	//resp.JSON(http.StatusOK, map[string]interface{}{"updated": true, "id": id})
}

func (c *Controller) Delete(resp *kyugo.Response, req *kyugo.Request) {

	//resp.JSON(http.StatusOK, map[string]interface{}{"deleted": true, "id": id})
}

func (ctrl *Controller) RegisterRoutes(router *kyugo.Router) {
	group := router.Group("/products")

	group.Get("/", kyugo.Adapt(ctrl.Index)).ValidateQuery(nil)
	group.Post("/", kyugo.Adapt(ctrl.Create)).ValidateBody(&dto.CreateProductRequest{}).Middleware(middleware.Example)
	group.Get("/{productID:[0-9]+}", kyugo.Adapt(ctrl.Show))
	group.Patch("/{productID:[0-9]+}", kyugo.Adapt(ctrl.Update)).ValidateBody(&dto.CreateProductRequest{})
	group.Delete("/{productID:[0-9]+}", kyugo.Adapt(ctrl.Delete))
}
