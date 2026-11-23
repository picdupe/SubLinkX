package models

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sublink/utils"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB
var isInitialized bool

func InitSqlite() error {
	if isInitialized {
		return nil
	}
	if err := os.MkdirAll("./db", 0755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}
	db, err := gorm.Open(sqlite.Open("./db/sublink.db"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("open sqlite database: %w", err)
	}
	DB = db
	if err := db.AutoMigrate(
		&User{}, &SubLogs{}, &GroupNode{}, &Node{}, &RelayServer{}, &RelayRoute{},
		&Subcription{}, &RelayConfigRevision{}, &RelayConfigTask{}, &RelayAgentEvent{},
	); err != nil {
		return fmt.Errorf("migrate sqlite database: %w", err)
	}
	var existingUser User
	if err := db.First(&existingUser).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		initialPassword, err := utils.RandString(32)
		if err != nil {
			return fmt.Errorf("generate initial admin password: %w", err)
		}
		admin := &User{
			Username: "admin",
			Password: initialPassword,
			Role:     "admin",
			Nickname: "管理员",
		}
		if err := admin.Create(); err != nil {
			return fmt.Errorf("create initial admin user: %w", err)
		}
		log.Printf("已创建初始管理员账号 admin，请立即修改初始化密码：%s", initialPassword)
	} else if err != nil {
		return fmt.Errorf("read initial admin user: %w", err)
	}
	isInitialized = true
	log.Println("数据库初始化成功")
	return nil
}
