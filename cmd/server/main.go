// Package main 应用程序入口
package main

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
	"go.uber.org/zap"

	"github.com/solariswu/peanut/internal/config"
	"github.com/solariswu/peanut/internal/handler"
	"github.com/solariswu/peanut/internal/middleware"
	"github.com/solariswu/peanut/internal/pkg/cache"
	"github.com/solariswu/peanut/internal/pkg/database"
	"github.com/solariswu/peanut/internal/repository"
	"github.com/solariswu/peanut/internal/service"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "configs/config.yaml", "配置文件路径")
}

func main() {
	flag.Parse()

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

	// 连接数据库
	db, err := database.NewPostgreSQL(&cfg.Database)
	if err != nil {
		logger.Fatal("连接数据库失败", zap.Error(err))
	}
	defer db.Close()
	logger.Info("数据库连接成功")

	// 连接 Redis
	rdb, err := cache.NewRedis(&cfg.Redis)
	if err != nil {
		logger.Fatal("连接 Redis 失败", zap.Error(err))
	}
	defer rdb.Close()
	logger.Info("Redis 连接成功")

	// 初始化服务
	userRepo := repository.NewUserRepository(db.DB())
	userSvc := service.NewUserService(userRepo)

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

	// 注册 API 路由
	api := router.Group("/api/v1")
	userHandler.RegisterRoutes(api)

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
