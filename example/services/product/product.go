package product

import (
	"fmt"
	"time"
)

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) GetByID(id int) (map[string]interface{}, error) {
	if id <= 0 {
		return nil, fmt.Errorf("not found")
	}
	return map[string]interface{}{"id": id, "name": fmt.Sprintf("Product %d", id), "price": 99.99, "created_at": time.Now()}, nil
}
