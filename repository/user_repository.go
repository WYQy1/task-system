package repository

import (
	"context"
	"database/sql"
	"errors"
	"task-system/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// **凡是 `XxxContext` 的方法，第一个实参永远是 ctx**。`QueryRowContext(ctx, sql)
func (r *UserRepository) GetTokenVersion(ctx context.Context, userID int) (int, error) {
	var version int
	err := r.db.QueryRowContext(ctx, "SELECT token_version FROM users WHERE id=?", userID).Scan(&version)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrUserNotFound
	}
	if err != nil {
		return 0, err
	}
	return version, nil
}
func (r *UserRepository) IncrementTokenVersion(ctx context.Context, userID int) (int64, error) {
	result, err := r.db.ExecContext(ctx, "UPDATE users SET token_version = token_version + 1 WHERE id = ?", userID)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		return 0, ErrUserNotFound
	}
	return rowsAffected, nil
}

func (r *UserRepository) Create(ctx context.Context, user model.User) (int64, error) {
	result, err := r.db.ExecContext(ctx, "INSERT INTO users (name, age, gender,password) VALUES (?, ?, ?,?)", user.Name, user.Age, user.Gender, user.Password)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func scanUser(scanner interface{ Scan(...any) error }) (model.User, error) {
	var user model.User
	var age sql.NullInt64
	var gender sql.NullString

	err := scanner.Scan(&user.ID, &user.Name, &age, &gender)
	if err != nil {
		return model.User{}, err
	}
	if age.Valid {
		user.Age = int(age.Int64)
	}
	if gender.Valid {
		user.Gender = gender.String
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (model.User, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, "SELECT id, name, age, gender FROM users WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}
func (r *UserRepository) GetByName(ctx context.Context, name string) (model.User, error) {
	var user model.User
	var age sql.NullInt64
	var gender sql.NullString
	err := r.db.QueryRowContext(ctx, "SELECT id, name, age, gender,password FROM users WHERE name = ?", name).Scan(&user.ID, &user.Name, &age, &gender, &user.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	if age.Valid {
		user.Age = int(age.Int64)
	}
	if gender.Valid {
		user.Gender = gender.String
	}
	return user, nil
}
func (r *UserRepository) GetAll(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, age, gender FROM users ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, user model.User) (int64, error) {
	result, err := r.db.ExecContext(ctx, "UPDATE users SET name = ?, age = ?, gender = ? WHERE id = ?", user.Name, user.Age, user.Gender, user.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *UserRepository) DeleteWithTasks(ctx context.Context, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM tasks WHERE user_id = ?", id); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return tx.Commit()
}
