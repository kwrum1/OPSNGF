package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mvrilo/go-redoc"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	mongodb "github.com/HUAHUAI23/simple-waf/pkg/database/mongo"
	"github.com/HUAHUAI23/simple-waf/server/config"
	_ "github.com/HUAHUAI23/simple-waf/server/docs" // 导入 swagger 文档
	"github.com/HUAHUAI23/simple-waf/server/router"
	"github.com/HUAHUAI23/simple-waf/server/service/daemon"
	"github.com/HUAHUAI23/simple-waf/server/validator"
)

//	@title			Simple-WAF API
//	@version		1.0
//	@description	简单的 Web 应用防火墙管理系统 API
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:2333
//	@BasePath	/api/v1

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				使用 Bearer {token} 格式进行身份验证
func main() {
	// Load configuration
	err := config.InitConfig()
	if err != nil {
		config.GlobalLogger.Error().Err(err).Msg("Failed to load configuration")
		return
	}

	// 连接数据库
	client, err := mongodb.Connect(config.Global.DBConfig.URI)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to connect to database")
		return
	}

	// 获取数据库
	db := client.Database(config.Global.DBConfig.Database)

	err = config.InitDB(db)
	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to initialize database")
	}

	// Create service runner and start background services
	runner, err := daemon.GetRunnerService()

	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to create service runner")
		return
	}

	err = runner.StartServices()
	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to start daemon services")
		return
	}

	// Set Gin mode based on configuration
	if config.Global.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize the Gin route
	route := gin.New()

	// Setup the router
	router.Setup(route, db)

	// 初始化验证器
	validator.InitValidators()
	// validators.InitStructValidators()

	// 设置 Swagger 文档
	route.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler,
		ginSwagger.URL("/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(2),
		ginSwagger.DocExpansion("list"),
		ginSwagger.DeepLinking(true),
		ginSwagger.PersistAuthorization(true),
	))

	// 获取Redoc处理器
	doc := redoc.Redoc{
		Title:       "Simple-WAF API",
		Description: "简单的 Web 应用防火墙管理系统 API",
		SpecFile:    "./docs/swagger.json",
		SpecPath:    "/swagger.json",
		DocsPath:    "/redoc",
	}
	redocHandler := doc.Handler()

	// 明确定义Redoc路由
	route.GET("/redoc", func(c *gin.Context) {
		redocHandler(c.Writer, c.Request)
	})
	route.GET("/swagger.json", func(c *gin.Context) {
		redocHandler(c.Writer, c.Request)
	})

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    config.Global.Bind,
		Handler: route,
	}

	// 创建一个错误通道
	serverError := make(chan error, 1)

	// 在goroutine中启动服务器
	go func() {
		config.Logger.Info().Msgf("Starting server on %s", config.Global.Bind)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			config.Logger.Error().Err(err).Msg("Server error")
			serverError <- err
		}
	}()

	// 等待中断信号或服务器错误
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 使用select等待任一通道有消息
	select {
	case <-quit:
		config.Logger.Info().Msg("Received shutdown signal, shutting down services...")
	case err := <-serverError:
		config.Logger.Error().Err(err).Msg("Server failed, initiating shutdown...")
	}

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试优雅关闭服务器
	if err := srv.Shutdown(ctx); err != nil {
		config.Logger.Error().Err(err).Msg("Server forced to shutdown")
	} else {
		config.Logger.Info().Msg("Server shutdown gracefully")
	}

	// 停止后台服务
	err = runner.StopServices()
	if err != nil {
		config.Logger.Error().Err(err).Msg("Failed to stop daemon services")
	}
	config.Logger.Info().Msg("Background services have been shut down, exiting...")

	// 如果是因为服务器错误而退出，使用非零状态码
	if len(serverError) > 0 {
		os.Exit(1)
	}
}
