package models

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// type Config struct {
// 	ID    int
// 	Key   string
// 	Value string
// }

// Config 配置结构体
type Config struct {
	JwtSecret  string `yaml:"jwt_secret"`  // JWT密钥
	ExpireDays int    `yaml:"expire_days"` // 过期天数
	Port       int    `yaml:"port"`        // 端口号
}

var comment string = `# jwt_secret: JWT密钥
# expire_days: token 过期天数
# port: 启动端口
`

// 初始化配置
func ConfigInit() error {
	if err := os.MkdirAll("./db", 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := restrictConfigDirectory("./db"); err != nil {
		return fmt.Errorf("secure config directory: %w", err)
	}
	cfg := Config{ExpireDays: 14, Port: 8000}
	if data, err := os.ReadFile("./db/config.yaml"); err == nil {
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("parse config file: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("read config file: %w", err)
	}
	if cfg.JwtSecret == "" {
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return fmt.Errorf("generate JWT secret: %w", err)
		}
		cfg.JwtSecret = hex.EncodeToString(secret)
	}
	if cfg.ExpireDays <= 0 {
		cfg.ExpireDays = 14
	}
	if cfg.Port <= 0 {
		cfg.Port = 8000
	}
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("serialize config: %w", err)
	}
	data = []byte(comment + string(data))
	return writeConfigFile(data)
}

// 读取配置
func ReadConfig() Config {
	file, err := os.ReadFile("./db/config.yaml")
	if err != nil {
		log.Println(err)
	}
	cfg := Config{ExpireDays: 14, Port: 8000}
	yaml.Unmarshal(file, &cfg)
	if cfg.ExpireDays <= 0 {
		cfg.ExpireDays = 14
	}
	if cfg.Port <= 0 {
		cfg.Port = 8000
	}
	return cfg
}

// 设置配置
func SetConfig(newCfg Config) error {
	oldCfg := ReadConfig() // 读取旧的配置文件
	// 覆盖新的字段
	if newCfg.JwtSecret != "" {
		oldCfg.JwtSecret = newCfg.JwtSecret
	}
	if newCfg.ExpireDays != 0 {
		oldCfg.ExpireDays = newCfg.ExpireDays
	}
	if newCfg.Port != 0 {
		oldCfg.Port = newCfg.Port
	}
	// 写入文件
	data, err := yaml.Marshal(&oldCfg)
	if err != nil {
		return fmt.Errorf("serialize config: %w", err)
	}
	data = []byte(comment + string(data)) // 添加注释
	return writeConfigFile(data)
}

func writeConfigFile(data []byte) error {
	file, err := os.OpenFile("./db/config.yaml", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	if err := restrictConfigFile(file.Name()); err != nil {
		_ = file.Close()
		return fmt.Errorf("restrict config file permissions: %w", err)
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return fmt.Errorf("write config file: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync config file: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close config file: %w", err)
	}
	return nil
}
