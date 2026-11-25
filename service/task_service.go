package service

import (
	"context"
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
func (s *TaskService) Create(ctx context.Context, task model.Task) (int64, error) {
	err := validateUserID(task.UserID)
	if err != nil {
		return 0, err
	}
	title, err := normalizeAndValidateTaskTitle(task.Title)
	if err != nil {
		return 0, err
	}
	task.Title = title
	id, err := s.taskRepo.Create(ctx, task)
	if errors.Is(err, repository.ErrUserNotFound) {
		return 0, ErrUserNotFound
	}
	return id, err
}

func (s *TaskService) GetByID(ctx context.Context, id int) (model.Task, error) {
	err := validateTaskID(id)
	if err != nil {
		return model.Task{}, err
	}
	task, err := s.taskRepo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrTaskNotFound) {
		return model.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return model.Task{}, err
	}
	return task, nil
}

func (s *TaskService) GetByUserID(ctx context.Context, userID int) ([]model.Task, error) {
	if err := validateUserID(userID); err != nil {
		return nil, err
	}
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return s.taskRepo.GetByUserID(ctx, userID)
}

func isValidTaskStatus(status string) bool {
	switch status {
	case "pending", "in_progress", "completed":
		return true
	default:
		return false
	}
}

func (s *TaskService) Update(ctx context.Context, task model.Task) (int64, error) {
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

	rowsAffected, err := s.taskRepo.Update(ctx, task)
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		_, err := s.GetByID(ctx, task.ID)
		if err != nil {
			return 0, err
		}
	}
	return rowsAffected, nil
}

func (s *TaskService) Delete(ctx context.Context, id int) (int64, error) {
	err := validateTaskID(id)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := s.taskRepo.Delete(ctx, id)
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		return 0, ErrTaskNotFound
	}
	return rowsAffected, nil
}
