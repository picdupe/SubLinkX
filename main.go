package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"sublink/middlewares"
	"sublink/models"
	"sublink/routers"
	"sublink/settings"
	"sublink/utils"

	"github.com/gin-gonic/gin"
)

//go:embed static/js/*
//go:embed static/css/*
//go:embed static/img/*
//go:embed static/*
var embeddedFiles embed.FS

//go:embed template
var Template embed.FS

// 版本号
var version string

func Templateinit() {
	// 设置template路径
	// 检查目录是否创建
	subFS, err := fs.Sub(Template, "template")
	if err != nil {
		log.Println(err)
		return // 如果出错，直接返回
	}
	entries, err := fs.ReadDir(subFS, ".")
	if err != nil {
		log.Println(err)
		return // 如果出错，直接返回
	}
	// 创建template目录
	_, err = os.Stat("./template")
	if os.IsNotExist(err) {
		err = os.Mkdir("./template", 0666)
		if err != nil {
			log.Println(err)
			return
		}
	}
	// 写入默认模板
	for _, entry := range entries {
		_, err := os.Stat("./template/" + entry.Name())
		//如果文件不存在则写入默认模板
		if os.IsNotExist(err) {
			data, err := fs.ReadFile(subFS, entry.Name())
			if err != nil {
				log.Println(err)
				continue
			}
			err = os.WriteFile("./template/"+entry.Name(), data, 0666)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func main() {
	if err := loadDotEnv(".env"); err != nil {
		log.Fatalf("load .env: %v", err)
	}
	// 初始化配置
	if err := models.ConfigInit(); err != nil {
		log.Fatalf("initialize configuration: %v", err)
	}
	config := models.ReadConfig() // 读取配置文件
	var port = config.Port        // 读取端口号
	// 获取版本号
	var Isversion bool
	version = "2.1"
	flag.BoolVar(&Isversion, "version", false, "显示版本号")
	flag.Parse()
	if Isversion {
		fmt.Println(version)
		return
	}
	// 初始化数据库
	if err := models.InitSqlite(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	// 获取命令行参数
	args := os.Args
	// 如果长度小于2则没有接收到任何参数
	if len(args) < 2 {
		Run(port)
		return
	}
	// 命令行参数选择
	settingCmd := flag.NewFlagSet("setting", flag.ExitOnError)
	var username, password string
	settingCmd.StringVar(&username, "username", "", "设置账号")
	settingCmd.StringVar(&password, "password", "", "设置密码")
	settingCmd.IntVar(&port, "port", 8000, "修改端口")
	switch args[1] {
	// 解析setting命令标志
	case "setting":
		settingCmd.Parse(args[2:])
		if err := settings.ResetUser(username, password); err != nil {
			log.Fatalf("reset administrator account: %v", err)
		}
		return
	case "run":
		settingCmd.Parse(args[2:])
		if err := models.SetConfig(models.Config{
			Port: port,
		}); err != nil {
			log.Fatalf("save configuration: %v", err)
		}
		Run(port)
	default:
		return

	}
}

func Run(port int) {
	// 初始化gin框架
	r := gin.Default()
	r.RemoteIPHeaders = []string{"CF-Connecting-IP", "True-Client-IP", "X-Forwarded-For", "X-Real-IP"}
	trustedProxies := []string{}
	if configured := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES")); configured != "" {
		for _, proxy := range strings.Split(configured, ",") {
			if proxy = strings.TrimSpace(proxy); proxy != "" {
				trustedProxies = append(trustedProxies, proxy)
			}
		}
	}
	if err := r.SetTrustedProxies(trustedProxies); err != nil {
		log.Fatalf("invalid TRUSTED_PROXIES configuration: %v", err)
	}
	allowedOrigins, err := middlewares.ParseAllowedOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if err != nil {
		log.Fatalf("invalid CORS_ALLOWED_ORIGINS configuration: %v", err)
	}
	allowedClientIPs, err := middlewares.ParseAllowedClientIPs(os.Getenv("ALLOWED_CLIENT_IPS"))
	if err != nil {
		log.Fatalf("invalid ALLOWED_CLIENT_IPS configuration: %v", err)
	}
	r.Use(middlewares.CORS(allowedOrigins))
	r.Use(middlewares.ClientIPAllowlist(allowedClientIPs))
	// 初始化日志配置
	utils.Loginit()
	// 初始化模板
	Templateinit()
	// 安装中间件
	r.Use(middlewares.AuthorToken) // jwt验证token
	// 设置静态资源路径
	staticFiles, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Println(err)
	}
	r.StaticFS("/static", http.FS(staticFiles))
	// 设置模板路径
	r.GET("/", func(c *gin.Context) {
		data, err := fs.ReadFile(staticFiles, "index.html")
		if err != nil {
			c.Error(err)
			return

		}
		c.Data(200, "text/html", data)
	})
	// 注册路由
	routers.User(r)
	routers.Mentus(r)
	routers.Subcription(r)
	routers.Nodes(r)
	routers.Clients(r)
	routers.Total(r)
	routers.Templates(r)
	routers.Version(r, version)
	routers.Relay(r)
	// 启动服务
	r.Run(fmt.Sprintf("0.0.0.0:%d", port))
}

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		key, value, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" {
			return fmt.Errorf("invalid .env entry at line %d", lineNumber)
		}
		value = strings.TrimSpace(value)
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}
