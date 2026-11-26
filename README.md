<div align="center">
<img src="webs/src/assets/logo.png" width="150px" height="150px" />
</div>

<div align="center">
    <img src="https://img.shields.io/badge/Vue-5.0.8-brightgreen.svg"/>
    <img src="https://img.shields.io/badge/Go-1.22.0-green.svg"/>
    <img src="https://img.shields.io/badge/Element Plus-2.6.1-blue.svg"/>
    <img src="https://img.shields.io/badge/license-MIT-green.svg"/>
    <a href="https://t.me/+u6gLWF0yP5NiZWQ1" target="_blank">
        <img src="https://img.shields.io/badge/TG-交流群-orange.svg"/>
    </a>
    <div align="center"> 中文 | <a href="README.en-US.md">English</div>
</div>

## [项目简介]

项目基于sublink项目二次开发：https://github.com/jaaksii/sublink

前端基于：https://github.com/youlaitech/vue3-element-admin

后端采用go+gin+gorm

默认账号admin 密码123456  自行修改

因为重写目前还有很多布局结构以及功能稍少

## [项目特色]

自由度和安全性较高，能够记录访问订阅，配置轻松

二进制编译无需Docker容器

目前仅支持客户端：v2ray clash surge

v2ray为base64通用格式

clash支持协议:ss ssr trojan vmess vless hy hy2 tuic

surge支持协议:ss trojan vmess hy2 tuic

## [项目预览]

![1712594176714](webs/src/assets/1.png)
![1712594176714](webs/src/assets/2.png)

## [2.1更新说明]

#### 后端更新

1. 修复底层代码
2. 修复各种奇葩bug
3. 建议卸载数据库(记得备份数据) 新数据库结构有些不一样可能会导致一些bug

#### 前端更新

1. 完善node页面




## [安装说明]
### linux方式：
```
curl -s -H "Cache-Control: no-cache" -H "Pragma: no-cache" https://raw.githubusercontent.com/gooaclok819/sublinkX/main/install.sh | sudo bash
```

```sublink``` 呼出菜单

然后输入安装脚本即可

### docker方式：

在项目目录中复制环境变量模板并配置允许访问面板的来源：

```
cp .env.example .env
```

Docker 容器默认以 UID/GID `10001` 运行。首次启动前先准备持久化目录并设置属主；宿主机使用其他 UID/GID 时，也要在 `.env` 中修改 `SUBLINK_UID` 和 `SUBLINK_GID`，并使用对应值设置目录属主：

```sh
mkdir -p db logs template
sudo chown -R 10001:10001 db logs template
```

编辑 `.env` 后，用 Compose 构建并启动：

```
docker compose up -d --build
```

Compose 会持久化 `db`、`template` 和 `logs` 目录。使用 `docker compose ps` 查看运行状态，使用 `docker compose logs -f sublink` 查看日志。

跨域来源和客户端来源 IP 通过项目根目录的 `.env` 配置。先复制 `.env.example` 为 `.env`，再分别配置 `CORS_ALLOWED_ORIGINS` 和 `ALLOWED_CLIENT_IPS`，多个值用逗号分隔，例如：

```env
CORS_ALLOWED_ORIGINS=https://panel.example.com,http://192.168.1.50:8081
ALLOWED_CLIENT_IPS=192.168.1.0/24,203.0.113.25
```

`CORS_ALLOWED_ORIGINS` 需包含协议和端口（HTTP 默认端口 80、HTTPS 默认端口 443 可省略），按浏览器来源域名/IP 精确匹配，不支持通配符。它仅是浏览器跨域策略，不是网络访问控制。`ALLOWED_CLIENT_IPS` 限制所有 HTTP 客户端的实际来源 IP，支持单个 IPv4/IPv6 地址及 CIDR；留空时仅允许本机回环访问。经反向代理访问时，`TRUSTED_PROXIES` 必须设置为应用看到的反代 IP/CIDR，否则服务端只会记录代理 IP。例如日志显示 `192.168.16.1` 时，将它加入 `TRUSTED_PROXIES`。应用在可信代理请求中识别 `CF-Connecting-IP`、`True-Client-IP`、`X-Forwarded-For` 和 `X-Real-IP`。若 CDN 直接连接应用，将 CDN 官方 IP 段加入 `TRUSTED_PROXIES`；若 CDN 经 Nginx 等反代连接，只需信任应用直接连接的反代，并确保反代传递或覆盖 CDN 提供的客户端 IP 头。不要信任任意来源转发的 IP 头；并将客户端实际来源网段配置到 `ALLOWED_CLIENT_IPS`。修改后重启应用容器使配置生效。

直接运行已发布镜像时，先准备 `.env` 和持久化目录，再执行：

```
docker run --name sublinkx --user 10001:10001 -p 8000:8000 --env-file .env \
  -v $PWD/db:/app/db -v $PWD/template:/app/template -v $PWD/logs:/app/logs \
  -d jaaksi/sublinkx
```

To support the development of my project, I plan to apply for a free VPS offered by ZMTO. My project currently involves Docker image support for multiple architectures (arm64 and amd64), as well as automation for building and pushing. Therefore, I am requesting a 4-core, 8GB RAM Ubuntu VPS with root access.

Thank you to the ZMTO team for your support. I look forward to leveraging this VPS to optimize my project's performance and development efficiency. If you have any questions or suggestions regarding my project, feel free to open an issue, and I will do my best to improve and optimize it.

Thank you for your attention and support!

Feel free to adjust any details as needed!

## Stargazers over time
[![Stargazers over time](https://starchart.cc/gooaclok819/sublinkX.svg?variant=adaptive)](https://starchart.cc/gooaclok819/sublinkX)

