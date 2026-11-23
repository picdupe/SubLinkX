package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	ID           int
	Username     string
	Password     string
	TokenVersion int64 `gorm:"not null;default:1" json:"-"`
	Role         string
	Nickname     string
}

func (user *User) Create() error { // 创建用户
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hash)
	return DB.Create(user).Error
}
func (user *User) Set(UpdateUser *User) error { // 设置用户
	updates := map[string]any{"username": UpdateUser.Username}
	if UpdateUser.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(UpdateUser.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		updates["password"] = string(hash)
		updates["token_version"] = gorm.Expr("token_version + 1")
	}
	return DB.Model(&User{}).Where("username = ?", user.Username).Updates(updates).Error
}
func (user *User) Verify() error { // 验证用户
	password := user.Password
	if err := DB.Where("username = ?", user.Username).First(user).Error; err != nil {
		return err
	}
	storedPassword := user.Password
	if _, err := bcrypt.Cost([]byte(storedPassword)); err == nil {
		return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	}
	if storedPassword != password {
		return bcrypt.ErrMismatchedHashAndPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	result := DB.Model(&User{}).Where("id = ? AND password = ?", user.ID, storedPassword).Update("password", string(hash))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return bcrypt.ErrMismatchedHashAndPassword
	}
	return nil
}

func (user *User) Find() error { // 查找用户
	return DB.Where("username = ? ", user.Username).First(user).Error
}

func (user *User) All() ([]User, error) { // 获取所有用户
	var users []User
	err := DB.Find(&users).Error
	return users, err
}

func (user *User) Del() error { // 删除用户
	return DB.Delete(user).Error
}
