package service

import (
	"errors"
	"strings"
	"task-system/model"
	"task-system/repository"
)

type TaskService struct {
	userRepo *repository.UserRepository
	taskRepo *repository.TaskRepository
}

func NewTaskService(userRepo *repository.UserRepository, taskRepo *repository.TaskRepository) *TaskService {
	return &TaskService{userRepo: userRepo, taskRepo: taskRepo}
}
func validateTaskID(id int) error {
	if id <= 0 {
		return ErrInvalidTaskID
	}
	return nil
}
func normalizeAndValidateTaskTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" || len([]rune(title)) > 100 {
		return "", ErrInvalidTaskTitle
	}
	return title, nil
}
func (s *TaskService) Create(task model.Task) (int64, error) {
	err := validateUserID(task.UserID)
	if err != nil {
		return 0, err
	}
	title, err := normalizeAndValidateTaskTitle(task.Title)
	if err != nil {
		return 0, err
	}
	task.Title = title
	id, err := s.taskRepo.Create(task)
	if errors.Is(err, repository.ErrUserNotFound) {
		return 0, ErrUserNotFound
	}
	return id, err
}

func (s *TaskService) GetByID(id int) (model.Task, error) {
	err := validateTaskID(id)
	if err != nil {
		return model.Task{}, err
	}
	task, err := s.taskRepo.GetByID(id)
	if errors.Is(err, repository.ErrTaskNotFound) {
		return model.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return model.Task{}, err
	}
	return task, nil
}

func (s *TaskService) GetByUserID(userID int) ([]model.Task, error) {
	if err := validateUserID(userID); err != nil {
		return nil, err
	}
	if _, err := s.userRepo.GetByID(userID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return s.taskRepo.GetByUserID(userID)
}

func isValidTaskStatus(status string) bool {
	switch status {
	case "pending", "in_progress", "completed":
		return true
	default:
		return false
	}
}

func (s *TaskService) Update(task model.Task) (int64, error) {
	err := validateTaskID(task.ID)
	if err != nil {
		return 0, err
	}
	title, err := normalizeAndValidateTaskTitle(task.Title)
	if err != nil {
		return 0, err
	}
	task.Title = title
	if !isValidTaskStatus(task.Status) {
		return 0, ErrInvalidTaskStatus
	}

	rowsAffected, err := s.taskRepo.Update(task)
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		_, err := s.GetByID(task.ID)
		if err != nil {
			return 0, err
		}
	}
	return rowsAffected, nil
}

func (s *TaskService) Delete(id int) (int64, error) {
	err := validateTaskID(id)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := s.taskRepo.Delete(id)
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		return 0, ErrTaskNotFound
	}
	return rowsAffected, nil
}
