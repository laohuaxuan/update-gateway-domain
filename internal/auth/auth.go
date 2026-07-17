package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"update-gateway-domain/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	store       *store.Store
	jwtSecret   []byte
	tokenExpiry time.Duration
	smtpHost    string
	smtpPort    int
	smtpUser    string
	smtpPass    string
	smtpFrom    string
}

type Config struct {
	JWTSecret           string        `yaml:"jwt_secret"`
	TokenExpiry         time.Duration `yaml:"token_expiry"`
	SMTPHost            string        `yaml:"smtp_host"`
	SMTPPort            int           `yaml:"smtp_port"`
	SMTPUser            string        `yaml:"smtp_user"`
	SMTPPass            string        `yaml:"smtp_pass"`
	SMTPFrom            string        `yaml:"smtp_from"`
	RootInitialEmail    string        `yaml:"root_initial_email"`
	RootInitialPhone    string        `yaml:"root_initial_phone"`
	RootInitialName     string        `yaml:"root_initial_name"`
	RootInitialPassword string        `yaml:"root_initial_password"`
}

var (
	ErrSMTPConfigInvalid = errors.New("smtp config invalid")
	ErrSMTPConnect       = errors.New("smtp connect failed")
	ErrSMTPAuth          = errors.New("smtp auth failed")
	ErrSMTPRecipient     = errors.New("smtp recipient rejected")
	ErrSMTPSend          = errors.New("smtp send failed")
)

func New(s *store.Store, cfg *Config) (*Auth, error) {
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("jwt_secret is required")
	}

	a := &Auth{
		store:       s,
		jwtSecret:   []byte(cfg.JWTSecret),
		tokenExpiry: cfg.TokenExpiry,
		smtpHost:    cfg.SMTPHost,
		smtpPort:    cfg.SMTPPort,
		smtpUser:    cfg.SMTPUser,
		smtpPass:    cfg.SMTPPass,
		smtpFrom:    cfg.SMTPFrom,
	}

	if a.tokenExpiry == 0 {
		a.tokenExpiry = 24 * time.Hour
	}

	if err := a.ensureRootUser(cfg); err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Auth) ensureRootUser(cfg *Config) error {
	user, err := a.store.GetUserByRole(store.RoleRoot)
	if err != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(cfg.RootInitialPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		newUser := &store.User{
			Name:         cfg.RootInitialName,
			Phone:        cfg.RootInitialPhone,
			Email:        cfg.RootInitialEmail,
			PasswordHash: string(hash),
			Role:         store.RoleRoot,
		}

		return a.store.CreateUser(newUser)
	}

	updates := make(map[string]interface{})
	if user.Phone != cfg.RootInitialPhone {
		existingPhoneUser, _ := a.store.GetUserByPhone(cfg.RootInitialPhone)
		if existingPhoneUser == nil || existingPhoneUser.ID == user.ID {
			updates["phone"] = cfg.RootInitialPhone
		}
	}
	if user.Email != cfg.RootInitialEmail {
		existingEmailUser, _ := a.store.GetUserByEmail(cfg.RootInitialEmail)
		if existingEmailUser == nil || existingEmailUser.ID == user.ID {
			updates["email"] = cfg.RootInitialEmail
		}
	}
	if len(updates) > 0 {
		if err := a.store.UpdateProfile(user.ID, updates); err != nil {
			return err
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.RootInitialPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return a.store.UpdatePassword(user.ID, string(hash))
}

func (a *Auth) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (a *Auth) ComparePassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (a *Auth) GenerateToken(user *store.User) (string, error) {
	claims := jwt.MapClaims{
		"id":   user.ID,
		"name": user.Name,
		"role": string(user.Role),
		"exp":  time.Now().Add(a.tokenExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.jwtSecret)
}

func (a *Auth) ParseToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return a.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func (a *Auth) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			c.Abort()
			return
		}

		token = strings.TrimPrefix(token, "Bearer ")

		claims, err := a.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", int64(claims["id"].(float64)))
		c.Set("user_role", claims["role"].(string))
		c.Next()
	}
}

