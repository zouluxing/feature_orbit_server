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

	"github.com/zouluxing/feature_orbit_server/internal/config"
	"github.com/zouluxing/feature_orbit_server/internal/handler"
	"github.com/zouluxing/feature_orbit_server/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Initialise the UMS client (fetches JWKS on startup, caches locally).
	umsClient, err := middleware.NewUMSClient(cfg.UMSBaseURL, cfg.UMSCacheTTL)
	if err != nil {
		log.Fatalf("UMS client init: %v", err)
	}

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Health (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "feature_orbit_server"})
	})

	api := r.Group("/api/v1")

	// Public routes
	public := api.Group("/")
	{
		public.GET("/features", handler.ListFeatures)
	}

	// Authenticated routes — UMS JWT middleware
	auth := api.Group("/", umsClient.GinMiddleware())
	{
		auth.POST("/features", handler.CreateFeature)
		auth.PUT("/features/:id", handler.UpdateFeature)
		auth.DELETE("/features/:id", umsClient.RequireRole("admin"), handler.DeleteFeature)
		auth.GET("/me", handler.GetMe)
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
