package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"update-gateway-domain/internal/audit"
	"update-gateway-domain/internal/auth"
	"update-gateway-domain/internal/config"
	"update-gateway-domain/internal/httpapi"
	"update-gateway-domain/internal/ldapauth"
	"update-gateway-domain/internal/mse"
	"update-gateway-domain/internal/service"
	"update-gateway-domain/internal/store"
)

func main() {
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	if err := createDatabaseIfNotExist(cfg); err != nil {
		log.Fatalf("create database failed: %v", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Auth.DBUser,
		cfg.Auth.DBPassword,
		cfg.Auth.DBHost,
		cfg.Auth.DBPort,
		cfg.Auth.DBName,
	)

	db, err := store.New(dsn)
	if err != nil {
		log.Fatalf("open database failed: %v", err)
	}

	tokenExpiry, err := time.ParseDuration(cfg.Auth.TokenExpiry)
	if err != nil {
		log.Fatalf("parse token_expiry failed: %v", err)
	}

	authService, err := auth.New(db, &auth.Config{
		JWTSecret:           cfg.Auth.JWTSecret,
		TokenExpiry:         tokenExpiry,
		SMTPHost:            cfg.Auth.SMTPHost,
		SMTPPort:            cfg.Auth.SMTPPort,
		SMTPUser:            cfg.Auth.SMTPUser,
		SMTPPass:            cfg.Auth.SMTPPass,
		SMTPFrom:            cfg.Auth.SMTPFrom,
		RootInitialName:     cfg.Auth.RootInitialName,
		RootInitialPhone:    cfg.Auth.RootInitialPhone,
		RootInitialEmail:    cfg.Auth.RootInitialEmail,
		RootInitialPassword: cfg.Auth.RootInitialPassword,
	})
	if err != nil {
		log.Fatalf("init auth service failed: %v", err)
	}

	clientManager := mse.NewClientManager(db)

	var auditLogger *audit.Logger
	if cfg.Audit.Enabled {
		auditLogger = audit.NewLogger(db)
	}

	domainService := service.NewDomainService(clientManager, cfg, auditLogger, db)
	ldapClient := ldapauth.New(cfg.Auth.LDAP)
	handler := httpapi.NewHandler(domainService, authService, ldapClient)

	router := gin.Default()
	handler.RegisterRoutes(router)
	registerFrontend(router)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("server started at %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("start server failed: %v", err)
	}
}

func createDatabaseIfNotExist(cfg *config.Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Auth.DBUser,
		cfg.Auth.DBPassword,
		cfg.Auth.DBHost,
		cfg.Auth.DBPort,
	)

	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	_, err = sqlDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", cfg.Auth.DBName))
	return err
}

func registerFrontend(router *gin.Engine) {
	distPath := "frontend/dist"
	if _, err := os.Stat(distPath); err != nil {
		return
	}

	router.Static("/assets", distPath+"/assets")
	router.GET("/", func(c *gin.Context) {
		c.File(distPath + "/index.html")
	})
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(404, gin.H{"error": "not found"})
			return
		}
		c.File(distPath + "/index.html")
	})
}
