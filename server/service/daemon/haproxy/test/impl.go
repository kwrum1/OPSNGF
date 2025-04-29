package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/HUAHUAI23/simple-waf/server/model"
	client_native "github.com/haproxytech/client-native/v6"
	"github.com/haproxytech/client-native/v6/configuration"
	cfg_opt "github.com/haproxytech/client-native/v6/configuration/options"
	"github.com/haproxytech/client-native/v6/models"
	"github.com/haproxytech/client-native/v6/options"
	runtime_api "github.com/haproxytech/client-native/v6/runtime"
	runtime_options "github.com/haproxytech/client-native/v6/runtime/options"
	spoe "github.com/haproxytech/client-native/v6/spoe"
)

// HAProxyServiceImpl 实现了HAProxyService接口
type HAProxyServiceImpl struct {
	ConfigBaseDir      string // 配置基础目录
	ConfigFile         string // 配置文件路径
	HaproxyBin         string // HAProxy二进制文件路径
	BackupsNumber      int    // 备份数量
	TransactionDir     string // 事务目录
	SpoeDir            string // SPOE目录
	SpoeTransactionDir string // SPOE事务目录
	SocketDir          string // 套接字目录
	PidFile            string // PID文件路径

	// 内部字段
	haproxyCmd    *exec.Cmd                   // HAProxy进程命令
	confClient    configuration.Configuration // 配置客户端
	runtimeClient runtime_api.Runtime         // 运行时客户端
	spoeClient    spoe.Spoe                   // SPOE客户端
	clientNative  client_native.HAProxyClient // 完整客户端
	ctx           context.Context             // 上下文
	mutex         sync.Mutex                  // 互斥锁
	isInitialized bool                        // 是否已初始化
}

// NewHAProxyService 创建一个新的HAProxy服务实例
func NewHAProxyService(configBaseDir, haproxyBin string) (*HAProxyServiceImpl, error) {
	// 获取用户主目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户主目录: %v", err)
	}

	// 如果未指定配置目录，使用默认目录
	if configBaseDir == "" {
		configBaseDir = filepath.Join(homeDir, "haproxy")
	}

	// 如果未指定二进制路径，假设在PATH中可用
	if haproxyBin == "" {
		haproxyBin = "haproxy"
	}

	return &HAProxyServiceImpl{
		ConfigBaseDir:      configBaseDir,
		ConfigFile:         filepath.Join(configBaseDir, "conf/haproxy.cfg"),
		HaproxyBin:         haproxyBin,
		BackupsNumber:      0,
		TransactionDir:     filepath.Join(configBaseDir, "conf/transactions"),
		SpoeDir:            filepath.Join(configBaseDir, "spoe"),
		SpoeTransactionDir: filepath.Join(configBaseDir, "spoe/transaction"),
		SocketDir:          filepath.Join(configBaseDir, "conf/sock"),
		PidFile:            filepath.Join(configBaseDir, "conf/haproxy.pid"),
		ctx:                context.Background(),
		isInitialized:      false,
	}, nil
}

// RemoveConfig 删除HAProxy配置
func (s *HAProxyServiceImpl) RemoveConfig() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 先停止HAProxy
	if err := s.stopHAProxy(); err != nil {
		log.Printf("停止HAProxy时出错: %v", err)
		// 继续执行，尝试删除配置
	}

	// 删除配置目录
	if err := os.RemoveAll(s.ConfigBaseDir); err != nil {
		return fmt.Errorf("删除配置目录失败: %v", err)
	}

	s.isInitialized = false
	return nil
}

// InitConfig 初始化HAProxy配置
func (s *HAProxyServiceImpl) InitConfig() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 创建必要的目录
	dirs := []string{
		filepath.Dir(s.ConfigFile),
		s.TransactionDir,
		s.SocketDir,
		filepath.Dir(s.PidFile),
		s.SpoeDir,
		s.SpoeTransactionDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("无法创建目录 %s: %v", dir, err)
		}
	}

	// 如果配置文件不存在，创建基本配置
	if _, err := os.Stat(s.ConfigFile); os.IsNotExist(err) {
		username := os.Getenv("USER")
		if username == "" {
			username = "haproxy"
		}

		basicConfig := fmt.Sprintf(`# _version = 1
global
    log stdout format raw local0
    maxconn 4000
    # user %s
    # group %s
defaults
    log global
    mode http
    option httplog
    timeout client 1m
    timeout server 1m
    timeout connect 10s
# 以下部分将由程序动态配置
`, username, username)

		if err := os.WriteFile(s.ConfigFile, []byte(basicConfig), 0644); err != nil {
			return fmt.Errorf("无法创建基本配置文件: %v", err)
		}
	}

	s.isInitialized = true
	return nil
}

