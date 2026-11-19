package service

import (
	"database/sql"
	"strings"
	"task-system/model"
	"task-system/repository"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Create(user model.User) (int64, error) {
	user.Name = strings.TrimSpace(user.Name)
	if user.Name == "" {
		return 0, ErrInvalidUserName
	}
	return s.userRepo.Create(user)
}

func (s *UserService) GetByID(id int) (model.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) GetAll() ([]model.User, error) {
	return s.userRepo.GetAll()
}

func (s *UserService) Update(user model.User) (int64, error) {
	user.Name = strings.TrimSpace(user.Name)
	if user.Name == "" {
		return 0, ErrInvalidUserName
	}

	rowsAffected, err := s.userRepo.Update(user)
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		_, err := s.userRepo.GetByID(user.ID)
		if err != nil {
			return 0, err
		}
	}
	return rowsAffected, nil
}

func (s *UserService) Delete(id int) error {
	return s.userRepo.DeleteWithTasks(id)
}

var _ = sql.ErrNoRows
