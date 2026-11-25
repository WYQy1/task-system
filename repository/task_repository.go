package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"task-system/model"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task model.Task) (int64, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", task.UserID, task.Title, task.Description)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1452 {
			return 0, fmt.Errorf("%w: %v", ErrUserNotFound, err)
		}
		return 0, err
	}
	return result.LastInsertId()
}

func (r *TaskRepository) GetByID(ctx context.Context, id int) (model.Task, error) {
	var task model.Task
	err := r.db.QueryRowContext(ctx, "SELECT id, user_id, title, description, status, created_at FROM tasks WHERE id = ?", id).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return model.Task{}, err
	}
	return task, nil
}

func (r *TaskRepository) GetByUserID(ctx context.Context, userID int) ([]model.Task, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, user_id, title, description, status, created_at FROM tasks WHERE user_id = ? ORDER BY id ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		var task model.Task
		if err := rows.Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, task model.Task) (int64, error) {
	result, err := r.db.ExecContext(ctx, "UPDATE tasks SET title = ?, description = ?, status = ? WHERE id = ?", task.Title, task.Description, task.Status, task.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *TaskRepository) Delete(ctx context.Context, id int) (int64, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