// Start 启动HAProxy
func (s *HAProxyServiceImpl) Start() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查是否已初始化
	if !s.isInitialized {
		if err := s.InitConfig(); err != nil {
			return fmt.Errorf("初始化配置失败: %v", err)
		}
	}

	// 检查HAProxy是否已经在运行
	running, _ := s.isHAProxyRunning()
	if running {
		return fmt.Errorf("HAProxy已经在运行")
	}

	// 启动HAProxy进程
	cmd := exec.Command(
		s.HaproxyBin,
		"-f", s.ConfigFile,
		"-p", s.PidFile,
		"-Ws",
		"-S", fmt.Sprintf("unix@%s", filepath.Join(s.SocketDir, "haproxy-master.sock")),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动HAProxy失败: %v", err)
	}

	s.haproxyCmd = cmd

	// 等待套接字文件创建
	socketPath := filepath.Join(s.SocketDir, "haproxy-master.sock")
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		if _, err := os.Stat(socketPath); err == nil {
			// 套接字已创建，继续等待一段时间确保HAProxy就绪
			time.Sleep(500 * time.Millisecond)
			break
		}

		if i == maxAttempts-1 {
			return fmt.Errorf("套接字文件未创建: %s", socketPath)
		}

		time.Sleep(500 * time.Millisecond)
	}

	// 初始化客户端
	if err := s.initClients(); err != nil {
		// 如果初始化客户端失败，尝试终止HAProxy进程
		s.stopHAProxy()
		return fmt.Errorf("初始化客户端失败: %v", err)
	}

	return nil
}

// Reload 重新加载HAProxy配置
func (s *HAProxyServiceImpl) Reload() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查运行时客户端是否已初始化
	if s.runtimeClient == nil {
		return fmt.Errorf("HAProxy未运行或运行时客户端未初始化")
	}

	// 重新加载HAProxy
	output, err := s.runtimeClient.Reload()
	if err != nil {
		return fmt.Errorf("重新加载HAProxy失败: %v", err)
	}

	log.Printf("HAProxy已重新加载: %s", output)
	return nil
}

// Stop 停止HAProxy
func (s *HAProxyServiceImpl) Stop() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.stopHAProxy()
}

// GetStatus 获取HAProxy状态
func (s *HAProxyServiceImpl) GetStatus() (*HAProxyStatus, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	status := &HAProxyStatus{
		ConfigFile: s.ConfigFile,
	}

	// 检查HAProxy是否运行
	running, err := s.isHAProxyRunning()
	if err != nil {
		return nil, fmt.Errorf("检查HAProxy运行状态时出错: %v", err)
	}

	status.Running = running

	if running {
		// 获取PID
		pid, err := s.getHAProxyPid()
		if err == nil {
			status.Pid = pid
		}

		// 如果运行时客户端已初始化，获取更多信息
		if s.runtimeClient != nil {
			info, err := s.runtimeClient.GetInfo()
			if err == nil && info.Info != nil {
				status.Version = info.Info.Version

				if info.Info.Uptime != nil {
					status.Uptime = *info.Info.Uptime
					status.StartTime = time.Now().Add(-time.Duration(status.Uptime) * time.Second)
				}
			}
		}
	}

	return status, nil
}

