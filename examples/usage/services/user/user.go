package user

import "fmt"

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) GetByID(id int) (map[string]interface{}, error) {
	if id <= 0 {
		return nil, fmt.Errorf("not found")
	}
	return map[string]interface{}{"id": id, "name": fmt.Sprintf("User %d", id)}, nil
}
