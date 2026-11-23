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
    <div align="center"> <a href="README.md">中文<a> | English</div>
</div>
## [Project Information]

Project based on sublink project secondary development: https://github.com/jaaksii/sublink

Front-end based on: https://github.com/youlaitech/vue3-element-admin

Backend using go+gin+gorm

Default account admin password 123456 self-modification

Because of the rewrite there are still a lot of layout structure and a little less functionality

## [Project Features]

High degree of freedom and security, the ability to record access to the subscription, easy configuration

Binary compilation without Docker container.

Currently only supports the client: v2ray clash surge

v2ray is a base64 universal format

clash supported protocols: ss ssr trojan vmess vless hy hy2 tuic

surge support protocol:ss trojan vmess hy2 tuic

## [Project Preview]

![1712594176714](webs/src/assets/1.png)
![1712594176714](webs/src/assets/2.png)

## [Updated Description]

####Backend Update

1. Fix and refactor a large number of Node templates and the underlying code for new groupings

2. Add grouping functionality to nodes

3. Fix bug that subscription resolution is empty or etc

####Front-end update

1. Refactor front-end node page to add grouping function (temporarily only some simple functions)

## [Installation instructions]
### linux method:
```
curl -s -H “Cache-Control: no-cache” -H “Pragma: no-cache” https://raw.githubusercontent.com/gooaclok819/sublinkX/main/install.sh | sudo bash
```

```sublink``` Calls out the menu.

Then just type in the install script

### docker method:

From the project directory, copy the environment template and configure the allowed panel origins:

```
cp .env.example .env
```

Then build and start the app with Docker Compose:

```
docker compose up -d --build
```

Compose persists the `db`, `template`, and `logs` directories. Check the service with `docker compose ps` and follow its logs with `docker compose logs -f sublink`.

Set `CORS_ALLOWED_ORIGINS` to the exact origins allowed to make browser cross-origin requests. Separate multiple origins with commas:

```env
CORS_ALLOWED_ORIGINS=https://panel.example.com,http://192.168.1.50:8081
```

Include the scheme and non-default port. List each domain or IP separately; wildcards are not supported. Every request with an `Origin` header, including requests from the panel, must match the list. This is a browser CORS policy: it does not restrict clients that omit `Origin` and does not replace firewall or reverse-proxy source-IP rules. Configure network source restrictions at the firewall or reverse proxy. Apply changes with `docker compose up -d`.

To run a published image directly on Linux, prepare `.env` and make the persistent directories writable by the container's default UID/GID `10001:10001` first:

```sh
mkdir -p db template logs
sudo chown -R 10001:10001 db template logs
```

Then run:

```
docker run --name sublinkx -p 8000:8000 --env-file .env \
  -v $PWD/db:/app/db -v $PWD/template:/app/template -v $PWD/logs:/app/logs \
  -d jaaksi/sublinkx
```

To support the development of my project, I plan to apply for a free VPS offered by ZMTO. My project currently involves Docker image support for multiple My project currently involves Docker image support for multiple architectures (arm64 and amd64), as well as automation for building and pushing. Therefore, I am requesting a 4-core, 8GB RAM Ubuntu VPS with root access.

Thank you to the ZTMO team for your support. I look forward to leveraging this VPS to optimize my project's performance and development efficiency. have any questions or suggestions regarding my project, feel free to open an issue, and I will do my best to improve and optimize it.

Thank you for your attention and support!

Feel free to adjust any details as needed!

## Stargazers over time
[![Stargazers over time](https://starchart.cc/gooaclok819/sublinkX.svg?variant=adaptive)](https://starchart.cc/gooaclok819/sublinkX)