// GetStats 获取HAProxy统计数据
func (s *HAProxyServiceImpl) GetStats() (*HAProxyStats, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.runtimeClient == nil {
		return nil, fmt.Errorf("HAProxy未运行或运行时客户端未初始化")
	}

	// 初始化统计数据
	stats := &HAProxyStats{
		Frontends: make(map[string]FrontendStats),
		Backends:  make(map[string]BackendStats),
	}

	// 获取所有统计信息
	runtimeStats, err := s.runtimeClient.GetStats()
	if err != nil {
		return nil, fmt.Errorf("获取统计数据失败: %v", err)
	}

	// 解析统计数据
	for _, stat := range runtimeStats {
		switch stat.Type {
		case "frontend":
			feStat := FrontendStats{
				Name:   stat.Name,
				Status: s.getSafeString(stat.Status),
			}

			if stat.ConnectionsTotal != nil {
				feStat.ConnectionsTotal = *stat.ConnectionsTotal
				stats.TotalConnections += feStat.ConnectionsTotal
			}

			if stat.ConnectionsCurrent != nil {
				feStat.ConnectionsCurrent = *stat.ConnectionsCurrent
				stats.CurrentConnections += feStat.ConnectionsCurrent
			}

			if stat.BytesIn != nil {
				feStat.BytesIn = *stat.BytesIn
				stats.BytesIn += feStat.BytesIn
			}

			if stat.BytesOut != nil {
				feStat.BytesOut = *stat.BytesOut
				stats.BytesOut += feStat.BytesOut
			}

			if stat.Rate != nil {
				feStat.RequestRate = *stat.Rate
				stats.RequestRate += feStat.RequestRate
			}

			if stat.ConnectionRate != nil {
				feStat.ConnectionRate = *stat.ConnectionRate
				stats.ConnectionRate += feStat.ConnectionRate
			}

			stats.Frontends[stat.Name] = feStat

		case "backend":
			beStat := BackendStats{
				Name:               stat.Name,
				Status:             s.getSafeString(stat.Status),
				ConnectionsTotal:   s.getSafeInt64(stat.ConnectionsTotal),
				ConnectionsCurrent: s.getSafeInt64(stat.ConnectionsCurrent),
				BytesIn:            s.getSafeInt64(stat.BytesIn),
				BytesOut:           s.getSafeInt64(stat.BytesOut),
				Servers:            make(map[string]ServerStats),
			}

			stats.Backends[stat.Name] = beStat

		case "server":
			if stat.BackendName != nil {
				serverStat := ServerStats{
					Name:               stat.Name,
					Status:             s.getSafeString(stat.Status),
					Weight:             int(s.getSafeInt64(stat.Weight)),
					ConnectionsTotal:   s.getSafeInt64(stat.ConnectionsTotal),
					ConnectionsCurrent: s.getSafeInt64(stat.ConnectionsCurrent),
					BytesIn:            s.getSafeInt64(stat.BytesIn),
					BytesOut:           s.getSafeInt64(stat.BytesOut),
					ResponseTime:       s.getSafeInt64(stat.ResponseTime),
				}

				if backend, ok := stats.Backends[*stat.BackendName]; ok {
					backend.Servers[stat.Name] = serverStat
					stats.Backends[*stat.BackendName] = backend
				}
			}
		}
	}

	return stats, nil
}

// IsRunning 检查HAProxy是否运行
func (s *HAProxyServiceImpl) IsRunning() (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.isHAProxyRunning()
}

// GetPid 获取HAProxy进程ID
func (s *HAProxyServiceImpl) GetPid() (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.getHAProxyPid()
}

// 添加其他方法实现...

