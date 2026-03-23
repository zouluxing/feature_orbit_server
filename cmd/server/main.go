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
	"github.com/zouluxing/feature_orbit_server/internal/model"
	"github.com/zouluxing/feature_orbit_server/internal/repository"
	"github.com/zouluxing/feature_orbit_server/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// ── UMS client (JWT verification via JWKS) ────────────────────────────
	umsClient, err := middleware.NewUMSClient(cfg.UMSBaseURL, cfg.UMSCacheTTL)
	if err != nil {
		log.Fatalf("UMS client init: %v", err)
	}

	// ── PostgreSQL ──────────────────────────────────────────────────────────
	gormCfg := &gorm.Config{}
	if !cfg.Debug {
		gormCfg.Logger = logger.Default.LogMode(logger.Silent)
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), gormCfg)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	if cfg.Debug {
		// Auto-migrate in dev; use golang-migrate in production
		if err := db.AutoMigrate(&model.Feature{}); err != nil {
			log.Fatalf("auto-migrate: %v", err)
		}
	}

	// ── Wire dependencies ─────────────────────────────────────────────────────
	featureRepo := repository.NewFeatureRepository(db)
	featureSvc  := service.NewFeatureService(featureRepo)
	featureH    := handler.NewFeatureHandler(featureSvc)

	// ── Router ────────────────────────────────────────────────────────────────
	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "feature_orbit_server"})
	})

	api := r.Group("/api/v1")

	// — Public routes
	api.GET("/features", featureH.ListFeatures)
	api.GET("/features/:id", featureH.GetFeature)

	// — Authenticated routes (UMS JWT middleware)
	authGroup := api.Group("/", umsClient.GinMiddleware())
	{
		authGroup.GET("/me", handler.GetMe)
		authGroup.POST("/features", featureH.CreateFeature)
		authGroup.PUT("/features/:id", featureH.UpdateFeature)
		// Delete: authenticated (owner) OR admin role
		authGroup.DELETE("/features/:id", featureH.DeleteFeature)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
