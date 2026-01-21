package controllers

import (
	"net/http"
	"strconv"
	"time"

	"kyugo.dev/kyugo/v1/examples/usage/dto"
	exmw "kyugo.dev/kyugo/v1/examples/usage/http/middleware"
	service "kyugo.dev/kyugo/v1/examples/usage/services"
	"kyugo.dev/kyugo/v1/response"
	pr "kyugo.dev/kyugo/v1/router"
	srv "kyugo.dev/kyugo/v1/server"
)

// ProductService defines the minimal service used by the controller.
type ProductService interface {
	GetByID(id int) (map[string]interface{}, error)
}

type Controller struct {
	ProductService ProductService
}

// Init injects services from the server into the controller.
func (ctrl *Controller) Init(s *srv.Server) {
	ctrl.ProductService = s.Service(service.Product).(ProductService)
}

func (c *Controller) Index(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]interface{}{"list": []int{1, 2, 3}})
}

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	// get validated request DTO

	product := dto.Product{
		ID:        1,
		Name:      "New Product",
		Price:     99.99,
		CreatedAt: time.Now(),
	}

	if c.ProductService != nil {
		if data, err := c.ProductService.GetByID(product.ID); err == nil && data != nil {
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
	}

	response.Success(w, map[string]interface{}{"created": true, "product": product})
}

func (c *Controller) Show(w http.ResponseWriter, r *http.Request) {
	pid := pr.Param(r, "productID")
	if pid == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "missing_parameter", "productID obrigatório", nil)
		return
	}
	id, err := strconv.Atoi(pid)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid_parameter", "productID inválido", nil)
		return
	}
	/*if c.UserService != nil {
	    if data, err := c.UserService.GetByID(id); err == nil {
	        response.Success(w, data)
	        return
	    }
	}*/
	response.Success(w, map[string]interface{}{"id": id})
}

func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	pid := pr.Param(r, "productID")
	if pid == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "missing_parameter", "productID obrigatório", nil)
		return
	}
	id, err := strconv.Atoi(pid)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid_parameter", "productID inválido", nil)
		return
	}
	response.Success(w, map[string]interface{}{"updated": true, "id": id})
}

func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	pid := pr.Param(r, "productID")
	if pid == "" {
		response.Error(w, http.StatusBadRequest, "invalid_request", "missing_parameter", "productID obrigatório", nil)
		return
	}
	id, err := strconv.Atoi(pid)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_request", "invalid_parameter", "productID inválido", nil)
		return
	}
	response.Success(w, map[string]interface{}{"deleted": true, "id": id})
}

// RegisterRoutes registers the controller routes into the provided router.
// Uses Group (subrouter equivalent) and returns chainable validators.
func (ctrl *Controller) RegisterRoutes(router *pr.Router) {
	group := router.Group("/products")

	group.Get("/", ctrl.Index).ValidateQuery(nil)
	group.Post("/", ctrl.Create).ValidateBody(&dto.CreateProductRequest{}).Middleware(exmw.Example)
	group.Get("/{productID:[0-9]+}", ctrl.Show)
	group.Patch("/{productID:[0-9]+}", ctrl.Update).ValidateBody(&dto.CreateProductRequest{})
	group.Delete("/{productID:[0-9]+}", ctrl.Delete)
}
