package service

import (
	"database/sql"
	"errors"
	"strings"
	"task-system/model"
	"task-system/repository"
)

type TaskService struct {
	taskRepo *repository.TaskRepository
}

func NewTaskService(taskRepo *repository.TaskRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo}
}

func (s *TaskService) Create(task model.Task) (int64, error) {
	task.Title = strings.TrimSpace(task.Title)
	if task.Title == "" || len([]rune(task.Title)) > 100 {
		return 0, ErrInvalidTaskTitle
	}

	id, err := s.taskRepo.Create(task)
	if errors.Is(err, repository.ErrUserNotFound) {
		return 0, ErrUserNotFound
	}
	return id, err
}

func (s *TaskService) GetByID(id int) (model.Task, error) {
	return s.taskRepo.GetByID(id)
}

func (s *TaskService) GetByUserID(userID int) ([]model.Task, error) {
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
	task.Title = strings.TrimSpace(task.Title)
	if task.Title == "" || len([]rune(task.Title)) > 100 {
		return 0, ErrInvalidTaskTitle
	}
	if !isValidTaskStatus(task.Status) {
		return 0, ErrInvalidTaskStatus
	}

	rowsAffected, err := s.taskRepo.Update(task)
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		_, err := s.taskRepo.GetByID(task.ID)
		if err != nil {
			return 0, err
		}
	}
	return rowsAffected, nil
}

func (s *TaskService) Delete(id int) (int64, error) {
	return s.taskRepo.Delete(id)
}

var _ = sql.ErrNoRows
