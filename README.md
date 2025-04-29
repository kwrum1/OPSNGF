简易下一代防火墙（Simple Next-Generation Firewall）
项目简介
本项目是一个现代化的下一代防火墙（NGFW）管理系统，基于 HAProxy 和 OWASP Coraza WAF 构建，集成了 Coraza SPOA 进行流量处理。系统提供全面的后端 API，用于管理 HAProxy 配置、Coraza WAF 规则以及流量检测。​

🌐 一键部署
您可以在 30 秒内运行该应用，默认用户名：admin，默认密码：admin123​



核心架构
简易下一代防火墙采用模块化架构，前端由 HAProxy 处理流量，Coraza WAF 通过 SPOE（Stream Processing Offload Engine）提供安全检测：​

mermaid
复制
编辑
graph TD
    Client[客户端] -->|HTTP 请求| HAProxy
    HAProxy -->|TCP 连接| SPOE[Coraza SPOE Agent]
    SPOE -->|消息类型识别| TypeCheck
    TypeCheck -->|coraza-req| ReqHandler[请求处理器]
    TypeCheck -->|coraza-res| ResHandler[响应处理器]
    ReqHandler -->|获取应用名称| ReqApp[查找应用]
    ResHandler -->|获取应用名称| ResApp[查找应用]
    ReqApp -->|处理请求| ReqProcess[请求处理器]
    ResApp -->|处理响应| ResProcess[响应处理器]
    ReqProcess --> Return[返回结果给 HAProxy]
    ResProcess --> Return
    HAProxy -->|应用动作| Action[允许/拒绝/记录]
    Action -->|响应| Client
功能特性
HAProxy 集成

完整的 HAProxy 生命周期管理（启动、停止、重启）

动态配置生成

实时状态监控

Coraza WAF 集成

支持 OWASP Core Rule Set（CRS）

兼容 ModSecurity SecLang 规则

自定义规则管理

WAF 引擎生命周期管理

高级安全功能

HTTP 请求和响应检测

实时攻击检测与防御

基于角色的访问控制（RBAC）

监控与日志

WAF 攻击日志与分析

流量统计

性能指标

API 驱动的工作流

基于 Gin 框架的 RESTful API

Swagger/ReDoc API 文档

JWT 认证

系统要求
Go 1.24.1 或更高版本

Node.js 23.10.0 和 pnpm 10.6.5（用于前端开发）

HAProxy 3.0（用于本地开发）

MongoDB 6.0

Docker 和 Docker Compose（用于容器化部署）

本地开发
克隆仓库：

bash
复制
编辑
git clone https://github.com/kwrum1/waf.git
cd waf
设置前端开发环境：

bash
复制
编辑
cd server/web
pnpm install
pnpm dev # 开发模式，支持热重载
# 或
pnpm build # 生产构建
cd ../..
配置后端环境：

bash
复制
编辑
cp server/.env.template server/.env
# 编辑 .env 文件，配置您的环境变量
运行 Go 后端服务：

bash
复制
编辑
go work use ./coraza-spoa ./pkg ./server
cd server
go run main.go
开发服务器将启动，访问地址：

API 服务器：http://localhost:2333/api/v1

Swagger UI：http://localhost:2333/swagger/index.html

ReDoc UI：http://localhost:2333/redoc

前端：http://localhost:2333/

Docker 部署
克隆仓库：

bash
复制
编辑
git clone https://github.com/kwrum1/waf.git
cd waf
构建 Docker 镜像：

bash
复制
编辑
docker build -t simple-ngfw:latest .
作为独立容器运行：

bash
复制
编辑
docker run -p 2333:2333 -p 8080:8080 -p 443:443 -p 80:80 -p 9443:9443 -p 8404:8404 simple-ngfw:latest
或者，使用 Docker Compose 进行完整部署，包括 MongoDB：

bash
复制
编辑
# 如有需要，编辑 docker-compose.yaml 文件，配置环境变量
docker-compose up -d
这将启动 MongoDB 和简易下一代防火墙服务，并进行所有必要的配置。

许可证
本项目基于 MIT 许可证，详见 LICENSE 文件。

鸣谢
OWASP Coraza WAF

Coraza SPOA

HAProxy

Go Gin 框架
