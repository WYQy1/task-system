package service

//只负责业务逻辑的判断 其他具体的sql操作交给repo层处理
import (
	"errors"
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

func validateUserID(id int) error {
	if id <= 0 {
		return ErrInvalidUserID
	}
	return nil

}
func normalizeAndValidateUserName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", ErrInvalidUserName
	}
	if len([]rune(name)) > 20 {
		return "", ErrNameTooLong
	}

	return name, nil
}
func validateUserAge(age int) error {
	if age < 1 || age > 150 {
		return ErrInvalidUserAge
	}
	return nil
}
func normalizeAndValidateUserGender(gender string) (string, error) {
	gender = strings.ToLower(strings.TrimSpace(gender))
	switch gender {
	case "男", "male":
		return "male", nil
	case "女", "female":
		return "female", nil
	case "":
		return "", nil
	default:
		return "", ErrInvalidUserGender
	}

}

func (s *UserService) Create(user model.User) (int64, error) {
	name, err := normalizeAndValidateUserName(user.Name)
	if err != nil {
		return 0, err
	}
	user.Name = name
	err = validateUserAge(user.Age)
	if err != nil {
		return 0, err
	}
	gender, err := normalizeAndValidateUserGender(user.Gender)
	if err != nil {
		return 0, err
	}
	user.Gender = gender           
	return s.userRepo.Create(user) 
}

func (s *UserService) GetByID(id int) (model.User, error) {
	err := validateUserID(id)
	if err != nil {
		return model.User{}, err
	}
	user, err := s.userRepo.GetByID(id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return model.User{}, ErrUserNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (s *UserService) GetAll() ([]model.User, error) {
	return s.userRepo.GetAll()
}

func (s *UserService) Update(user model.User) (int64, error) {
	err := validateUserID(user.ID)
	if err != nil {
		return 0, err
	}
	name, err := normalizeAndValidateUserName(user.Name)
	if err != nil {
		return 0, err
	}
	user.Name = name
	err = validateUserAge(user.Age)
	if err != nil {
		return 0, err
	}
	gender, err := normalizeAndValidateUserGender(user.Gender)
	if err != nil {
		return 0, err
	}
	user.Gender = gender
	rowsAffected, err := s.userRepo.Update(user)
	if err != nil {
		return 0, err
	}
	if rowsAffected == 0 {
		_, err := s.GetByID(user.ID)
		if err != nil {
			return 0, err
		}
	}
	return rowsAffected, nil
}

func (s *UserService) Delete(id int) error {
	err := validateUserID(id)
	if err != nil {
		return err
	}
	err = s.userRepo.DeleteWithTasks(id)
	if errors.Is(err, repository.ErrUserNotFound) {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}
	return nil
}
