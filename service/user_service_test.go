package service

import (
	"errors"
	"strings"
	"testing"

	"task-system/model"
)

func TestUserServiceCreate(t *testing.T) {
	db, userService := newTestUserService(t)
	id, err := userService.Create(model.User{Name: "  Alice  ", Age: 23, Gender: " 男 "})
	if err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	cleanupTestUser(t, db, id)

	checkUser(t, getUserFromDB(t, db, id), int(id), "Alice", 23, "male")
}

func TestUserServiceCreateInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		user    model.User
		wantErr error
	}{
		{"空姓名", model.User{Name: "", Age: 20, Gender: "male"}, ErrInvalidUserName},
		{"姓名超长", model.User{Name: strings.Repeat("测", 21), Age: 20, Gender: "male"}, ErrNameTooLong},
		{"年龄为0", model.User{Name: "张三", Age: 0, Gender: "male"}, ErrInvalidUserAge},
		{"年龄过高", model.User{Name: "张三", Age: 151, Gender: "male"}, ErrInvalidUserAge},
		{"非法性别", model.User{Name: "张三", Age: 20, Gender: "unknown"}, ErrInvalidUserGender},
	}

	svc := NewUserService(nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := svc.Create(tt.user); !errors.Is(err, tt.wantErr) {
				t.Errorf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestUserServiceGetByID(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)

	user, err := userService.GetByID(id)
	if err != nil {
		t.Fatalf("查询用户失败: %v", err)
	}
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
			if _, err := svc.GetByID(tt.id); !errors.Is(err, ErrInvalidUserID) {
				t.Errorf("期望 ErrInvalidUserID，实际得到 %v", err)
			}
		})
	}
}

func TestUserServiceGetByIDNotFound(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)
	deleteUserForTest(t, db, id)

	if _, err := userService.GetByID(id); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("期望 ErrUserNotFound，实际得到 %v", err)
	}
}

func TestUserServiceUpdate(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)

	rowsAffected, err := userService.Update(model.User{ID: id, Name: " Bob ", Age: 24, Gender: "女"})
	if err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}
	if rowsAffected != 1 {
		t.Fatalf("期望影响 1 行，实际 %d", rowsAffected)
	}

	checkUser(t, getUserFromDB(t, db, int64(id)), id, "Bob", 24, "female")
}

func TestUserServiceUpdateNoChanges(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)

	rowsAffected, err := userService.Update(model.User{ID: id, Name: "TestUser", Age: 25, Gender: "male"})
	if err != nil {
		t.Fatalf("无变化更新不应报错: %v", err)
	}
	if rowsAffected != 0 && rowsAffected != 1 {
		t.Fatalf("影响行数异常: %d", rowsAffected)
	}

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
			if _, err := svc.Update(user); !errors.Is(err, tt.wantErr) {
				t.Fatalf("期望错误 %v，实际得到 %v", tt.wantErr, err)
			}
		})
	}
}

func TestUserServiceUpdateNotFound(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)
	deleteUserForTest(t, db, id)

	if _, err := userService.Update(model.User{ID: id, Name: "Alice", Age: 23, Gender: "male"}); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("期望 ErrUserNotFound，实际得到 %v", err)
	}
}

func TestUserServiceDelete(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)

	if err := userService.Delete(id); err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", id).Scan(&count); err != nil {
		t.Fatalf("查询用户数量失败: %v", err)
	}
	if count != 0 {
		t.Fatalf("期望用户已删除，但依然存在")
	}
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
			if err := svc.Delete(tt.id); !errors.Is(err, ErrInvalidUserID) {
				t.Fatalf("期望 ErrInvalidUserID，实际得到 %v", err)
			}
		})
	}
}

func TestUserServiceDeleteNotFound(t *testing.T) {
	db, userService := newTestUserService(t)
	id := createTestUser(t, db)
	deleteUserForTest(t, db, id)

	if err := userService.Delete(id); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("期望 ErrUserNotFound，实际得到 %v", err)
	}
}
