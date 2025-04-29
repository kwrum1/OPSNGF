# 简易下一代防火墙（Simple Next-Generation Firewall）

> 🚧 本项目正在积极开发中，AI 模块即将上线。

---

## 🌐 快速体验 / Quick Start

懒得做了 下次来体验把
---

## 🧩 核心架构 / Core Architecture


graph TD
    Client[Client] -->|HTTP Request| HAProxy
    HAProxy -->|TCP Connection| SPOE[Coraza SPOE Agent]
    SPOE -->|Message Type Recognition| TypeCheck
    TypeCheck -->|coraza-req| ReqHandler[Request Handler]
    TypeCheck -->|coraza-res| ResHandler[Response Handler]
    ReqHandler -->|Get App Name| ReqApp[Find Application]
    ResHandler -->|Get App Name| ResApp[Find Application]
    ReqApp -->|Process Request| ReqProcess[Request Processor]
    ResApp -->|Process Response| ResProcess[Response Processor]
    ReqProcess --> Return[Return Results to HAProxy]
    ResProcess --> Return
    HAProxy -->|Apply Action| Action[Allow/Deny/Log]
    Action -->|Response| Client


✨ 功能特性 / Features
🔐 安全防护 / Security Protection
支持 OWASP Core Rule Set (CRS)
支持Suricata 特性 / Suricata Features

兼容 ModSecurity SecLang 规则

自定义规则管理

HTTP 请求与响应检查

实时攻击检测与阻断​

⚙️ 系统管理 / System Management
HAProxy 生命周期管理（启动、停止、重启）

动态配置生成

实时状态监控

WAF 引擎管理​
IPS/IDS 规则控制管理

📊 监控与日志 / Monitoring & Logging
攻击日志与分析

流量统计

性能指标​

🔗 API 与认证 / API & Authentication
基于 Gin 的 RESTful API

Swagger / ReDoc 文档

JWT 身份验证


本地开发 / Local Development
前置条件 / Prerequisites
Go 1.24.1 或更高版本

Node.js 23.10.0 与 pnpm 10.6.5（用于前端开发）

HAProxy 3.0（用于本地开发）

MongoDB 6.0

Docker 与 Docker Compose（用于容器化部署）​

开发步骤 / Development Steps
克隆仓库：​

bash

git clone https://github.com/HUAHUAI23/simple-waf.git
cd simple-waf
设置前端开发环境：​


cd server/web
pnpm install
pnpm dev # 开发模式，支持热重载
# 或
pnpm build # 生产构建
cd ../..
配置后端环境：​

cp server/.env.template server/.env
# 编辑 .env 文件，根据需要修改配置
运行 Go 后端服务：​

go work use ./coraza-spoa ./pkg ./server
cd server
go run main.go
开发服务器将启动，访问地址：​

API 服务器：http://localhost:2333/api/v1

Swagger UI：http://localhost:2333/swagger/index.html

ReDoc UI：http://localhost:2333/redoc

前端页面：http://localhost:2333/​

Docker 部署 / Docker Deployment
克隆仓库：​

git clone https://github.com/HUAHUAI23/simple-waf.git
cd simple-waf
构建 Docker 镜像：​


docker build -t simple-waf:latest .
以独立容器运行：​


docker run -p 2333:2333 -p 8080:8080 -p 443:443 -p 80:80 -p 9443:9443 -p 8404:8404 simple-waf:latest
或使用 Docker Compose 进行完整部署（包含 MongoDB）：​


# 如有需要，编辑 docker-compose.yaml 配置环境变量
docker-compose up -d
这将启动 MongoDB 和简易下一代防火墙服务，包含所有必要配置。​

📄 许可证 / License
本项目基于 MIT 许可证开源。详情请参阅 LICENSE 文件。​

🙏 致谢 / Acknowledgements
OWASP Coraza WAF

Coraza SPOA

HAProxy

Go Gin Framework

Suricata
