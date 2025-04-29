package router

import (
    "errors"
    "strings"

    "github.com/kwrum1/waf/server/controller"
    "github.com/kwrum1/waf/server/middleware"
    "github.com/kwrum1/waf/server/model"
    "github.com/kwrum1/waf/server/repository"
    "github.com/kwrum1/waf/server/service"
    "github.com/kwrum1/waf/server/utils/response"

    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/v2/mongo"
)

// Setup configures all the routes for the application
func Setup(route *gin.Engine, db *mongo.Database) {
    // 基础中间件
    route.Use(middleware.RequestID())
    route.Use(middleware.Logger())
    route.Use(middleware.Cors())
    route.Use(gin.CustomRecovery(middleware.CustomErrorHandler))

    // 创建仓库
    userRepo := repository.NewUserRepository(db)
    roleRepo := repository.NewRoleRepository(db)
    siteRepo := repository.NewSiteRepository(db)
    wafLogRepo := repository.NewWAFLogRepository(db)
    certRepo := repository.NewCertificateRepository(db)
    configRepo := repository.NewConfigRepository(db)

    // 创建服务
    authService := service.NewAuthService(userRepo, roleRepo)
    siteService := service.NewSiteService(siteRepo)
    wafLogService := service.NewWAFLogService(wafLogRepo)
    certService := service.NewCertificateService(certRepo)
    runnerService, _ := service.NewRunnerService()
    configService := service.NewConfigService(configRepo)

    // 创建控制器
    authController := controller.NewAuthController(authService)
    siteController := controller.NewSiteController(siteService)
    wafLogController := controller.NewWAFLogController(wafLogService)
    certController := controller.NewCertificateController(certService)
    runnerController := controller.NewRunnerController(runnerService)
    configController := controller.NewConfigController(configService)

    // 将仓库添加到上下文中，供中间件使用
    route.Use(func(c *gin.Context) {
        c.Set("userRepo", userRepo)
        c.Set("roleRepo", roleRepo)
        c.Next()
    })

    // 健康检查端点
    route.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // API v1 路由
    api := route.Group("/api/v1")

    // 认证相关路由 - 不需要权限检查
    auth := api.Group("/auth")
    {
        auth.POST("/login", authController.Login)

        // 需要认证的路由
        authRequired := auth.Group("")
        authRequired.Use(middleware.JWTAuth())
        {
            // 密码重置接口 - 任何已认证用户都可访问
            authRequired.POST("/reset-password", authController.ResetPassword)

            // 需要密码重置检查的路由
            passwordChecked := authRequired.Group("")
            passwordChecked.Use(middleware.PasswordResetRequired())
            {
                // 获取个人信息 - 任何已认证用户都可访问
                passwordChecked.GET("/me", authController.GetUserInfo)
            }
        }
    }

    // 需要认证和密码重置检查的API路由
    authenticated := api.Group("")
    authenticated.Use(middleware.JWTAuth())
    authenticated.Use(middleware.PasswordResetRequired())

    // 用户管理模块
    userRoutes := authenticated.Group("/users")
    {
        userRoutes.POST("", middleware.HasPermission(model.PermUserCreate), authController.CreateUser)
        userRoutes.GET("", middleware.HasPermission(model.PermUserRead), authController.GetUsers)
        userRoutes.PUT("/:id", middleware.HasPermission(model.PermUserUpdate), authController.UpdateUser)
        userRoutes.DELETE("/:id", middleware.HasPermission(model.PermUserDelete), authController.DeleteUser)
    }

    // 站点管理模块
    siteRoutes := authenticated.Group("/site")
    {
        siteRoutes.POST("", middleware.HasPermission(model.PermSiteCreate), siteController.CreateSite)
        siteRoutes.GET("", middleware.HasPermission(model.PermSiteRead), siteController.GetSites)
        siteRoutes.GET("/:id", middleware.HasPermission(model.PermSiteRead), siteController.GetSiteByID)
        siteRoutes.PUT("/:id", middleware.HasPermission(model.PermSiteUpdate), siteController.UpdateSite)
        siteRoutes.DELETE("/:id", middleware.HasPermission(model.PermSiteDelete), siteController.DeleteSite)
    }

    // 证书管理路由
    certRoutes := authenticated.Group("/certificate")
    {
        certRoutes.POST("", middleware.HasPermission(model.PermCertCreate), certController.CreateCertificate)
        certRoutes.GET("", middleware.HasPermission(model.PermCertRead), certController.GetCertificates)
        certRoutes.GET("/:id", middleware.HasPermission(model.PermCertRead), certController.GetCertificateByID)
        certRoutes.PUT("/:id", middleware.HasPermission(model.PermCertUpdate), certController.UpdateCertificate)
        certRoutes.DELETE("/:id", middleware.HasPermission(model.PermCertDelete), certController.DeleteCertificate)
    }

    // 日志
    wafLogRoutes := authenticated.Group("/log")
    {
        wafLogRoutes.GET("/event", middleware.HasPermission(model.PermWAFLogRead), wafLogController.GetAttackEvents)
        wafLogRoutes.GET("", middleware.HasPermission(model.PermWAFLogRead), wafLogController.GetAttackLogs)
    }

    // 配置管理模块
    runnerRoutes := authenticated.Group("/runner")
    {
        runnerRoutes.GET("/status", middleware.HasPermission(model.PermConfigRead), runnerController.GetStatus)
        runnerRoutes.POST("/control", middleware.HasPermission(model.PermConfigUpdate), runnerController.Control)
    }
    configRoutes := authenticated.Group("/config")
    {
        configRoutes.GET("", middleware.HasPermission(model.PermConfigRead), configController.GetConfig)
        configRoutes.PATCH("", middleware.HasPermission(model.PermConfigUpdate), configController.PatchConfig)
    }

    // Suricata 事件查询路由
    suriRepo := repository.NewSuricataRepository()
    suriSvc  := service.NewSuricataService(suriRepo)
    suriCtrl := controller.NewSuricataController(suriSvc)
    suriAPI  := api.Group("/suricata")
    suriAPI.GET("/events", suriCtrl.ListEvents)

    // 审计日志模块
    auditRoutes := authenticated.Group("/audit")
    {
        auditRoutes.GET("", middleware.HasPermission(model.PermAuditRead), nil)
    }

    // 系统管理模块
    systemRoutes := authenticated.Group("/system")
    {
        systemRoutes.GET("/status", middleware.HasPermission(model.PermSystemStatus), nil)
        systemRoutes.POST("/restart", middleware.HasPermission(model.PermSystemRestart), nil)
    }

    // ===== 前端静态资源托管 =====
    route.Static("/assets", "./web/dist/assets")
    route.Static("/locales", "./web/dist/locales")
    route.StaticFile("/favicon.ico", "./web/dist/favicon.ico")
    route.StaticFile("/logo.png", "./web/dist/logo.png")
    route.StaticFile("/logo32.png", "./web/dist/logo32.png")
    route.StaticFile("/logo.svg", "./web/dist/logo.svg")

    // NoRoute 处理
    route.NoRoute(func(c *gin.Context) {
        if strings.HasPrefix(c.Request.URL.Path, "/api") {
            response.BadRequest(c, errors.New("API路由不存在"), true)
            return
        }
        c.File("./web/dist/index.html")
    })
}
