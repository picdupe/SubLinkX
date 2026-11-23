package settings

import (
	"fmt"
	"sublink/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 重置默认用户
func ResetUser(username string, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password are required")
	}
	if len(password) < 8 || len(password) > 72 {
		return fmt.Errorf("password must be between 8 and 72 bytes")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := &models.User{Username: username, Password: string(hash), Role: "admin", Nickname: "管理员"}
	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		var maxTokenVersion int64
		if err := tx.Model(&models.User{}).Select("COALESCE(MAX(token_version), 0)").Scan(&maxTokenVersion).Error; err != nil {
			return err
		}
		user.TokenVersion = maxTokenVersion + 1
		if err := tx.Where("1 = 1").Delete(&models.User{}).Error; err != nil {
			return err
		}
		return tx.Create(user).Error
	}); err != nil {
		return err
	}
	return nil
}
