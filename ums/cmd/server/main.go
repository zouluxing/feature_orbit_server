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
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/zouluxing/ums/internal/config"
	"github.com/zouluxing/ums/internal/handler"
	"github.com/zouluxing/ums/internal/repository"
	"github.com/zouluxing/ums/internal/service"
	"github.com/zouluxing/ums/pkg/crypto"
	"github.com/zouluxing/ums/pkg/utils"
)

func main() {
	cfgPath := os.Getenv("UMS_CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// ── RSA key pair ─────────────────────────────────────────────────────────
	if cfg.JWT.PrivateKeyPath == "" {
		cfg.JWT.PrivateKeyPath = "configs/ums_rsa_private.pem"
		cfg.JWT.PublicKeyPath = "configs/ums_rsa_public.pem"
	}
	if _, err := os.Stat(cfg.JWT.PrivateKeyPath); os.IsNotExist(err) {
		log.Println("RSA keys not found, generating 2048-bit key pair...")
		if err := utils.GenerateRSAKeyPairFiles(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath); err != nil {
			log.Fatalf("generate RSA keys: %v", err)
		}
		log.Printf("RSA keys written to %s / %s", cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath)
	}

	jwtMgr, err := utils.NewJWTManager(
		cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath,
		cfg.JWT.Issuer, cfg.JWT.KeyID,
		cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL,
	)
	if err != nil {
		log.Fatalf("jwt manager: %v", err)
	}

	// ── MFA encryption ────────────────────────────────────────────────────────
	var mfaEnc service.MFAEncryptor
	if mfaKeyHex := os.Getenv("UMS_MFA_ENCRYPTION_KEY"); mfaKeyHex != "" {
		enc, err := crypto.NewAESEncryptor(mfaKeyHex)
		if err != nil {
			log.Fatalf("MFA encryptor init: %v", err)
		}
		mfaEnc = enc
		log.Println("[security] MFA secrets encrypted at rest (AES-256-GCM)")
	} else {
		if cfg.App.Env == "production" {
			log.Fatal("[security] UMS_MFA_ENCRYPTION_KEY must be set in production")
		}
		log.Println("[WARN] UMS_MFA_ENCRYPTION_KEY not set — MFA secrets in plaintext (dev only)")
	}

	// ── PostgreSQL ────────────────────────────────────────────────────────────
	gormCfg := &gorm.Config{}
	if !cfg.App.Debug {
		gormCfg.Logger = logger.Default.LogMode(logger.Silent)
	}
	db, err := gorm.Open(postgres.Open(cfg.DB.URL), gormCfg)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// 数据库初始化策略：不使用 GORM AutoMigrate，完全依赖 SQL 迁移文件
	// docker-entrypoint-initdb.d 在容器首次创建时执行 migrations/000001_init.up.sql
	// 应用启动时只做连通性验证，不重复执行 DDL
	if err := verifyDatabaseReady(db); err != nil {
		log.Fatalf("database not ready: %v", err)
	}
	log.Println("database schema verified")

	// seed 基础数据（幂等，只插入缺少的部分）
	seedSystemData(db)

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("connect redis: %v", err)
	}

	// ── Wire dependencies ─────────────────────────────────────────────────────
	cacheSvc  := service.NewCacheService(rdb)
	userRepo  := repository.NewUserRepository(db)
	roleRepo  := repository.NewRoleRepository(db)
	permRepo  := repository.NewPermissionRepository(db)
	oauthRepo := repository.NewOAuth2Repository(db)

	authSvc   := service.NewAuthService(userRepo, cacheSvc, jwtMgr, cfg, mfaEnc)
	userSvc   := service.NewUserService(userRepo, roleRepo, mfaEnc)
	permSvc   := service.NewPermissionService(roleRepo, permRepo, userRepo)
	oauth2Svc := service.NewOAuth2Service(oauthRepo, userRepo, jwtMgr, cfg)

	authH   := handler.NewAuthHandler(authSvc, jwtMgr)
	userH   := handler.NewUserHandler(userSvc)
	permH   := handler.NewPermissionHandler(permSvc)
	oauth2H := handler.NewOAuth2Handler(oauth2Svc, authSvc)

	// ── Router ────────────────────────────────────────────────────────────────
	if !cfg.App.Debug {
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
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "ums"})
	})
	r.GET("/.well-known/jwks.json", authH.JWKS)

	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.POST("/register", authH.Register)
	auth.POST("/login", authH.Login)
	auth.POST("/token/refresh", authH.RefreshToken)
	auth.POST("/logout", handler.JWTAuth(jwtMgr, cacheSvc), authH.Logout)

	authMW  := handler.JWTAuth(jwtMgr, cacheSvc)
	adminMW := handler.RequireRole("admin")
	users := v1.Group("/users", authMW)
	users.GET("/me", userH.GetMe)
	users.PUT("/me", userH.UpdateMe)
	users.PUT("/me/password", userH.UpdateMyPassword)
	users.POST("/me/mfa/enable", userH.EnableMFA)
	users.POST("/me/mfa/disable", userH.DisableMFA)
	users.GET("", adminMW, userH.ListUsers)
	users.POST("", adminMW, userH.CreateUser)
	users.GET("/:uuid", adminMW, userH.GetUser)
	users.PUT("/:uuid", adminMW, userH.UpdateUser)
	users.PATCH("/:uuid/status", adminMW, userH.PatchUserStatus)
	users.DELETE("/:uuid", adminMW, userH.DeleteUser)
	users.PUT("/:uuid/password", adminMW, userH.ResetPassword)
	users.POST("/:uuid/roles", adminMW, userH.AssignRoles)
	users.DELETE("/:uuid/roles/:role_id", adminMW, userH.RemoveRole)
	users.GET("/:uuid/permissions", adminMW, userH.GetUserPermissions)

	roles := v1.Group("/roles", authMW, adminMW)
	roles.GET("", permH.ListRoles)
	roles.POST("", permH.CreateRole)
	roles.DELETE("/:id", permH.DeleteRole)
	roles.POST("/:id/permissions", permH.SetRolePermissions)
	v1.GET("/permissions", authMW, adminMW, permH.ListPermissions)

	oauth2 := v1.Group("/oauth2")
	oauth2.GET("/authorize", handler.JWTAuthWithUserID(jwtMgr, cacheSvc), oauth2H.Authorize)
	oauth2.POST("/token", oauth2H.Token)
	oauth2.POST("/introspect", oauth2H.Introspect)
	clients := oauth2.Group("/clients", authMW)
	clients.GET("", oauth2H.ListClients)
	clients.POST("", oauth2H.CreateClient)
	clients.PUT("/:id", oauth2H.UpdateClient)
	clients.DELETE("/:id", oauth2H.DeleteClient)
	clients.POST("/:id/secret/rotate", oauth2H.RotateSecret)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		log.Printf("UMS listening on :%d (env=%s)", cfg.App.Port, cfg.App.Env)
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
	log.Println("UMS stopped.")
}