func (a *Auth) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		a.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		role, _ := c.Get("user_role")
		if role != string(store.RoleAdmin) && role != string(store.RoleRoot) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin role required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (a *Auth) RequireRoot() gin.HandlerFunc {
	return func(c *gin.Context) {
		a.RequireAuth()(c)
		if c.IsAborted() {
			return
		}

		role, _ := c.Get("user_role")
		if role != string(store.RoleRoot) {
			c.JSON(http.StatusForbidden, gin.H{"error": "root role required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func (a *Auth) GenerateResetToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (a *Auth) SendPasswordResetEmail(to, resetURL string, ttl time.Duration) error {
	if err := a.sendMail(
		to,
		"MSE网关管理系统密码重置通知",
		fmt.Sprintf(
			"您好，\n\n我们收到了您的密码重置请求。请点击下面的链接设置新密码：\n%s\n\n该链接将在 %d 分钟后失效，且只能使用一次。\n如果这不是您的操作，请忽略本邮件并尽快检查账号安全。\n",
			resetURL,
			int(ttl.Minutes()),
		),
	); err != nil {
		return err
	}
	return nil
}

func (a *Auth) SendPasswordChangedEmail(to, ip, device string, changedAt time.Time) error {
	if err := a.sendMail(
		to,
		"MSE网关管理系统密码重置成功通知",
		fmt.Sprintf(
			"您好，\n\n您的账号密码已于 %s 修改成功。\n登录 IP：%s\n设备信息：%s\n\n如果这不是您的操作，请立即联系管理员。\n",
			changedAt.Format("2006-01-02 15:04:05"),
			ip,
			device,
		),
	); err != nil {
		return err
	}
	return nil
}

func (a *Auth) sendMail(to, subject, body string) error {
	if a.smtpHost == "" || a.smtpPort == 0 || a.smtpFrom == "" {
		return fmt.Errorf("%w: smtp_host/smtp_port/smtp_from required", ErrSMTPConfigInvalid)
	}

	headers := map[string]string{
		"From":         a.smtpFrom,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
	}

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	addr := fmt.Sprintf("%s:%d", a.smtpHost, a.smtpPort)
	if a.smtpPort == 465 {
		return a.sendMailWithTLS(addr, to, msg.Bytes())
	}

	var auth smtp.Auth
	if a.smtpUser != "" && a.smtpPass != "" {
		auth = smtp.PlainAuth("", a.smtpUser, a.smtpPass, a.smtpHost)
	}
	if err := smtp.SendMail(addr, auth, a.smtpFrom, []string{to}, msg.Bytes()); err != nil {
		return wrapSMTPError(err)
	}
	return nil
}

func (a *Auth) sendMailWithTLS(addr, to string, msg []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: a.smtpHost})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSMTPConnect, err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, a.smtpHost)
	if err != nil {
		return err
	}
	defer client.Close()

	if a.smtpUser != "" && a.smtpPass != "" {
		auth := smtp.PlainAuth("", a.smtpUser, a.smtpPass, a.smtpHost)
		if ok, _ := client.Extension("AUTH"); ok {
			if err := client.Auth(auth); err != nil {
				return wrapSMTPError(err)
			}
		}
	}

	if err := client.Mail(a.smtpFrom); err != nil {
		return wrapSMTPError(err)
	}
	if err := client.Rcpt(to); err != nil {
		return wrapSMTPError(err)
	}

	wc, err := client.Data()
	if err != nil {
		return wrapSMTPError(err)
	}
	if _, err := wc.Write(msg); err != nil {
		return wrapSMTPError(err)
	}
	if err := wc.Close(); err != nil {
		return wrapSMTPError(err)
	}
	return client.Quit()
}

func wrapSMTPError(err error) error {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "535"), strings.Contains(msg, "authentication failed"), strings.Contains(msg, "auth"):
		return fmt.Errorf("%w: %v", ErrSMTPAuth, err)
	case strings.Contains(msg, "550"), strings.Contains(msg, "553"), strings.Contains(msg, "rcpt"), strings.Contains(msg, "recipient"):
		return fmt.Errorf("%w: %v", ErrSMTPRecipient, err)
	case strings.Contains(msg, "connect"), strings.Contains(msg, "dial"), strings.Contains(msg, "timeout"), strings.Contains(msg, "refused"):
		return fmt.Errorf("%w: %v", ErrSMTPConnect, err)
	default:
		return fmt.Errorf("%w: %v", ErrSMTPSend, err)
	}
}

func ClientDevice(userAgent string) string {
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		return "unknown"
	}
	return ua
}

func ClientIP(remoteAddr string, forwarded string) string {
	xff := strings.TrimSpace(forwarded)
	if xff != "" {
		first := strings.TrimSpace(strings.Split(xff, ",")[0])
		if first != "" {
			return first
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		if remoteAddr == "" {
			return "unknown"
		}
		return remoteAddr
	}
	return host
}

func (a *Auth) GetStore() *store.Store {
	return a.store
}