// AddSiteConfig 添加站点配置
func (s *HAProxyServiceImpl) AddSiteConfig(site model.Site) error {
	// 验证站点配置
	if err := model.ValidateSite(&site); err != nil {
		return fmt.Errorf("站点配置无效: %v", err)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}

	// 获取版本并开始事务
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}

	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 配置前端
	frontendName := "fe_" + site.Domain
	frontend := &models.Frontend{
		FrontendBase: models.FrontendBase{
			Name:           frontendName,
			Mode:           "http",
			DefaultBackend: site.Backend.Name,
			Enabled:        true,
		},
	}

	if err := s.confClient.CreateFrontend(frontend, transaction.ID, 0); err != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("创建前端失败: %v", err)
	}

	// 配置绑定
	bind := &models.Bind{
		BindParams: models.BindParams{
			Name: "bind_" + site.Domain,
		},
		Address: "*",
		Port:    s.Int64P(int64(site.ListenPort)),
	}

	// 如果启用了HTTPS，配置SSL
	if site.EnableHTTPS && site.Certificate.CertName != "" {
		bind.Ssl = true
		bind.SslCertificate = site.Certificate.CertName
	}

	if err := s.confClient.CreateBind("frontend", frontendName, bind, transaction.ID, 0); err != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("创建绑定失败: %v", err)
	}

	// 配置后端
	backend := &models.Backend{
		BackendBase: models.BackendBase{
			Name:    site.Backend.Name,
			Mode:    "http",
			Enabled: true,
		},
	}

	if err := s.confClient.CreateBackend(backend, transaction.ID, 0); err != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("创建后端失败: %v", err)
	}

	// 配置服务器
	// 继续 AddSiteConfig 方法中配置服务器部分
	// 配置服务器
	for _, server := range site.Backend.Servers {
		serverModel := &models.Server{
			Name:    server.Name,
			Address: server.Host,
			Port:    s.Int64P(int64(server.Port)),
			ServerParams: models.ServerParams{
				Weight: s.Int64P(int64(server.Weight)),
				Check:  "enabled",
			},
		}

		if err := s.confClient.CreateServer("backend", site.Backend.Name, serverModel, transaction.ID, 0); err != nil {
			s.confClient.DeleteTransaction(transaction.ID)
			return fmt.Errorf("创建服务器 %s 失败: %v", server.Name, err)
		}
	}

	// 如果启用了WAF，配置WAF规则
	if site.WAFEnabled {
		// 添加SPOE过滤器
		spoeFilter := &models.Filter{
			Type:       "spoe",
			SpoeEngine: "coraza",
			SpoeConfig: filepath.Join(s.SpoeDir, "coraza-spoa.yaml"),
		}

		if err := s.confClient.CreateFilter(0, "frontend", frontendName, spoeFilter, transaction.ID, 0); err != nil {
			s.confClient.DeleteTransaction(transaction.ID)
			return fmt.Errorf("创建SPOE过滤器失败: %v", err)
		}

		// 添加WAF模式的HTTP变量
		wafModeVar := &models.HTTPRequestRule{
			Type:      "set-var",
			VarName:   "txn.waf_mode",
			VarFormat: string(site.WAFMode),
		}

		if err := s.confClient.CreateHTTPRequestRule(0, "frontend", frontendName, wafModeVar, transaction.ID, 0); err != nil {
			s.confClient.DeleteTransaction(transaction.ID)
			return fmt.Errorf("设置WAF模式变量失败: %v", err)
		}

		// 根据WAF模式添加相应的防护规则
		if site.WAFMode == model.WAFModeProtection {
			// 添加拦截规则
			httpReqRules := []struct {
				index int64
				rule  *models.HTTPRequestRule
			}{
				{1, &models.HTTPRequestRule{
					Type:       "deny",
					DenyStatus: s.Int64P(403),
					HdrFormat:  "WAF Protection",
					Cond:       "if",
					CondTest:   "{ var(txn.coraza.action) -m str deny }",
				}},
				{2, &models.HTTPRequestRule{
					Type:     "silent-drop",
					Cond:     "if",
					CondTest: "{ var(txn.coraza.action) -m str drop }",
				}},
				{3, &models.HTTPRequestRule{
					Type:       "redirect",
					RedirCode:  s.Int64P(302),
					RedirType:  "location",
					RedirValue: "%[var(txn.coraza.data)]",
					Cond:       "if",
					CondTest:   "{ var(txn.coraza.action) -m str redirect }",
				}},
			}

			for _, item := range httpReqRules {
				if err := s.confClient.CreateHTTPRequestRule(item.index, "frontend", frontendName, item.rule, transaction.ID, 0); err != nil {
					s.confClient.DeleteTransaction(transaction.ID)
					return fmt.Errorf("创建HTTP请求规则失败: %v", err)
				}
			}
		}
	}

	// 提交事务
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 如果HAProxy已经运行，重新加载配置
	running, _ := s.isHAProxyRunning()
	if running && s.runtimeClient != nil {
		if _, err := s.runtimeClient.Reload(); err != nil {
			return fmt.Errorf("重新加载HAProxy配置失败: %v", err)
		}
	}

	return nil
}

// UpdateSiteConfig 更新站点配置
func (s *HAProxyServiceImpl) UpdateSiteConfig(site model.Site) error {
	// 先删除站点，然后重新添加
	if err := s.RemoveSiteConfig(site.Domain); err != nil {
		return fmt.Errorf("删除旧站点配置失败: %v", err)
	}

	return s.AddSiteConfig(site)
}

// RemoveSiteConfig 删除站点配置
func (s *HAProxyServiceImpl) RemoveSiteConfig(domain string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}

	// 获取版本并开始事务
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}

	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	frontendName := "fe_" + domain

	// 检查前端是否存在
	_, frontendErr := s.confClient.GetFrontend(frontendName, transaction.ID)
	if frontendErr == nil {
		// 前端存在，删除它
		if err := s.confClient.DeleteFrontend(frontendName, transaction.ID, 0); err != nil {
			s.confClient.DeleteTransaction(transaction.ID)
			return fmt.Errorf("删除前端失败: %v", err)
		}
	}

	// 提交事务
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 如果HAProxy已经运行，重新加载配置
	running, _ := s.isHAProxyRunning()
	if running && s.runtimeClient != nil {
		if _, err := s.runtimeClient.Reload(); err != nil {
			return fmt.Errorf("重新加载HAProxy配置失败: %v", err)
		}
	}

	return nil
}

