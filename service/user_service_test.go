package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"task-system/model"
	"task-system/repository"
)

// 新增员工统一走注册（AuthService.Register）：与登录同一套密码哈希和校验，
// 也顺带避免管理员代建时写出空密码账号。
func TestAuthServiceRegister(t *testing.T) {
	db := openTestDB(t)
	authService := NewAuthService(repository.NewUserRepository(db))

	id, err := authService.Register(context.Background(), "  Alice  ", "secret123", 23, " 男 ")
	require.NoError(t, err)
	cleanupTestUser(t, db, id)
	checkUser(t, getUserFromDB(t, db, id), int(id), "Alice", 23, "male")
}

func TestAuthServiceRegisterInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		password string
		age      int
		gender   string
		wantErr  error
	}{
		{"空姓名", "", "secret123", 20, "male", ErrInvalidUserName},
		{"姓名超长", strings.Repeat("测", 21), "secret123", 20, "male", ErrNameTooLong},
		{"密码过短", "张三", "123", 20, "male", ErrInvalidPassword},
		{"年龄为0", "张三", "secret123", 0, "male", ErrInvalidUserAge},
		{"年龄过高", "张三", "secret123", 151, "male", ErrInvalidUserAge},
		{"非法性别", "张三", "secret123", 20, "unknown", ErrInvalidUserGender},
	}
	svc := NewAuthService(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Register(context.Background(), tt.userName, tt.password, tt.age, tt.gender)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUserServiceGetByID(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)
	user, err := userService.GetByID(context.Background(), id)
	require.NoError(t, err)
	checkUser(t, user, id, "TestUser", 25, "male")
}

func TestUserServiceGetByIDInvalidID(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{"ID为0", 0},
		{"负数ID", -1},
	}
	svc := NewUserService(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetByID(context.Background(), tt.id)
			assert.ErrorIs(t, err, ErrInvalidUserID)
		})
	}
}

func TestUserServiceGetByIDNotFound(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)
	deleteUserForTest(t, db, id)
	_, err := userService.GetByID(context.Background(), id)
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserServiceUpdate(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)
	rowsAffected, err := userService.Update(context.Background(), model.User{ID: id, Name: " Bob ", Age: 24, Gender: "女"})
	require.NoError(t, err)
	assert.Equal(t, int64(1), rowsAffected)
	checkUser(t, getUserFromDB(t, db, int64(id)), id, "Bob", 24, "female")
}

func TestUserServiceUpdateNoChanges(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)
	rowsAffected, err := userService.Update(context.Background(), model.User{ID: id, Name: "TestUser", Age: 25, Gender: "male"})
	require.NoError(t, err)
	assert.True(t, rowsAffected == 0 || rowsAffected == 1, "影响行数异常: %d", rowsAffected)
	checkUser(t, getUserFromDB(t, db, int64(id)), id, "TestUser", 25, "male")
}

func TestUserServiceUpdateInvalidInput(t *testing.T) {
	base := model.User{ID: 1, Name: "张三", Age: 28, Gender: "male"}
	tests := []struct {
		name    string
		mutate  func(*model.User)
		wantErr error
	}{
		{"非法ID", func(u *model.User) { u.ID = 0 }, ErrInvalidUserID},
		{"空姓名", func(u *model.User) { u.Name = "   " }, ErrInvalidUserName},
		{"姓名超长", func(u *model.User) { u.Name = strings.Repeat("中", 21) }, ErrNameTooLong},
		{"年龄过低", func(u *model.User) { u.Age = 0 }, ErrInvalidUserAge},
		{"年龄过高", func(u *model.User) { u.Age = 151 }, ErrInvalidUserAge},
		{"非法性别", func(u *model.User) { u.Gender = "unknown" }, ErrInvalidUserGender},
	}
	svc := NewUserService(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := base
			tt.mutate(&user)
			_, err := svc.Update(context.Background(), user)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestUserServiceUpdateNotFound(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)
	deleteUserForTest(t, db, id)
	_, err := userService.Update(context.Background(), model.User{ID: id, Name: "Alice", Age: 23, Gender: "male"})
	require.ErrorIs(t, err, ErrUserNotFound)
}

func TestUserServiceDelete(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)
	require.NoError(t, userService.Delete(context.Background(), id))
	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", id).Scan(&count))
	assert.Zero(t, count, "期望用户已删除")
}

func TestUserServiceDeleteInvalidID(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{"ID为0", 0},
		{"负数ID", -1},
	}
	svc := NewUserService(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Delete(context.Background(), tt.id)
			assert.ErrorIs(t, err, ErrInvalidUserID)
		})
	}
}

func TestUserServiceDeleteNotFound(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)
	deleteUserForTest(t, db, id)
	err := userService.Delete(context.Background(), id)
	require.ErrorIs(t, err, ErrUserNotFound)
}
