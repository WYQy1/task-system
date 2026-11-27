package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"
	"task-system/model"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// taskColumns 统一列顺序，避免不同语句的 SELECT 顺序与 Scan 对不上。
const taskColumns = "id, user_id, title, description, status, created_at, due_at, reminded_at"

// nullableTime 把 *time.Time 转成驱动能接受的值：nil → NULL。
func nullableTime(t *time.Time) any {
	if t == nil {
		return nil
	}
	return *t
}

func scanTask(scanner interface{ Scan(...any) error }) (model.Task, error) {
	var task model.Task
	var dueAt, remindedAt sql.NullTime
	err := scanner.Scan(&task.ID, &task.UserID, &task.Title, &task.Description,
		&task.Status, &task.CreatedAt, &dueAt, &remindedAt)
	if err != nil {
		return model.Task{}, err
	}
	if dueAt.Valid {
		v := dueAt.Time
		task.DueAt = &v
	}
	if remindedAt.Valid {
		v := remindedAt.Time
		task.RemindedAt = &v
	}
	return task, nil
}

func (r *TaskRepository) Create(ctx context.Context, task model.Task) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		"INSERT INTO tasks (user_id, title, description, due_at) VALUES (?, ?, ?, ?)",
		task.UserID, task.Title, task.Description, nullableTime(task.DueAt))
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
	task, err := scanTask(r.db.QueryRowContext(ctx, "SELECT "+taskColumns+" FROM tasks WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.Task{}, ErrTaskNotFound
	}
	if err != nil {
		return model.Task{}, err
	}
	return task, nil
}

func (r *TaskRepository) GetByUserID(ctx context.Context, userID int) ([]model.Task, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT "+taskColumns+" FROM tasks WHERE user_id = ? ORDER BY id ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
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
	result, err := r.db.ExecContext(ctx,
		"UPDATE tasks SET title = ?, description = ?, status = ?, due_at = ? WHERE id = ?",
		task.Title, task.Description, task.Status, nullableTime(task.DueAt), task.ID)
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

// ScanOverdue 捞出「未完成 + 已过截止时间 + 还没提醒过」的任务，供超期扫描器使用。
// 用 reminded_at IS NULL 过滤，保证同一个任务不会每个扫描周期都被重复提醒。
func (r *TaskRepository) ScanOverdue(ctx context.Context, limit int) ([]model.Task, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT "+taskColumns+" FROM tasks "+
			"WHERE status <> 'completed' AND due_at IS NOT NULL AND due_at < NOW() AND reminded_at IS NULL "+
			"ORDER BY due_at ASC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]model.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}

// MarkReminded 写入提醒时间。带上 reminded_at IS NULL 条件，多个实例同时扫也不会重复提醒。
func (r *TaskRepository) MarkReminded(ctx context.Context, id int) (int64, error) {
	result, err := r.db.ExecContext(ctx,
		"UPDATE tasks SET reminded_at = NOW() WHERE id = ? AND reminded_at IS NULL", id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
