package dto

import (
	"time"
)

// ConfigPatchRequest 配置补丁更新请求
// @Description 用于部分更新配置的请求参数
type ConfigPatchRequest struct {
	Name            *string          `json:"name,omitempty" binding:"omitempty" example:"AppConfig"`        // 配置名称
	Engine          *EnginePatchDTO  `json:"engine,omitempty" binding:"omitempty"`                          // 引擎配置
	Haproxy         *HaproxyPatchDTO `json:"haproxy,omitempty" binding:"omitempty"`                         // HAProxy配置
	IsResponseCheck *bool            `json:"isResponseCheck,omitempty" binding:"omitempty" example:"false"` // 是否检查响应
	IsDebug         *bool            `json:"isDebug,omitempty" binding:"omitempty" example:"false"`         // 是否开启调试模式
}

// EnginePatchDTO 引擎配置补丁DTO
type EnginePatchDTO struct {
	Bind            *string             `json:"bind,omitempty" binding:"omitempty" example:"127.0.0.1:2342"`  // 引擎绑定地址
	UseBuiltinRules *bool               `json:"useBuiltinRules,omitempty" binding:"omitempty" example:"true"` // 是否使用内置规则
	AppConfig       []AppConfigPatchDTO `json:"appConfig,omitempty" binding:"omitempty,dive"`                 // 应用配置列表
}

// AppConfigPatchDTO 应用配置补丁DTO
type AppConfigPatchDTO struct {
	Name           *string `json:"name,omitempty" binding:"omitempty" example:"coraza"`          // 应用名称
	Directives     *string `json:"directives,omitempty" binding:"omitempty"`                     // 指令配置
	TransactionTTL *int64  `json:"transactionTTL,omitempty" binding:"omitempty" example:"60000"` // 事务超时时间(毫秒)
	LogLevel       *string `json:"logLevel,omitempty" binding:"omitempty" example:"info"`        // 日志级别
	LogFile        *string `json:"logFile,omitempty" binding:"omitempty" example:"/dev/stdout"`  // 日志文件
	LogFormat      *string `json:"logFormat,omitempty" binding:"omitempty" example:"console"`    // 日志格式
}

// HaproxyPatchDTO HAProxy配置补丁DTO
type HaproxyPatchDTO struct {
	ConfigBaseDir *string `json:"configBaseDir,omitempty" binding:"omitempty" example:"/simple-waf"` // 配置文件根目录
	HaproxyBin    *string `json:"haproxyBin,omitempty" binding:"omitempty" example:"haproxy"`        // HAProxy二进制文件路径
	BackupsNumber *int    `json:"backupsNumber,omitempty" binding:"omitempty" example:"5"`           // 备份数量
	SpoeAgentAddr *string `json:"spoeAgentAddr,omitempty" binding:"omitempty" example:"127.0.0.1"`   // SPOE代理地址
	SpoeAgentPort *int    `json:"spoeAgentPort,omitempty" binding:"omitempty" example:"2342"`        // SPOE代理端口
	Thread        *int    `json:"thread,omitempty" binding:"omitempty,min=0,max=256" example:"4"`    // 线程数
}

// ConfigResponse 配置响应
// @Description 配置响应
type ConfigResponse struct {
	ID              string     `json:"id,omitempty"`    // 配置ID
	Name            string     `json:"name"`            // 配置名称
	Engine          EngineDTO  `json:"engine"`          // 引擎配置
	Haproxy         HaproxyDTO `json:"haproxy"`         // HAProxy配置
	CreatedAt       time.Time  `json:"createdAt"`       // 创建时间
	UpdatedAt       time.Time  `json:"updatedAt"`       // 更新时间
	IsResponseCheck bool       `json:"isResponseCheck"` // 是否检查响应
	IsDebug         bool       `json:"isDebug"`         // 是否开启调试模式
}

// EngineDTO 引擎配置DTO
type EngineDTO struct {
	Bind            string         `json:"bind"`            // 引擎绑定地址
	UseBuiltinRules bool           `json:"useBuiltinRules"` // 是否使用内置规则
	AppConfig       []AppConfigDTO `json:"appConfig"`       // 应用配置列表
}

// AppConfigDTO 应用配置DTO
type AppConfigDTO struct {
	Name           string `json:"name"`                           // 应用名称
	Directives     string `json:"directives"`                     // 指令配置
	TransactionTTL int64  `json:"transactionTTL" example:"60000"` // 事务超时时间(毫秒)
	LogLevel       string `json:"logLevel"`                       // 日志级别
	LogFile        string `json:"logFile"`                        // 日志文件
	LogFormat      string `json:"logFormat"`                      // 日志格式
}

// HaproxyDTO HAProxy配置DTO
type HaproxyDTO struct {
	ConfigBaseDir string `json:"configBaseDir"` // 配置文件根目录
	HaproxyBin    string `json:"haproxyBin"`    // HAProxy二进制文件路径
	BackupsNumber int    `json:"backupsNumber"` // 备份数量
	SpoeAgentAddr string `json:"spoeAgentAddr"` // SPOE代理地址
	SpoeAgentPort int    `json:"spoeAgentPort"` // SPOE代理端口
	Thread        int    `json:"thread"`        // 线程数
}

// 将 time.Duration 转换为毫秒表示的 int64
func DurationToMillis(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

// 将毫秒表示的 int64 转换为 time.Duration
func MillisToDuration(millis int64) time.Duration {
	return time.Duration(millis) * time.Millisecond
}