// AddCertificate 添加证书
func (s *HAProxyServiceImpl) AddCertificate(cert model.Certificate) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}

	// 在本地文件系统创建证书文件
	certDir := filepath.Join(s.ConfigBaseDir, "certs")
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("创建证书目录失败: %v", err)
	}

	// 保存公钥
	certPath := filepath.Join(certDir, cert.CertName+".pem")
	if err := os.WriteFile(certPath, []byte(cert.PublicKey), 0644); err != nil {
		return fmt.Errorf("保存证书公钥失败: %v", err)
	}

	// 保存私钥
	keyPath := filepath.Join(certDir, cert.CertName+".key")
	if err := os.WriteFile(keyPath, []byte(cert.PrivateKey), 0600); err != nil {
		return fmt.Errorf("保存证书私钥失败: %v", err)
	}

	// 获取版本并开始事务
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}

	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 检查证书存储是否存在
	crtStore := &models.CrtStore{
		Name:    "sites",
		CrtBase: certDir,
		KeyBase: certDir,
	}

	_, crtStoreErr := s.confClient.GetCrtStore("sites", transaction.ID)
	if crtStoreErr != nil {
		// 证书存储不存在，创建它
		if err := s.confClient.CreateCrtStore(crtStore, transaction.ID, 0); err != nil {
			s.confClient.DeleteTransaction(transaction.ID)
			return fmt.Errorf("创建证书存储失败: %v", err)
		}
	}

	// 加载证书
	crtLoad := &models.CrtLoad{
		Certificate: cert.CertName + ".pem",
		Key:         cert.CertName + ".key",
		Alias:       cert.CertName,
	}

	if err := s.confClient.CreateCrtLoad("sites", crtLoad, transaction.ID, 0); err != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("加载证书失败: %v", err)
	}

	// 提交事务
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 如果HAProxy已经运行，重新加载配置
	running, _ := s.isHAProxyRunning()
	if running && s.runtimeClient != nil {
		if _, err := s.runtimeClient.Reload(); err != nil {
			return fmt.Errorf("重新加载HAProxy配置失败: %v", err)
		}
	}

	return nil
}

// RemoveCertificate 删除证书
func (s *HAProxyServiceImpl) RemoveCertificate(certName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}

	// 删除本地证书文件
	certDir := filepath.Join(s.ConfigBaseDir, "certs")
	certPath := filepath.Join(certDir, certName+".pem")
	keyPath := filepath.Join(certDir, certName+".key")

	// 尝试删除文件，但不要因为文件不存在而失败
	if _, err := os.Stat(certPath); err == nil {
		if err := os.Remove(certPath); err != nil {
			return fmt.Errorf("删除证书公钥文件失败: %v", err)
		}
	}

	if _, err := os.Stat(keyPath); err == nil {
		if err := os.Remove(keyPath); err != nil {
			return fmt.Errorf("删除证书私钥文件失败: %v", err)
		}
	}

	// 获取版本并开始事务
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}

	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 检查证书是否已加载
	crtLoads, _, err := s.confClient.GetCrtLoads("sites", transaction.ID, 0, 0)
	if err == nil {
		// 查找并删除指定证书
		for _, crtLoad := range crtLoads {
			if crtLoad.Alias == certName {
				if err := s.confClient.DeleteCrtLoad("sites", crtLoad.Alias, transaction.ID, 0); err != nil {
					s.confClient.DeleteTransaction(transaction.ID)
					return fmt.Errorf("删除证书加载记录失败: %v", err)
				}
				break
			}
		}
	}

	// 提交事务
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 如果HAProxy已经运行，重新加载配置
	running, _ := s.isHAProxyRunning()
	if running && s.runtimeClient != nil {
		if _, err := s.runtimeClient.Reload(); err != nil {
			return fmt.Errorf("重新加载HAProxy配置失败: %v", err)
		}
	}

	return nil
}

// GetBackends 获取所有后端
func (s *HAProxyServiceImpl) GetBackends() ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return nil, err
	}

	backends, _, err := s.confClient.GetBackends("", 0, 0)
	if err != nil {
		return nil, fmt.Errorf("获取后端列表失败: %v", err)
	}

	var backendNames []string
	for _, backend := range backends {
		backendNames = append(backendNames, backend.Name)
	}

	return backendNames, nil
}

// AddBackend 添加后端
func (s *HAProxyServiceImpl) AddBackend(backend model.Backend) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}

	// 获取版本并开始事务
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}

	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 创建后端
	backendModel := &models.Backend{
		BackendBase: models.BackendBase{
			Name:    backend.Name,
			Mode:    "http",
			Enabled: true,
		},
	}

	if err := s.confClient.CreateBackend(backendModel, transaction.ID, 0); err != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("创建后端失败: %v", err)
	}

	// 添加服务器
	for _, server := range backend.Servers {
		serverModel := &models.Server{
			Name:    server.Name,
			Address: server.Host,
			Port:    s.Int64P(int64(server.Port)),
			ServerParams: models.ServerParams{
				Weight: s.Int64P(int64(server.Weight)),
				Check:  "enabled",
			},
		}

		if err := s.confClient.CreateServer("backend", backend.Name, serverModel, transaction.ID, 0); err != nil {
			s.confClient.DeleteTransaction(transaction.ID)
			return fmt.Errorf("创建服务器 %s 失败: %v", server.Name, err)
		}
	}

	// 提交事务
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 如果HAProxy已经运行，重新加载配置
	running, _ := s.isHAProxyRunning()
	if running && s.runtimeClient != nil {
		if _, err := s.runtimeClient.Reload(); err != nil {
			return fmt.Errorf("重新加载HAProxy配置失败: %v", err)
		}
	}

	return nil
}

