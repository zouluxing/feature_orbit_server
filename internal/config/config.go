package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port        int
	Debug       bool
	DatabaseURL string
	UMSBaseURL  string
	UMSCacheTTL time.Duration
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.SetDefault("port", 8080)
	v.SetDefault("debug", true)
	v.SetDefault("ums_base_url", "http://localhost:8081")
	v.SetDefault("ums_cache_ttl", "1h")
	cacheTTL, _ := time.ParseDuration(v.GetString("ums_cache_ttl"))
	return &Config{
		Port: v.GetInt("port"), Debug: v.GetBool("debug"),
		DatabaseURL: v.GetString("database_url"),
		UMSBaseURL: v.GetString("ums_base_url"), UMSCacheTTL: cacheTTL,
	}, nil
}
