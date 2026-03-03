// Package main 应用程序入口
package main

// @title Peanut API
// @version 1.0
// @description Go + Gin + SQLite 脚手架 API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@peanut.local

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/solariswu/peanut/api/v1/docs"
	"go.uber.org/zap"

	"github.com/solariswu/peanut/internal/agent/geo"
	"github.com/solariswu/peanut/internal/config"
	"github.com/solariswu/peanut/internal/handler"
	"github.com/solariswu/peanut/internal/middleware"
	"github.com/solariswu/peanut/internal/model"
	"github.com/solariswu/peanut/internal/pkg/database"
	"github.com/solariswu/peanut/internal/pkg/progress"
	"github.com/solariswu/peanut/internal/repository"
	"github.com/solariswu/peanut/internal/service"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "configs/config.yaml", "配置文件路径")
}

func main() {
	flag.Parse()

	// 加载 .env 文件（如果存在）
	if err := godotenv.Load(); err != nil {
		fmt.Println("未找到 .env 文件，使用系统环境变量")
	}

	// 加载配置
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger, err := initLogger(cfg)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// 连接 SQLite 数据库
	db, err := database.NewSQLite("peanut.db")
	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}
	defer db.Close()
	logger.Info("SQLite 数据库连接成功")

	// 执行数据库迁移
	if err := db.DB().AutoMigrate(&model.User{}, &model.GEOAnalysis{}); err != nil {
		logger.Warn("数据库迁移失败", zap.Error(err))
	} else {
		logger.Info("数据库迁移成功")
	}

	// 初始化服务
	userRepo := repository.NewUserRepository(db.DB())
	userSvc := service.NewUserService(userRepo)

	// 初始化 GEO 服务（使用 Google AI Overview）
	geoService, err := geo.NewDefaultService()
	if err != nil {
		logger.Warn("创建 GEO 服务失败", zap.Error(err))
		// 不退出，GEO 服务可选
	} else {
		logger.Info("GEO 服务初始化成功")
	}

	// 初始化进度管理器
	progressMgr := progress.NewManager()
	logger.Info("进度管理器初始化成功")

	// 初始化 GEO 分析服务（数据库版本）
	var geoAnalysisHandler *handler.GEOAnalysisHandler
	if geoService != nil {
		geoAnalysisRepo := repository.NewGEOAnalysisRepository(db.DB())
		geoAnalysisSvc := service.NewGEOAnalysisService(geoAnalysisRepo, geoService, progressMgr)
		geoAnalysisHandler = handler.NewGEOAnalysisHandler(geoAnalysisSvc, progressMgr)
		logger.Info("GEO 分析服务初始化成功")
	}

	// 初始化处理器
	userHandler := handler.NewUserHandler(userSvc)
	healthHandler := handler.NewHealthHandler()

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	router := gin.New()
	router.Use(
		middleware.Recovery(logger),
		middleware.Logger(logger),
		middleware.CORS(),
	)

	// 注册健康检查路由
	healthHandler.RegisterRoutes(router)

	// Swagger 文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 注册 API 路由
	api := router.Group("/api/v1")
	userHandler.RegisterRoutes(api)

	// 注册 GEO 分析路由
	if geoAnalysisHandler != nil {
		geoAnalysisHandler.RegisterRoutes(api)
		logger.Info("GEO 分析路由已注册")
	}

	// 启动服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 优雅关闭
	go func() {
		logger.Info("启动服务器", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("启动服务器失败", zap.Error(err))
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务器强制关闭", zap.Error(err))
	}

	logger.Info("服务器已退出")
}

// initLogger 初始化日志
func initLogger(cfg *config.Config) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Log.Format == "json" {
		logger, err = zap.NewProduction()
	} else {
		config := zap.NewDevelopmentConfig()
		switch cfg.Log.Level {
		case "debug":
			config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		case "info":
			config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		case "warn":
			config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		case "error":
			config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
		}
		logger, err = config.Build()
	}

	return logger, err
}
