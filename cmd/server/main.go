package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zouluxing/feature_orbit_server/internal/config"
	"github.com/zouluxing/feature_orbit_server/internal/handler"
	"github.com/zouluxing/feature_orbit_server/internal/middleware"
	"github.com/zouluxing/feature_orbit_server/internal/repository"
	"github.com/zouluxing/feature_orbit_server/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil { log.Fatalf("load config: %v", err) }

	umsClient, err := middleware.NewUMSClient(cfg.UMSBaseURL, cfg.UMSCacheTTL)
	if err != nil { log.Fatalf("UMS client init: %v", err) }

	gormCfg := &gorm.Config{}
	if !cfg.Debug { gormCfg.Logger = logger.Default.LogMode(logger.Silent) }
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), gormCfg)
	if err != nil { log.Fatalf("connect postgres: %v", err) }
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(10); sqlDB.SetMaxIdleConns(5); sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// 数据库初始化策略：不使用 GORM AutoMigrate，完全依赖 SQL 迁移文件。
	// docker-compose 通过 /docker-entrypoint-initdb.d 在容器首次创建时执行
	// migrations/001_create_features.up.sql，应用启动时只验证表存在，不触发任何 DDL。
	// 这样避免 GORM 自动生成的约束名（uni_xxx）与 migration SQL 手写约束名（uq_xxx）冲突。
	if err := verifyDatabaseReady(db); err != nil {
		log.Fatalf("database not ready: %v", err)
	}
	log.Println("database schema verified")

	featureRepo := repository.NewFeatureRepository(db)
	featureSvc  := service.NewFeatureService(featureRepo)
	featureH    := handler.NewFeatureHandler(featureSvc)

	if !cfg.Debug { gin.SetMode(gin.ReleaseMode) }
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == http.MethodOptions { c.AbortWithStatus(http.StatusNoContent); return }
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "feature_orbit_server"})
	})

	api := r.Group("/api/v1")
	api.GET("/features", featureH.ListFeatures)
	api.GET("/features/:id", featureH.GetFeature)

	auth := api.Group("/", umsClient.GinMiddleware())
	auth.GET("/me", handler.GetMe)
	auth.POST("/features", featureH.CreateFeature)
	auth.PUT("/features/:id", featureH.UpdateFeature)
	auth.DELETE("/features/:id", featureH.DeleteFeature)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}
	go func() {
		log.Printf("feature_orbit_server listening on :%d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("feature_orbit_server stopped.")
}

// verifyDatabaseReady 检查关键表是否存在，不执行任何 DDL。
// 表不存在时说明 migration 还未运行，报错提示而不是不断重启。
func verifyDatabaseReady(db *gorm.DB) error {
	required := []string{"features"}
	for _, table := range required {
		var count int64
		result := db.Raw(
			"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name=?",
			table,
		).Scan(&count)
		if result.Error != nil {
			return fmt.Errorf("check table %s: %w", table, result.Error)
		}
		if count == 0 {
			return fmt.Errorf(
				"table '%s' not found — migration has not run yet.\n"+
					"Fix: docker compose down -v && docker compose up -d",
				table,
			)
		}
	}
	return nil
}
