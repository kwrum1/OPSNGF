package test

import (
	"time"

	"github.com/HUAHUAI23/simple-waf/server/model"
)

// HAProxyStatus 表示HAProxy的运行状态
type HAProxyStatus struct {
	Running    bool      // 是否运行中
	Pid        int       // 进程ID
	Uptime     int64     // 运行时间(秒)
	StartTime  time.Time // 启动时间
	Version    string    // HAProxy版本
	ConfigFile string    // 配置文件路径
}

// FrontendStats 表示前端统计数据
type FrontendStats struct {
	Name               string // 前端名称
	Status             string // 状态(OPEN/CLOSED)
	ConnectionsTotal   int64  // 总连接数
	ConnectionsCurrent int64  // 当前连接数
	BytesIn            int64  // 入站流量(字节)
	BytesOut           int64  // 出站流量(字节)
	RequestRate        int64  // 请求速率(每秒)
	ConnectionRate     int64  // 连接速率(每秒)
}

// ServerStats 表示服务器统计数据
type ServerStats struct {
	Name               string // 服务器名称
	Status             string // 状态(UP/DOWN/MAINT)
	Weight             int    // 权重
	ConnectionsTotal   int64  // 总连接数
	ConnectionsCurrent int64  // 当前连接数
	BytesIn            int64  // 入站流量(字节)
	BytesOut           int64  // 出站流量(字节)
	ResponseTime       int64  // 响应时间(ms)
}

// BackendStats 表示后端统计数据
type BackendStats struct {
	Name               string                 // 后端名称
	Status             string                 // 状态(UP/DOWN)
	ConnectionsTotal   int64                  // 总连接数
	ConnectionsCurrent int64                  // 当前连接数
	BytesIn            int64                  // 入站流量(字节)
	BytesOut           int64                  // 出站流量(字节)
	Servers            map[string]ServerStats // 服务器统计
}

// HAProxyStats 表示HAProxy统计数据
type HAProxyStats struct {
	TotalConnections   int64                    // 总连接数
	CurrentConnections int64                    // 当前连接数
	BytesIn            int64                    // 入站流量(字节)
	BytesOut           int64                    // 出站流量(字节)
	ConnectionRate     int64                    // 连接速率(每秒)
	RequestRate        int64                    // 请求速率(每秒)
	Frontends          map[string]FrontendStats // 前端统计
	Backends           map[string]BackendStats  // 后端统计
}

// HAProxyService 定义了HAProxy服务的操作接口
type HAProxyService interface {
	// 基本操作
	RemoveConfig() error
	InitConfig() error
	Start() error
	Reload() error
	Stop() error
	AddSiteConfig(site model.Site) error

	// 状态和统计
	GetStatus() (*HAProxyStatus, error)
	GetStats() (*HAProxyStats, error)
	IsRunning() (bool, error)
	GetPid() (int, error)

	// 站点管理
	UpdateSiteConfig(site model.Site) error
	RemoveSiteConfig(domain string) error

	// 证书管理
	AddCertificate(cert model.Certificate) error
	RemoveCertificate(certName string) error

	// 后端和服务器管理
	GetBackends() ([]string, error)
	AddBackend(backend model.Backend) error
	RemoveBackend(backendName string) error
	AddServer(backendName string, server model.Server) error
	RemoveServer(backendName string, serverName string) error
	SetServerWeight(backendName string, serverName string, weight int) error
	SetServerState(backendName string, serverName string, enabled bool) error

	// WAF相关
	SetWAFMode(domain string, mode model.WAFMode) error
}
