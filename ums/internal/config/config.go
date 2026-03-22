package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App       AppConfig
	DB        DBConfig
	Redis     RedisConfig
	JWT       JWTConfig
	OAuth2    OAuth2Config
	RateLimit RateLimitConfig
}

type AppConfig struct {
	Name  string
	Env   string
	Port  int
	Debug bool
}

type DBConfig struct {
	URL          string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type JWTConfig struct {
	PrivateKeyPath string
	PublicKeyPath  string
	AccessTTL      time.Duration
	RefreshTTL     time.Duration
	Issuer         string
	KeyID          string
}

type OAuth2Config struct {
	AuthCodeTTL time.Duration
}

type RateLimitConfig struct {
	LoginMaxAttempts int
	LoginWindowSecs  int
	LockDurationMins int
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("UMS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("app.port", 8081)
	v.SetDefault("app.env", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("db.max_open_conns", 10)
	v.SetDefault("db.max_idle_conns", 5)
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.access_ttl", "15m")
	v.SetDefault("jwt.refresh_ttl", "168h")
	v.SetDefault("jwt.issuer", "ums")
	v.SetDefault("jwt.key_id", "ums-rsa-1")
	v.SetDefault("oauth2.auth_code_ttl", "10m")
	v.SetDefault("rate_limit.login_max_attempts", 5)
	v.SetDefault("rate_limit.login_window_secs", 300)
	v.SetDefault("rate_limit.lock_duration_mins", 30)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	accessTTL, _  := time.ParseDuration(v.GetString("jwt.access_ttl"))
	refreshTTL, _ := time.ParseDuration(v.GetString("jwt.refresh_ttl"))
	authCodeTTL, _ := time.ParseDuration(v.GetString("oauth2.auth_code_ttl"))
	return &Config{
		App:   AppConfig{Name: v.GetString("app.name"), Env: v.GetString("app.env"), Port: v.GetInt("app.port"), Debug: v.GetBool("app.debug")},
		DB:    DBConfig{URL: v.GetString("db.url"), MaxOpenConns: v.GetInt("db.max_open_conns"), MaxIdleConns: v.GetInt("db.max_idle_conns")},
		Redis: RedisConfig{Addr: v.GetString("redis.addr"), Password: v.GetString("redis.password"), DB: v.GetInt("redis.db")},
		JWT: JWTConfig{
			PrivateKeyPath: v.GetString("jwt.private_key_path"),
			PublicKeyPath:  v.GetString("jwt.public_key_path"),
			AccessTTL: accessTTL, RefreshTTL: refreshTTL,
			Issuer: v.GetString("jwt.issuer"), KeyID: v.GetString("jwt.key_id"),
		},
		OAuth2:    OAuth2Config{AuthCodeTTL: authCodeTTL},
		RateLimit: RateLimitConfig{LoginMaxAttempts: v.GetInt("rate_limit.login_max_attempts"), LoginWindowSecs: v.GetInt("rate_limit.login_window_secs"), LockDurationMins: v.GetInt("rate_limit.lock_duration_mins")},
	}, nil
}