// RemoveBackend 删除后端
func (s *HAProxyServiceImpl) RemoveBackend(backendName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}

	// 获取版本并开始事务
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}

	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 检查后端是否存在
	_, backendErr := s.confClient.GetBackend(backendName, transaction.ID)
	if backendErr == nil {
		// 后端存在，删除它
		if err := s.confClient.DeleteBackend(backendName, transaction.ID, 0); err != nil {
			s.confClient.DeleteTransaction(transaction.ID)
			return fmt.Errorf("删除后端失败: %v", err)
		}
	} else {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("后端 %s 不存在", backendName)
	}

	// 提交事务
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 如果HAProxy已经运行，重新加载配置
	running, _ := s.isHAProxyRunning()
	if running && s.runtimeClient != nil {
		if _, err := s.runtimeClient.Reload(); err != nil {
			return fmt.Errorf("重新加载HAProxy配置失败: %v", err)
		}
	}

	return nil
}

// AddServer 添加服务器到后端
func (s *HAProxyServiceImpl) AddServer(backendName string, server model.Server) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}

	// 获取版本并开始事务
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}

	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 检查后端是否存在
	_, backendErr := s.confClient.GetBackend(backendName, transaction.ID)
	if backendErr != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("后端 %s 不存在", backendName)
	}

	// 创建服务器
	serverModel := &models.Server{
		Name:    server.Name,
		Address: server.Host,
		Port:    s.Int64P(int64(server.Port)),
		ServerParams: models.ServerParams{
			Weight: s.Int64P(int64(server.Weight)),
			Check:  "enabled",
		},
	}

	if err := s.confClient.CreateServer("backend", backendName, serverModel, transaction.ID, 0); err != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("创建服务器失败: %v", err)
	}

	// 提交事务
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 如果HAProxy已经运行，重新加载配置
	running, _ := s.isHAProxyRunning()
	if running && s.runtimeClient != nil {
		if _, err := s.runtimeClient.Reload(); err != nil {
			return fmt.Errorf("重新加载HAProxy配置失败: %v", err)
		}
	}

	return nil
}

// RemoveServer 从后端删除服务器
func (s *HAProxyServiceImpl) RemoveServer(backendName string, serverName string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}

	// 获取版本并开始事务
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}

	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 检查后端是否存在
	_, backendErr := s.confClient.GetBackend(backendName, transaction.ID)
	if backendErr != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("后端 %s 不存在", backendName)
	}

	// 检查服务器是否存在
	_, serverErr := s.confClient.GetServer("backend", backendName, serverName, transaction.ID)
	if serverErr != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("服务器 %s 不存在", serverName)
	}

	// 删除服务器
	if err := s.confClient.DeleteServer("backend", backendName, serverName, transaction.ID, 0); err != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("删除服务器失败: %v", err)
	}

	// 提交事务
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 如果HAProxy已经运行，重新加载配置
	running, _ := s.isHAProxyRunning()
	if running && s.runtimeClient != nil {
		if _, err := s.runtimeClient.Reload(); err != nil {
			return fmt.Errorf("重新加载HAProxy配置失败: %v", err)
		}
	}

	return nil
}

// SetServerWeight 设置服务器权重
func (s *HAProxyServiceImpl) SetServerWeight(backendName string, serverName string, weight int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 如果运行时客户端未初始化，返回错误
	if s.runtimeClient == nil {
		return fmt.Errorf("HAProxy未运行或运行时客户端未初始化")
	}

	// 设置服务器权重
	err := s.runtimeClient.SetServerWeight(backendName, serverName, weight)
	if err != nil {
		return fmt.Errorf("设置服务器权重失败: %v", err)
	}

	return nil
}

// SetServerState 设置服务器状态(启用/禁用)
func (s *HAProxyServiceImpl) SetServerState(backendName string, serverName string, enabled bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 如果运行时客户端未初始化，返回错误
	if s.runtimeClient == nil {
		return fmt.Errorf("HAProxy未运行或运行时客户端未初始化")
	}

	// 设置服务器状态
	var action string
	if enabled {
		action = "ready"
	} else {
		action = "maint"
	}

	err := s.runtimeClient.SetServerState(backendName, serverName, action)
	if err != nil {
		return fmt.Errorf("设置服务器状态失败: %v", err)
	}

	return nil
}

