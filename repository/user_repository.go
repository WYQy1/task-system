package repository

import (
	"database/sql"
	"task-system/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user model.User) (int64, error) {
	result, err := r.db.Exec("INSERT INTO users (name, age, gender) VALUES (?, ?, ?)", user.Name, user.Age, user.Gender)
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

func (r *UserRepository) GetByID(id int) (model.User, error) {
	return scanUser(r.db.QueryRow("SELECT id, name, age, gender FROM users WHERE id = ?", id))
}

func (r *UserRepository) GetAll() ([]model.User, error) {
	rows, err := r.db.Query("SELECT id, name, age, gender FROM users ORDER BY id ASC")
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

func (r *UserRepository) Update(user model.User) (int64, error) {
	result, err := r.db.Exec("UPDATE users SET name = ?, age = ?, gender = ? WHERE id = ?", user.Name, user.Age, user.Gender, user.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (r *UserRepository) DeleteWithTasks(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM tasks WHERE user_id = ?", id); err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}
