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
	"github.com/zouluxing/ums/internal/model"
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

	// ── RSA key pair ──────────────────────────────────────────────────────────
	if cfg.JWT.PrivateKeyPath == "" {
		cfg.JWT.PrivateKeyPath = "configs/ums_rsa_private.pem"
		cfg.JWT.PublicKeyPath = "configs/ums_rsa_public.pem"
	}
	if _, err := os.Stat(cfg.JWT.PrivateKeyPath); os.IsNotExist(err) {
		log.Println("RSA keys not found, generating 2048-bit key pair...")
		if err := utils.GenerateRSAKeyPairFiles(cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath); err != nil {
			log.Fatalf("generate RSA keys: %v", err)
		}
	}

	jwtMgr, err := utils.NewJWTManager(
		cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath,
		cfg.JWT.Issuer, cfg.JWT.KeyID,
		cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL,
	)
	if err != nil {
		log.Fatalf("jwt manager: %v", err)
	}

	// ── MFA encryption (RISK-001) ─────────────────────────────────────────────
	// Production: set UMS_MFA_ENCRYPTION_KEY to a 64-char hex string.
	// Generate with: openssl rand -hex 32
	var mfaEnc service.MFAEncryptor
	mfaKeyHex := os.Getenv("UMS_MFA_ENCRYPTION_KEY")
	if mfaKeyHex != "" {
		enc, err := crypto.NewAESEncryptor(mfaKeyHex)
		if err != nil {
			log.Fatalf("MFA encryptor init: %v", err)
		}
		mfaEnc = enc
		log.Println("[security] MFA secrets will be encrypted at rest (AES-256-GCM)")
	} else {
		if cfg.App.Env == "production" {
			log.Fatal("[security] UMS_MFA_ENCRYPTION_KEY must be set in production")
		}
		log.Println("[WARN] UMS_MFA_ENCRYPTION_KEY not set — MFA secrets stored in plaintext (dev only)")
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

	if cfg.App.Env == "development" {
		if err := db.AutoMigrate(
			&model.User{}, &model.Role{}, &model.Permission{},
			&model.UserRole{}, &model.RolePermission{},
			&model.OAuth2Client{}, &model.OAuth2AuthCode{}, &model.OAuth2Token{},
		); err != nil {
			log.Fatalf("auto-migrate: %v", err)
		}
		seedSystemData(db)
	}

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
	oauth2H := handler.NewOAuth2Handler(oauth2Svc, jwtMgr, authSvc)

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

	oauth := r.Group("/api/v1/oauth2")
	{
		oauth.GET("/authorize", handler.JWTAuthWithUserID(jwtMgr, cacheSvc), oauth2H.Authorize)
		oauth.POST("/token", oauth2H.Token)
		oauth.POST("/introspect", oauth2H.Introspect)
		clients := oauth.Group("/clients", handler.JWTAuth(jwtMgr, cacheSvc))
		{
			clients.GET("", oauth2H.ListClients)
			clients.POST("", oauth2H.CreateClient)
			clients.PUT("/:id", oauth2H.UpdateClient)
			clients.DELETE("/:id", oauth2H.DeleteClient)
			clients.POST("/:id/secret/rotate", oauth2H.RotateSecret)
		}
	}

	v1 := r.Group("/api/v1")
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.POST("/token/refresh", authH.RefreshToken)
		auth.POST("/logout", handler.JWTAuth(jwtMgr, cacheSvc), authH.Logout)
	}

	authMW  := handler.JWTAuth(jwtMgr, cacheSvc)
	adminMW := handler.RequireRole("admin")
	users := v1.Group("/users", authMW)
	{
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
	}

	roles := v1.Group("/roles", authMW, adminMW)
	{
		roles.GET("", permH.ListRoles)
		roles.POST("", permH.CreateRole)
		roles.DELETE("/:id", permH.DeleteRole)
		roles.POST("/:id/permissions", permH.SetRolePermissions)
	}
	v1.GET("/permissions", authMW, adminMW, permH.ListPermissions)

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

func seedSystemData(db *gorm.DB) {
	perms := []model.Permission{
		{Resource: "user", Action: "create"}, {Resource: "user", Action: "read"},
		{Resource: "user", Action: "update"}, {Resource: "user", Action: "delete"},
		{Resource: "role", Action: "create"}, {Resource: "role", Action: "read"},
		{Resource: "role", Action: "update"}, {Resource: "role", Action: "delete"},
		{Resource: "oauth2_client", Action: "create"}, {Resource: "oauth2_client", Action: "read"},
	}
	for i := range perms {
		db.Where(model.Permission{Resource: perms[i].Resource, Action: perms[i].Action}).FirstOrCreate(&perms[i])
	}
	desc := func(s string) *string { return &s }
	for _, role := range []model.Role{
		{Name: "admin", IsSystem: true, Description: desc("超级管理员")},
		{Name: "editor", IsSystem: true, Description: desc("编辑者")},
		{Name: "viewer", IsSystem: true, Description: desc("只读用户")},
	} {
		role := role
		db.Where(model.Role{Name: role.Name}).FirstOrCreate(&role)
	}
	var adminRole model.Role
	db.First(&adminRole, "name = ?", "admin")
	var allPerms []model.Permission
	db.Find(&allPerms)
	db.Model(&adminRole).Association("Permissions").Replace(allPerms)
}
