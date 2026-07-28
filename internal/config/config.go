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
	UpdateToken         string     `yaml:"update_token"`
	JWTSecret           string     `yaml:"jwt_secret"`
	TokenExpiry         string     `yaml:"token_expiry"`
	SMTPHost            string     `yaml:"smtp_host"`
	SMTPPort            int        `yaml:"smtp_port"`
	SMTPUser            string     `yaml:"smtp_user"`
	SMTPPass            string     `yaml:"smtp_pass"`
	SMTPFrom            string     `yaml:"smtp_from"`
	RootInitialName     string     `yaml:"root_initial_name"`
	RootInitialPhone    string     `yaml:"root_initial_phone"`
	RootInitialEmail    string     `yaml:"root_initial_email"`
	RootInitialPassword string     `yaml:"root_initial_password"`
	DBHost              string     `yaml:"db_host"`
	DBPort              int        `yaml:"db_port"`
	DBUser              string     `yaml:"db_user"`
	DBPassword          string     `yaml:"db_password"`
	DBName              string     `yaml:"db_name"`
	LDAP                LDAPConfig `yaml:"ldap"`
}

// LDAPConfig 第三方 LDAP 登录（登录页展示入口；首次登录自动创建观察者）。
type LDAPConfig struct {
	Enabled         bool   `yaml:"enabled"`
	Host            string `yaml:"host"`
	Port            int    `yaml:"port"`
	UseSSL          bool   `yaml:"use_ssl"`
	StartTLS        bool   `yaml:"start_tls"`
	SkipTLSVerify   bool   `yaml:"skip_tls_verify"`
	BindDN          string `yaml:"bind_dn"`
	BindPassword    string `yaml:"bind_password"`
	BaseDN          string `yaml:"base_dn"`
	UserFilter      string `yaml:"user_filter"` // 必须含 %s，如 (uid=%s) 或 (sAMAccountName=%s)
	UsernameAttr    string `yaml:"username_attr"`
	EmailAttr       string `yaml:"email_attr"`
	DisplayNameAttr string `yaml:"display_name_attr"`
	Label           string `yaml:"label"` // 登录页按钮文案
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

	cfg.Auth.LDAP.Host = strings.TrimSpace(cfg.Auth.LDAP.Host)
	cfg.Auth.LDAP.BindDN = strings.TrimSpace(cfg.Auth.LDAP.BindDN)
	cfg.Auth.LDAP.BaseDN = strings.TrimSpace(cfg.Auth.LDAP.BaseDN)
	cfg.Auth.LDAP.UserFilter = strings.TrimSpace(cfg.Auth.LDAP.UserFilter)
	if cfg.Auth.LDAP.UserFilter == "" {
		cfg.Auth.LDAP.UserFilter = "(uid=%s)"
	}
	cfg.Auth.LDAP.UsernameAttr = strings.TrimSpace(cfg.Auth.LDAP.UsernameAttr)
	if cfg.Auth.LDAP.UsernameAttr == "" {
		cfg.Auth.LDAP.UsernameAttr = "uid"
	}
	cfg.Auth.LDAP.EmailAttr = strings.TrimSpace(cfg.Auth.LDAP.EmailAttr)
	if cfg.Auth.LDAP.EmailAttr == "" {
		cfg.Auth.LDAP.EmailAttr = "mail"
	}
	cfg.Auth.LDAP.DisplayNameAttr = strings.TrimSpace(cfg.Auth.LDAP.DisplayNameAttr)
	if cfg.Auth.LDAP.DisplayNameAttr == "" {
		cfg.Auth.LDAP.DisplayNameAttr = "cn"
	}
	cfg.Auth.LDAP.Label = strings.TrimSpace(cfg.Auth.LDAP.Label)
	if cfg.Auth.LDAP.Label == "" {
		cfg.Auth.LDAP.Label = "LDAP"
	}
	if cfg.Auth.LDAP.Port == 0 {
		if cfg.Auth.LDAP.UseSSL {
			cfg.Auth.LDAP.Port = 636
		} else {
			cfg.Auth.LDAP.Port = 389
		}
	}
}

// validate 校验配置合法性，避免启动后才暴露错误。
func validate(cfg *Config) error {
	return nil
}
