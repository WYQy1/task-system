package service

import (
	"context"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"task-system/model"
	"task-system/repository"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func validatePassword(password string) error {
	if len(password) < 6 {
		return ErrInvalidPassword
	}
	return nil
}

func (a *AuthService) Register(ctx context.Context, name, password string, age int, gender string) (int64, error) {
	name, err := normalizeAndValidateUserName(name)
	if err != nil {
		return 0, err
	}
	if err := validatePassword(password); err != nil {
		return 0, err
	}
	if err := validateUserAge(age); err != nil {
		return 0, err
	}
	gender, err = normalizeAndValidateUserGender(gender)
	if err != nil {
		return 0, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	user := model.User{Name: name, Age: age, Gender: gender, Password: string(hash)}
	return a.userRepo.Create(ctx, user)
}

func (a *AuthService) Login(ctx context.Context, name, password string) (model.User, error) {
	name, err := normalizeAndValidateUserName(name)
	if err != nil {
		return model.User{}, err // 名字不合法 → 原样返回（respondError 会映射成 400）
	}
	user, err := a.userRepo.GetByName(ctx, name)
	if errors.Is(err, repository.ErrUserNotFound) {
		return model.User{}, ErrWrongPassword // 用户不存在 → 伪装成"密码错误"
	}
	if err != nil {
		return model.User{}, err // 其他数据库错误 → 原样返回（500）
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return model.User{}, ErrWrongPassword
	}
	return user, nil
}

func (a *AuthService) GetTokenVersion(ctx context.Context, userID int) (int, error) {
	return a.userRepo.GetTokenVersion(ctx, userID)
}

func (a *AuthService) Logout(ctx context.Context, userID int) error {
	_, err := a.userRepo.IncrementTokenVersion(ctx, userID)
	return err
}