// SetWAFMode 设置站点的WAF模式
func (s *HAProxyServiceImpl) SetWAFMode(domain string, mode model.WAFMode) error {
	// 确保模式有效
	if !model.IsValidWAFMode(mode) {
		return fmt.Errorf("无效的WAF模式: %s", mode)
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 确保配置客户端初始化
	if err := s.ensureConfClient(); err != nil {
		return err
	}

	// 获取版本并开始事务
	version, err := s.confClient.GetVersion("")
	if err != nil {
		return fmt.Errorf("获取版本失败: %v", err)
	}

	transaction, err := s.confClient.StartTransaction(version)
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	frontendName := "fe_" + domain

	// 检查前端是否存在
	_, frontendErr := s.confClient.GetFrontend(frontendName, transaction.ID)
	if frontendErr != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("站点 %s 不存在", domain)
	}

	// 获取HTTP请求规则
	httpReqRules, _, err := s.confClient.GetHTTPRequestRules(0, "frontend", frontendName, transaction.ID, 0, 0)
	if err != nil {
		s.confClient.DeleteTransaction(transaction.ID)
		return fmt.Errorf("获取HTTP请求规则失败: %v", err)
	}

	// 查找并更新WAF模式变量
	var wafModeVarFound bool
	for _, rule := range httpReqRules {
		if rule.Type == "set-var" && rule.VarName == "txn.waf_mode" {
			// 更新WAF模式变量
			rule.VarFormat = string(mode)

			if err := s.confClient.UpdateHTTPRequestRule(0, "frontend", frontendName, rule, transaction.ID, 0); err != nil {
				s.confClient.DeleteTransaction(transaction.ID)
				return fmt.Errorf("更新WAF模式变量失败: %v", err)
			}

			wafModeVarFound = true
			break
		}
	}

	// 如果未找到WAF模式变量，添加一个
	if !wafModeVarFound {
		wafModeVar := &models.HTTPRequestRule{
			Type:      "set-var",
			VarName:   "txn.waf_mode",
			VarFormat: string(mode),
		}

		if err := s.confClient.CreateHTTPRequestRule(0, "frontend", frontendName, wafModeVar, transaction.ID, 0); err != nil {
			s.confClient.DeleteTransaction(transaction.ID)
			return fmt.Errorf("创建WAF模式变量失败: %v", err)
		}
	}

	// 提交事务
	_, err = s.confClient.CommitTransaction(transaction.ID)
	if err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	// 如果HAProxy已经运行，重新加载配置
	running, _ := s.isHAProxyRunning()
	if running && s.runtimeClient != nil {
		if _, err := s.runtimeClient.Reload(); err != nil {
			return fmt.Errorf("重新加载HAProxy配置失败: %v", err)
		}
	}

	return nil
}

// ======================== 内部辅助方法 ========================

// initClients 初始化所有客户端
func (s *HAProxyServiceImpl) initClients() error {
	// 初始化配置客户端
	confClient, err := configuration.New(s.ctx,
		cfg_opt.ConfigurationFile(s.ConfigFile),
		cfg_opt.HAProxyBin(s.HaproxyBin),
		cfg_opt.Backups(s.BackupsNumber),
		cfg_opt.UsePersistentTransactions,
		cfg_opt.TransactionsDir(s.TransactionDir),
		cfg_opt.MasterWorker,
		cfg_opt.UseMd5Hash,
	)
	if err != nil {
		return fmt.Errorf("初始化配置客户端失败: %v", err)
	}
	s.confClient = confClient

	// 初始化SPOE客户端
	prms := spoe.Params{
		SpoeDir:        s.SpoeDir,
		TransactionDir: s.SpoeTransactionDir,
	}
	spoeClient, err := spoe.NewSpoe(prms)
	if err != nil {
		return fmt.Errorf("初始化SPOE客户端失败: %v", err)
	}
	s.spoeClient = spoeClient

	// 初始化运行时客户端
	masterSocketPath := filepath.Join(s.SocketDir, "haproxy-master.sock")
	ms := runtime_options.MasterSocket(masterSocketPath)
	runtimeClient, err := runtime_api.New(s.ctx, ms)
	if err != nil {
		return fmt.Errorf("初始化运行时客户端失败: %v", err)
	}
	s.runtimeClient = runtimeClient

	// 组合客户端
	clientOpts := []options.Option{
		options.Configuration(s.confClient),
		options.Runtime(s.runtimeClient),
		options.Spoe(s.spoeClient),
	}

	clientNative, err := client_native.New(s.ctx, clientOpts...)
	if err != nil {
		return fmt.Errorf("初始化客户端失败: %v", err)
	}
	s.clientNative = clientNative

	return nil
}

