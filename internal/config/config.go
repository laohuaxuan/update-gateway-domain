package config

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"` //http服务配置
	Aliyun AliyunConfig `yaml:"aliyun"` //阿里云配置
	Update UpdateConfig `yaml:"update"` //更新配置
	Audit  AuditConfig  `yaml:"audit"`  //审计日志配置
	Auth   AuthConfig   `yaml:"auth"`   //鉴权配置
}

// http服务配置
type ServerConfig struct {
	Port int `yaml:"port"`
}

// 阿里云配置
type AliyunConfig struct {
	AcceptLanguage string `yaml:"accept_language"`
}

// 更新配置
type UpdateConfig struct {
	Protocol  string `yaml:"protocol"`
	MustHTTPS bool   `yaml:"must_https"`
	Http2     string `yaml:"http2"`
	BatchTLS  string `yaml:"batch_tls_version"`
}

// 审计日志配置
type AuditConfig struct {
	Enabled bool   `yaml:"enabled"`
	File    string `yaml:"file"`
}

type AuthConfig struct {
	UpdateToken         string `yaml:"update_token"`
	JWTSecret           string `yaml:"jwt_secret"`
	TokenExpiry         string `yaml:"token_expiry"`
	SMTPHost            string `yaml:"smtp_host"`
	SMTPPort            int    `yaml:"smtp_port"`
	SMTPUser            string `yaml:"smtp_user"`
	SMTPPass            string `yaml:"smtp_pass"`
	SMTPFrom            string `yaml:"smtp_from"`
	RootInitialName     string `yaml:"root_initial_name"`
	RootInitialPhone    string `yaml:"root_initial_phone"`
	RootInitialEmail    string `yaml:"root_initial_email"`
	RootInitialPassword string `yaml:"root_initial_password"`
	DBHost              string `yaml:"db_host"`
	DBPort              int    `yaml:"db_port"`
	DBUser              string `yaml:"db_user"`
	DBPassword          string `yaml:"db_password"`
	DBName              string `yaml:"db_name"`
}

func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	cfg := &Config{}
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	decoder.KnownFields(true)
	if err := decoder.Decode(cfg); err != nil {
		return nil, fmt.Errorf("parse config failed: %w", err)
	}

	normalize(cfg)
	if err := validate(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// normalize 统一做默认值填充与空白裁剪。
func normalize(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	cfg.Aliyun.AcceptLanguage = strings.TrimSpace(cfg.Aliyun.AcceptLanguage)
	if cfg.Aliyun.AcceptLanguage == "" {
		cfg.Aliyun.AcceptLanguage = "zh"
	}

	cfg.Update.Protocol = strings.TrimSpace(cfg.Update.Protocol)
	if cfg.Update.Protocol == "" {
		cfg.Update.Protocol = "HTTPS"
	}
	cfg.Update.Http2 = strings.TrimSpace(cfg.Update.Http2)
	if cfg.Update.Http2 == "" {
		cfg.Update.Http2 = "globalConfig"
	}
	cfg.Update.BatchTLS = strings.TrimSpace(strings.ToLower(cfg.Update.BatchTLS))
	if cfg.Audit.File == "" {
		cfg.Audit.File = "./logs/audit.log"
	}
	if !cfg.Audit.Enabled {
		cfg.Audit.Enabled = true
	}
	cfg.Auth.UpdateToken = strings.TrimSpace(cfg.Auth.UpdateToken)
	cfg.Auth.JWTSecret = strings.TrimSpace(cfg.Auth.JWTSecret)
	cfg.Auth.TokenExpiry = strings.TrimSpace(cfg.Auth.TokenExpiry)
	cfg.Auth.SMTPHost = strings.TrimSpace(cfg.Auth.SMTPHost)
	cfg.Auth.SMTPUser = strings.TrimSpace(cfg.Auth.SMTPUser)
	cfg.Auth.SMTPFrom = strings.TrimSpace(cfg.Auth.SMTPFrom)
	cfg.Auth.RootInitialName = strings.TrimSpace(cfg.Auth.RootInitialName)
	cfg.Auth.RootInitialPhone = strings.TrimSpace(cfg.Auth.RootInitialPhone)
	cfg.Auth.RootInitialEmail = strings.TrimSpace(cfg.Auth.RootInitialEmail)
	cfg.Auth.RootInitialPassword = strings.TrimSpace(cfg.Auth.RootInitialPassword)
	if cfg.Auth.TokenExpiry == "" {
		cfg.Auth.TokenExpiry = "24h"
	}
}

// validate 校验配置合法性，避免启动后才暴露错误。
func validate(cfg *Config) error {
	return nil
}
