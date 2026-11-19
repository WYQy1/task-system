package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"task-system/model"

	"github.com/go-sql-driver/mysql"
)

var ErrUserNotFound = errors.New("员工不存在")

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(task model.Task) (int64, error) {
	result, err := r.db.Exec("INSERT INTO tasks (user_id, title, description) VALUES (?, ?, ?)", task.UserID, task.Title, task.Description)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1452 {
			return 0, fmt.Errorf("%w: %v", ErrUserNotFound, err)
		}
		return 0, err
	}
	return result.LastInsertId()
}

func (r *TaskRepository) GetByID(id int) (model.Task, error) {
	var task model.Task
	err := r.db.QueryRow("SELECT id, user_id, title, description, status, created_at FROM tasks WHERE id = ?", id).Scan(&task.ID, &task.UserID, &task.Title, &task.Description, &task.Status, &task.CreatedAt)
	return task, err
}

func (r *TaskRepository) GetByUserID(userID int) ([]model.Task, error) {
	rows, err := r.db.Query("SELECT id, user_id, title, description, status, created_at FROM tasks WHERE user_id = ? ORDER BY id ASC", userID)
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

func (r *TaskRepository) Update(task model.Task) (int64, error) {
	result, err := r.db.Exec("UPDATE tasks SET title = ?, description = ?, status = ? WHERE id = ?", task.Title, task.Description, task.Status, task.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *TaskRepository) Delete(id int) (int64, error) {
	result, err := r.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