// ensureConfClient 确保配置客户端已初始化
func (s *HAProxyServiceImpl) ensureConfClient() error {
	if s.confClient == nil {
		confClient, err := configuration.New(s.ctx,
			cfg_opt.ConfigurationFile(s.ConfigFile),
			cfg_opt.HAProxyBin(s.HaproxyBin),
			cfg_opt.Backups(s.BackupsNumber),
			cfg_opt.UsePersistentTransactions,
			cfg_opt.TransactionsDir(s.TransactionDir),
			cfg_opt.MasterWorker,
			cfg_opt.UseMd5Hash,
		)
		if err != nil {
			return fmt.Errorf("初始化配置客户端失败: %v", err)
		}
		s.confClient = confClient
	}
	return nil
}

// stopHAProxy 停止HAProxy进程
func (s *HAProxyServiceImpl) stopHAProxy() error {
	// 如果实例中存储了HAProxy命令，使用它终止进程
	if s.haproxyCmd != nil && s.haproxyCmd.Process != nil {
		// 尝试优雅地终止进程
		if err := s.haproxyCmd.Process.Signal(os.Interrupt); err != nil {
			log.Printf("发送中断信号失败: %v", err)
			// 强制终止
			if err := s.haproxyCmd.Process.Kill(); err != nil {
				log.Printf("强制终止进程失败: %v", err)
			}
		}

		// 等待进程完全退出
		s.haproxyCmd.Wait()
		s.haproxyCmd = nil
	} else {
		// 尝试读取PID文件
		pid, err := s.getHAProxyPid()
		if err == nil && pid > 0 {
			process, err := os.FindProcess(pid)
			if err == nil {
				// 尝试优雅地终止进程
				if err := process.Signal(os.Interrupt); err != nil {
					log.Printf("发送中断信号失败: %v", err)
					// 强制终止
					if err := process.Kill(); err != nil {
						log.Printf("强制终止进程失败: %v", err)
					}
				}

				// 等待进程退出
				_, err = process.Wait()
				if err != nil {
					log.Printf("等待进程退出时出错: %v", err)
				}
			}
		} else {
			// 无法获取PID，尝试pkill
			exec.Command("pkill", "-f", s.HaproxyBin).Run()
		}
	}

	// 重置客户端
	s.runtimeClient = nil

	// 删除套接字文件
	socketPath := filepath.Join(s.SocketDir, "haproxy-master.sock")
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.Remove(socketPath); err != nil {
			log.Printf("删除套接字文件失败: %v", err)
		}
	}

	// 删除PID文件
	if _, err := os.Stat(s.PidFile); err == nil {
		if err := os.Remove(s.PidFile); err != nil {
			log.Printf("删除PID文件失败: %v", err)
		}
	}

	return nil
}

// isHAProxyRunning 检查HAProxy是否运行
func (s *HAProxyServiceImpl) isHAProxyRunning() (bool, error) {
	// 如果实例中存储了HAProxy命令，检查它
	if s.haproxyCmd != nil && s.haproxyCmd.Process != nil {
		// 尝试发送信号0检查进程是否存在
		if err := s.haproxyCmd.Process.Signal(syscall.Signal(0)); err != nil {
			// 进程不存在
			s.haproxyCmd = nil
			return false, nil
		}
		return true, nil
	}

	// 尝试读取PID文件
	pid, err := s.getHAProxyPid()
	if err != nil || pid <= 0 {
		return false, nil
	}

	// 检查进程是否存在
	process, err := os.FindProcess(pid)
	if err != nil {
		return false, nil
	}

	// 在Unix系统中，FindProcess总是成功的，需要发送信号0来确认进程存在
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return false, nil
	}

	return true, nil
}

// getHAProxyPid 从PID文件获取HAProxy进程ID
func (s *HAProxyServiceImpl) getHAProxyPid() (int, error) {
	if _, err := os.Stat(s.PidFile); os.IsNotExist(err) {
		return 0, fmt.Errorf("PID文件不存在")
	}

	pidBytes, err := os.ReadFile(s.PidFile)
	if err != nil {
		return 0, fmt.Errorf("读取PID文件失败: %v", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return 0, fmt.Errorf("解析PID失败: %v", err)
	}

	return pid, nil
}

// Int64P 返回指向int64的指针
func (s *HAProxyServiceImpl) Int64P(v int64) *int64 {
	return &v
}

// getSafeString 安全获取字符串指针的值
func (s *HAProxyServiceImpl) getSafeString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// getSafeInt64 安全获取int64指针的值
func (s *HAProxyServiceImpl) getSafeInt64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}