// verifyDatabaseReady 检查关键表是否存在，不执行任何 DDL。
// 表不存在时说明迁移文件还没执行，给出明确错误提示而不是暴庋重启。
func verifyDatabaseReady(db *gorm.DB) error {
	requiredTables := []string{"users", "roles", "permissions", "user_roles", "role_permissions"}
	for _, table := range requiredTables {
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
				"table '%s' not found — migrations may not have run yet.\n"+
					"If this is a fresh start, stop the ums container, delete the postgres volume, "+
					"and restart: docker compose down -v && docker compose up -d",
				table,
			)
		}
	}
	return nil
}

// seedSystemData 必须幂等：只插入不存在的行，不修改已有数据。
func seedSystemData(db *gorm.DB) {
	type perm struct{ Resource, Action, Description string }
	perms := []perm{
		{"user", "create", "创建用户"}, {"user", "read", "查看用户"},
		{"user", "update", "修改用户"}, {"user", "delete", "删除用户"},
		{"role", "create", "创建角色"}, {"role", "read", "查看角色"},
		{"role", "update", "修改角色"}, {"role", "delete", "删除角色"},
		{"oauth2_client", "create", "注册OAuth应用"}, {"oauth2_client", "read", "查看OAuth应用"},
	}
	for _, p := range perms {
		db.Exec(
			"INSERT INTO permissions (resource, action, description) VALUES (?,?,?) ON CONFLICT DO NOTHING",
			p.Resource, p.Action, p.Description,
		)
	}
	type role struct{ Name, Description string }
	for _, r := range []role{{"admin", "超级管理员"}, {"editor", "编辑者"}, {"viewer", "只读用户"}} {
		db.Exec(
			"INSERT INTO roles (name, description, is_system) VALUES (?,?,TRUE) ON CONFLICT DO NOTHING",
			r.Name, r.Description,
		)
	}
	// admin 角色拥有所有权限
	db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT r.id, p.id FROM roles r, permissions p WHERE r.name = 'admin'
		ON CONFLICT DO NOTHING
	`)
}
