package httpapi

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"update-gateway-domain/internal/auth"
	"update-gateway-domain/internal/ldapauth"
	"update-gateway-domain/internal/service"
	"update-gateway-domain/internal/store"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var numericRegex = regexp.MustCompile(`^\d+$`)
var hasUppercaseRegex = regexp.MustCompile(`[A-Z]`)
var hasLowercaseRegex = regexp.MustCompile(`[a-z]`)
var hasDigitRegex = regexp.MustCompile(`[0-9]`)
var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

const ldapPasswordHint = "请在AD域修改账号密码"

func isNumeric(s string) bool {
	return numericRegex.MatchString(s)
}

func isPasswordValid(password string) bool {
	if len(password) < 6 {
		return false
	}
	return hasUppercaseRegex.MatchString(password) &&
		hasLowercaseRegex.MatchString(password) &&
		hasDigitRegex.MatchString(password)
}

type Handler struct {
	service *service.DomainService
	auth    *auth.Auth
	ldap    *ldapauth.Client
}

type ldapLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type updateOneRequest struct {
	AccountID  string `json:"account_id"`
	GatewayID  string `json:"gateway_id"`
	Domain     string `json:"domain"`
	Protocol   string `json:"protocol"`
	TLSVersion string `json:"tls_version"`
	CertID     string `json:"cert_identifier"`
	DryRun     bool   `json:"dry_run"`
	MustHTTPS  string `json:"must_https"`
	HTTP2      string `json:"http2"`
}

type batchUpdateRequest struct {
	AccountID  string `json:"account_id"`
	GatewayID  string `json:"gateway_id"`
	Protocol   string `json:"protocol"`
	TLSVersion string `json:"tls_version"`
	DryRun     bool   `json:"dry_run"`
	MustHTTPS  string `json:"must_https"`
	HTTP2      string `json:"http2"`
}

type registerRequest struct {
	Username     string `json:"username"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	CaptchaToken string `json:"captcha_token"`
	CaptchaCode  string `json:"captcha_code"`
}

type loginRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	CaptchaToken string `json:"captcha_token"`
	CaptchaCode  string `json:"captcha_code"`
}

type resetPasswordRequest struct {
	Account string `json:"account"`
}

type changePasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

type updateProfileRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
}

type updateOwnPasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type verifyOwnPasswordRequest struct {
	OldPassword string `json:"old_password"`
}

type addAccountRequest struct {
	Name            string `json:"name"`
	ID              string `json:"id"`
	RegionID        string `json:"region_id"`
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	AcceptLanguage  string `json:"accept_language"`
}

type addGatewayRequest struct {
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	ID        string `json:"id"`
	BatchTLS  string `json:"batch_tls_version"`
}

type createUserRequest struct {
	Username string `json:"username"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type adminResetPasswordRequest struct {
	UserID int64 `json:"user_id"`
}

func NewHandler(svc *service.DomainService, auth *auth.Auth, ldap *ldapauth.Client) *Handler {
	return &Handler{service: svc, auth: auth, ldap: ldap}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.Use(cors())

	r.GET("/api/health", h.health)

	authGroup := r.Group("/api")
	{
		authGroup.GET("/captcha", h.captchaImage)
		authGroup.GET("/check-exists", h.checkExists)
		authGroup.GET("/auth/providers", h.authProviders)
		authGroup.POST("/register", h.register)
		authGroup.POST("/login", h.login)
		authGroup.POST("/login/ldap", h.loginLDAP)
		authGroup.POST("/reset-password", h.requestResetPassword)
		authGroup.POST("/change-password", h.changePassword)
	}

	apiGroup := r.Group("/api")
	apiGroup.Use(h.auth.RequireAuth())
	{
		apiGroup.GET("/tls-versions", h.tlsVersions)
		apiGroup.GET("/accounts", h.accounts)
		apiGroup.GET("/gateways", h.gateways)
		apiGroup.GET("/domains", h.listDomains)
		apiGroup.GET("/user-info", h.getUserInfo)
		apiGroup.PUT("/user-profile", h.updateUserProfile)
		apiGroup.POST("/user-password/verify-old", h.verifyOwnPassword)
		apiGroup.PUT("/user-password", h.updateOwnPassword)

		adminGroup := apiGroup.Group("")
		adminGroup.Use(h.auth.RequireAdmin())
		{
			adminGroup.GET("/audit-logs", h.auditLogs)
			adminGroup.POST("/batch-update", h.batchUpdate)
			adminGroup.POST("/update-domain", h.updateOneDomain)
			adminGroup.GET("/gateway-certificates", h.gatewayCertificates)

			adminGroup.POST("/accounts", h.addAccount)
			adminGroup.GET("/accounts/all", h.getAllAccounts)
			adminGroup.DELETE("/accounts/:id", h.deleteAccount)

			adminGroup.POST("/gateways", h.addGateway)
			adminGroup.GET("/gateways/all", h.getAllGateways)
			adminGroup.DELETE("/gateways/:id", h.deleteGateway)
		}

		rootGroup := apiGroup.Group("")
		rootGroup.Use(h.auth.RequireRoot())
		{
			rootGroup.GET("/users", h.listUsers)
			rootGroup.POST("/users", h.createUser)
			rootGroup.PUT("/users/:id/role", h.updateUserRole)
			rootGroup.PUT("/users/:id/reset-password", h.adminResetPassword)
			rootGroup.DELETE("/users/:id", h.deleteUser)
		}
	}
}

func (h *Handler) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) getUserInfo(c *gin.Context) {
	userID, _ := c.Get("user_id")
	role, _ := c.Get("user_role")

	user, err := h.auth.GetStore().GetUserByID(userID.(int64))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"user_id": userID, "role": role})
		return
	}

	authSource := strings.TrimSpace(user.AuthSource)
	if authSource == "" {
		authSource = store.AuthSourceLocal
	}
	c.JSON(http.StatusOK, gin.H{
		"user_id":     userID,
		"name":        user.Name,
		"phone":       user.PhoneValue(),
		"email":       user.Email,
		"role":        role,
		"auth_source": authSource,
	})
}

func (h *Handler) updateUserProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	if err := h.auth.GetStore().UpdateProfile(userID.(int64), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}

func (h *Handler) updateOwnPassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req updateOwnPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req.OldPassword = strings.TrimSpace(req.OldPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	req.ConfirmPassword = strings.TrimSpace(req.ConfirmPassword)

	if req.OldPassword == "" || req.NewPassword == "" || req.ConfirmPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请完整填写旧密码、新密码和确认密码"})
		return
	}
	if !isPasswordValid(req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码不符合规范（需包含大小写字母、数字，且不少于6位）"})
		return
	}
	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码不一致"})
		return
	}

	user, err := h.auth.GetStore().GetUserByID(userID.(int64))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	if user.IsLDAP() {
		c.JSON(http.StatusBadRequest, gin.H{"error": ldapPasswordHint})
		return
	}
	if !h.auth.ComparePassword(user.PasswordHash, req.OldPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码不正确"})
		return
	}
	if h.auth.ComparePassword(user.PasswordHash, req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码不能与旧密码相同"})
		return
	}

	hash, err := h.auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}
	if err := h.auth.GetStore().UpdatePassword(user.ID, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

func (h *Handler) verifyOwnPassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req verifyOwnPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req.OldPassword = strings.TrimSpace(req.OldPassword)
	if req.OldPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "old_password is required"})
		return
	}

	user, err := h.auth.GetStore().GetUserByID(userID.(int64))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	if user.IsLDAP() {
		c.JSON(http.StatusBadRequest, gin.H{"error": ldapPasswordHint})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": h.auth.ComparePassword(user.PasswordHash, req.OldPassword)})
}

func (h *Handler) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req.CaptchaToken = strings.TrimSpace(req.CaptchaToken)
	req.CaptchaCode = strings.TrimSpace(req.CaptchaCode)
	if req.CaptchaToken == "" || req.CaptchaCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "captcha token and code are required"})
		return
	}
	if len(req.CaptchaCode) != 4 || !isNumeric(req.CaptchaCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "captcha code must be 4 digits"})
		return
	}
	if !h.verifyCaptcha(req.CaptchaToken, req.CaptchaCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired captcha"})
		return
	}

	if req.Username == "" || req.Phone == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username, phone, email, password are required"})
		return
	}

	if len(req.Username) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username must be at least 3 characters"})
		return
	}

	if len(req.Username) > 24 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username must be at most 24 characters"})
		return
	}

	adminBlacklist := []string{"root", "admin", "administrator", "superadmin", "super_admin", "admin_root", "root_admin", "system", "adminstrator"}
	lowerUsername := strings.ToLower(req.Username)
	for _, name := range adminBlacklist {
		if lowerUsername == name {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该用户名不允许注册"})
			return
		}
	}

	if !isPasswordValid(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码不符合规范（需包含大小写字母、数字，且不少于6位）"})
		return
	}

	_, err := h.auth.GetStore().GetUserByName(req.Username)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already registered"})
		return
	}

	_, err = h.auth.GetStore().GetUserByEmail(req.Email)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already registered"})
		return
	}

	_, err = h.auth.GetStore().GetUserByPhone(req.Phone)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone already registered"})
		return
	}

	hash, err := h.auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	user := &store.User{
		Name:         req.Username,
		Phone:        store.PhonePtr(req.Phone),
		Email:        req.Email,
		PasswordHash: hash,
		AuthSource:   store.AuthSourceLocal,
		Role:         store.RoleWatcher,
	}

	if err := h.auth.GetStore().CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user registered successfully"})
}

func (h *Handler) authProviders(c *gin.Context) {
	providers := make([]gin.H, 0, 1)
	if h.ldap != nil && h.ldap.Enabled() {
		providers = append(providers, gin.H{
			"id":    "ldap",
			"type":  "ldap",
			"label": h.ldap.Label(),
		})
	}
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

func (h *Handler) loginLDAP(c *gin.Context) {
	if h.ldap == nil || !h.ldap.Enabled() {
		c.JSON(http.StatusNotFound, gin.H{"error": "LDAP 登录未启用"})
		return
	}
	var req ldapLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	info, err := h.ldap.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	user, _, err := h.ensureLDAPUser(info)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	token, err := h.auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token":       token,
		"role":        string(user.Role),
		"name":        user.Name,
		"phone":       user.PhoneValue(),
		"email":       user.Email,
		"auth_source": store.AuthSourceLDAP,
	})
}

// ensureLDAPUser 按 LDAP 用户名匹配本地账号；不存在则创建观察者。
func (h *Handler) ensureLDAPUser(info *ldapauth.UserInfo) (*store.User, bool, error) {
	username := strings.TrimSpace(info.Username)
	if username == "" {
		return nil, false, fmt.Errorf("LDAP 用户名无效")
	}
	st := h.auth.GetStore()

	user, err := st.GetUserByName(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 软删除用户仍占用唯一键时，恢复并复用，避免 Duplicate entry
		if deleted, e := st.GetUserByNameUnscoped(username); e == nil && deleted != nil && deleted.DeletedAt.Valid {
			if restoreErr := st.RestoreUser(deleted.ID); restoreErr != nil {
				return nil, false, restoreErr
			}
			user = deleted
			err = nil
		}
	}
	if err == nil {
		if syncErr := h.syncLDAPUserFields(user, info, username); syncErr != nil {
			return nil, false, syncErr
		}
		user, _ = st.GetUserByName(username)
		return user, false, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err
	}

	email := resolveLDAPEmail(info.Email, username)
	if existing, e := st.GetUserByEmail(email); e == nil && existing != nil {
		email = ldapPlaceholderEmail(username + "." + fmt.Sprintf("%d", time.Now().Unix()%100000))
	}

	randPwd := generateRandomPassword(16)
	hash, err := h.auth.HashPassword(randPwd)
	if err != nil {
		return nil, false, fmt.Errorf("hash password failed: %w", err)
	}
	newUser := &store.User{
		Name:         username,
		Phone:        nil,
		Email:        email,
		PasswordHash: hash,
		AuthSource:   store.AuthSourceLDAP,
		Role:         store.RoleWatcher,
	}
	if err := st.CreateUser(newUser); err != nil {
		// 并发创建或软删竞态：回查已有账号并复用
		if existing, e := st.GetUserByName(username); e == nil && existing != nil {
			if syncErr := h.syncLDAPUserFields(existing, info, username); syncErr != nil {
				return nil, false, syncErr
			}
			existing, _ = st.GetUserByName(username)
			return existing, false, nil
		}
		if deleted, e := st.GetUserByNameUnscoped(username); e == nil && deleted != nil {
			if deleted.DeletedAt.Valid {
				_ = st.RestoreUser(deleted.ID)
			}
			if syncErr := h.syncLDAPUserFields(deleted, info, username); syncErr != nil {
				return nil, false, syncErr
			}
			deleted, _ = st.GetUserByName(username)
			return deleted, false, nil
		}
		return nil, false, err
	}
	return newUser, true, nil
}

func (h *Handler) syncLDAPUserFields(user *store.User, info *ldapauth.UserInfo, username string) error {
	updates := map[string]interface{}{}
	if !user.IsLDAP() {
		updates["auth_source"] = store.AuthSourceLDAP
	}
	if email := resolveLDAPEmail(info.Email, username); email != "" && email != user.Email {
		if shouldSyncLDAPEmail(user.Email, email) {
			if existing, e := h.auth.GetStore().GetUserByEmail(email); e != nil || existing == nil || existing.ID == user.ID {
				updates["email"] = email
			}
		}
	}
	if len(updates) == 0 {
		return nil
	}
	return h.auth.GetStore().UpdateUserFields(user.ID, updates)
}

func ldapPlaceholderEmail(username string) string {
	safe := strings.ToLower(username)
	var b strings.Builder
	for _, r := range safe {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	local := b.String()
	if local == "" {
		local = "ldapuser"
	}
	if len(local) > 64 {
		local = local[:64]
	}
	return local + "@ldap.local"
}

func resolveLDAPEmail(raw, username string) string {
	email := strings.TrimSpace(raw)
	if email != "" && emailRegex.MatchString(email) {
		return strings.ToLower(email)
	}
	return ldapPlaceholderEmail(username)
}

func shouldSyncLDAPEmail(current, fromLDAP string) bool {
	current = strings.TrimSpace(strings.ToLower(current))
	fromLDAP = strings.TrimSpace(strings.ToLower(fromLDAP))
	if fromLDAP == "" || fromLDAP == current {
		return false
	}
	if strings.HasSuffix(current, "@ldap.local") {
		return true
	}
	return current == ""
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	req.CaptchaToken = strings.TrimSpace(req.CaptchaToken)
	req.CaptchaCode = strings.TrimSpace(req.CaptchaCode)
	if req.CaptchaToken == "" || req.CaptchaCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "captcha token and code are required"})
		return
	}
	if len(req.CaptchaCode) != 4 || !isNumeric(req.CaptchaCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "captcha code must be 4 digits"})
		return
	}
	if !h.verifyCaptcha(req.CaptchaToken, req.CaptchaCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired captcha"})
		return
	}

	var user *store.User
	var err error
	if len(req.Username) == 11 && isNumeric(req.Username) {
		user, err = h.auth.GetStore().GetUserByPhone(req.Username)
	} else {
		user, err = h.auth.GetStore().GetUserByName(req.Username)
	}
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if user.IsLDAP() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "该账号为 LDAP 用户，请使用 AD 登录"})
		return
	}

	if !h.auth.ComparePassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := h.auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	authSource := strings.TrimSpace(user.AuthSource)
	if authSource == "" {
		authSource = store.AuthSourceLocal
	}
	c.JSON(http.StatusOK, gin.H{
		"token":       token,
		"role":        string(user.Role),
		"name":        user.Name,
		"phone":       user.PhoneValue(),
		"email":       user.Email,
		"auth_source": authSource,
	})
}

func (h *Handler) checkExists(c *gin.Context) {
	field := c.Query("field")
	value := c.Query("value")

	if field == "" || value == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var exists bool
	var err error

	switch field {
	case "username":
		_, err = h.auth.GetStore().GetUserByName(value)
	case "phone":
		_, err = h.auth.GetStore().GetUserByPhone(value)
	case "email":
		_, err = h.auth.GetStore().GetUserByEmail(value)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid field"})
		return
	}

	exists = err == nil
	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

func (h *Handler) requestResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	account := strings.TrimSpace(req.Account)
	if account == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account is required"})
		return
	}

	var user *store.User
	var err error
	if len(account) == 11 && isNumeric(account) {
		user, err = h.auth.GetStore().GetUserByPhone(account)
	} else {
		user, err = h.auth.GetStore().GetUserByName(account)
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "重置密码已发送到您的注册邮箱xxxx@xxx.xx"})
		return
	}
	if user.IsLDAP() {
		c.JSON(http.StatusBadRequest, gin.H{"error": ldapPasswordHint})
		return
	}

	requestCount, err := h.auth.GetStore().CountPasswordResetRequestsSince(user.ID, time.Now().Add(-1*time.Hour))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send reset link"})
		return
	}
	if requestCount >= 5 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many reset requests, please try again later"})
		return
	}

	if err := h.auth.GetStore().DeleteResetToken(user.ID); err != nil {
		fmt.Printf("[WARN] delete reset token failed: user_id=%d err=%v\n", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send reset link"})
		return
	}

	token := h.auth.GenerateResetToken()
	expireDuration := 30 * time.Minute
	expiresAt := time.Now().Add(expireDuration).Unix()
	clientIP := c.ClientIP()
	clientDevice := c.Request.UserAgent()

	resetToken := &store.ResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	if err := h.auth.GetStore().SaveResetToken(resetToken); err != nil {
		fmt.Printf("[WARN] save reset token failed: user_id=%d err=%v\n", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send reset link"})
		return
	}

	if err := h.auth.GetStore().CreatePasswordResetRequestLog(&store.PasswordResetRequestLog{
		UserID: user.ID,
		Email:  user.Email,
		IP:     clientIP,
		Device: clientDevice,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to send reset link"})
		return
	}

	_ = h.auth.GetStore().CreatePasswordSecurityAuditLog(&store.PasswordSecurityAuditLog{
		UserID: user.ID,
		Email:  user.Email,
		Event:  "password_reset_request",
		IP:     clientIP,
		Device: clientDevice,
		Detail: "password reset link requested",
	})

	resetURL := buildResetPasswordURL(c, token)
	if err := h.auth.SendPasswordResetEmail(user.Email, resetURL, expireDuration); err != nil {
		fmt.Printf("[WARN] send reset email failed: user_id=%d email=%s err=%v\n", user.ID, user.Email, err)
		switch {
		case errors.Is(err, auth.ErrSMTPConfigInvalid):
			c.JSON(http.StatusInternalServerError, gin.H{"error": "mail service is not configured"})
		case errors.Is(err, auth.ErrSMTPConnect):
			c.JSON(http.StatusBadGateway, gin.H{"error": "mail service is unavailable"})
		case errors.Is(err, auth.ErrSMTPAuth):
			c.JSON(http.StatusBadGateway, gin.H{"error": "mail service authentication failed"})
		case errors.Is(err, auth.ErrSMTPRecipient):
			c.JSON(http.StatusBadRequest, gin.H{"error": "registered email is unavailable"})
		default:
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to send reset email"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("重置密码已发送到您的注册邮箱%s", user.Email)})
}

func (h *Handler) changePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if !isPasswordValid(req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码不符合规范（需包含大小写字母、数字，且不少于6位）"})
		return
	}

	resetToken, err := h.auth.GetStore().GetResetToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired token"})
		return
	}

	if time.Now().Unix() > resetToken.ExpiresAt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token expired"})
		return
	}

	user, err := h.auth.GetStore().GetUserByID(resetToken.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
		return
	}
	if user.IsLDAP() {
		c.JSON(http.StatusBadRequest, gin.H{"error": ldapPasswordHint})
		return
	}
	if h.auth.ComparePassword(user.PasswordHash, req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must be different from current password"})
		return
	}

	hash, err := h.auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	if err := h.auth.GetStore().UpdatePassword(resetToken.UserID, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	if err := h.auth.GetStore().DeleteResetTokenByToken(req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	clientIP := c.ClientIP()
	clientDevice := c.Request.UserAgent()
	changedAt := time.Now()
	_ = h.auth.GetStore().CreatePasswordSecurityAuditLog(&store.PasswordSecurityAuditLog{
		UserID: user.ID,
		Email:  user.Email,
		Event:  "password_changed",
		IP:     clientIP,
		Device: clientDevice,
		Detail: "password changed via reset link",
	})
	_ = h.auth.SendPasswordChangedEmail(user.Email, clientIP, clientDevice, changedAt)

	c.JSON(http.StatusOK, gin.H{"message": "password updated successfully"})
}

func buildResetPasswordURL(c *gin.Context, token string) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if proto := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); proto != "" {
		scheme = strings.ToLower(strings.Split(proto, ",")[0])
	}
	host := c.Request.Host
	if host == "" {
		host = "localhost:8080"
	}
	return fmt.Sprintf("%s://%s/reset-password?token=%s", scheme, host, url.QueryEscape(token))
}

func (h *Handler) listUsers(c *gin.Context) {
	users, err := h.auth.GetStore().ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, 0, len(users))
	for _, u := range users {
		authSource := strings.TrimSpace(u.AuthSource)
		if authSource == "" {
			authSource = store.AuthSourceLocal
		}
		result = append(result, gin.H{
			"id":          u.ID,
			"name":        u.Name,
			"phone":       u.PhoneValue(),
			"email":       u.Email,
			"role":        string(u.Role),
			"auth_source": authSource,
		})
	}

	c.JSON(http.StatusOK, gin.H{"users": result})
}

func (h *Handler) createUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Username == "" || req.Phone == "" || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username, phone, email are required"})
		return
	}

	if len(req.Username) < 3 || len(req.Username) > 24 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username must be between 3 and 24 characters"})
		return
	}

	adminBlacklist := []string{"root", "admin", "administrator", "superadmin", "super_admin", "admin_root", "root_admin", "system", "adminstrator"}
	lowerUsername := strings.ToLower(req.Username)
	for _, name := range adminBlacklist {
		if lowerUsername == name {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该用户名不允许创建"})
			return
		}
	}

	_, err := h.auth.GetStore().GetUserByName(req.Username)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username already exists"})
		return
	}

	_, err = h.auth.GetStore().GetUserByEmail(req.Email)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email already exists"})
		return
	}

	_, err = h.auth.GetStore().GetUserByPhone(req.Phone)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone already exists"})
		return
	}

	role := store.Role(req.Role)
	if role == "" {
		role = store.RoleWatcher
	}
	if role != store.RoleAdmin && role != store.RoleWatcher {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	randomPassword := generateRandomPassword(8)
	hash, err := h.auth.HashPassword(randomPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	user := &store.User{
		Name:         req.Username,
		Phone:        store.PhonePtr(req.Phone),
		Email:        req.Email,
		PasswordHash: hash,
		AuthSource:   store.AuthSourceLocal,
		Role:         role,
	}

	if err := h.auth.GetStore().CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "user created successfully",
		"password": randomPassword,
	})
}

func generateRandomPassword(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func (h *Handler) updateUserRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	role := store.Role(req.Role)
	if role != store.RoleAdmin && role != store.RoleWatcher {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	user, err := h.auth.GetStore().GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.Role == store.RoleRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot change root role"})
		return
	}

	if err := h.auth.GetStore().UpdateUserRole(id, role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated successfully"})
}

func (h *Handler) adminResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.auth.GetStore().GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.Role == store.RoleRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot reset root password"})
		return
	}
	if user.IsLDAP() {
		c.JSON(http.StatusBadRequest, gin.H{"error": ldapPasswordHint})
		return
	}

	randomPassword := generateRandomPassword(8)
	hash, err := h.auth.HashPassword(randomPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		return
	}

	if err := h.auth.GetStore().UpdatePassword(id, hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "password reset successfully",
		"password": randomPassword,
	})
}

func (h *Handler) deleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.auth.GetStore().GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	if user.Role == store.RoleRoot {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete root user"})
		return
	}

	if err := h.auth.GetStore().DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
}

func (h *Handler) tlsVersions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"versions": service.SupportedTLSVersions(),
	})
}

func (h *Handler) listDomains(c *gin.Context) {
	accountID := c.Query("account_id")
	gatewayID := c.Query("gateway_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	searchType := c.Query("search_type")
	searchKeyword := c.Query("search_keyword")
	fmt.Printf("[DEBUG] listDomains: accountID=%s, gatewayID=%s, page=%d, size=%d, searchType=%s, searchKeyword=%s\n", accountID, gatewayID, page, size, searchType, searchKeyword)
	result, err := h.service.ListDomains(c.Request.Context(), accountID, gatewayID, page, size, searchType, searchKeyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) accounts(c *gin.Context) {
	accounts, err := h.service.ListAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"accounts": accounts})
}

func (h *Handler) gateways(c *gin.Context) {
	accountID := c.Query("account_id")
	gateways, err := h.service.ListGateways(accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"gateways": gateways})
}

func (h *Handler) gatewayCertificates(c *gin.Context) {
	accountID := c.Query("account_id")
	gatewayID := c.Query("gateway_id")
	certs, err := h.service.ListGatewayCertificates(c.Request.Context(), accountID, gatewayID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"certificates": certs})
}

func (h *Handler) auditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	result, err := h.service.QueryAuditLogs(service.AuditQuery{
		Page:      page,
		PageSize:  size,
		AccountID: c.Query("account_id"),
		GatewayID: c.Query("gateway_id"),
		Action:    c.Query("action"),
		Result:    c.Query("result"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) batchUpdate(c *gin.Context) {
	req := &batchUpdateRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.service.BatchUpdate(c.Request.Context(), req.AccountID, req.GatewayID, req.Protocol, req.TLSVersion, req.DryRun, req.MustHTTPS, req.HTTP2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) updateOneDomain(c *gin.Context) {
	req := &updateOneRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	result, err := h.service.UpdateOneDomain(c.Request.Context(), req.AccountID, req.GatewayID, req.Domain, req.Protocol, req.TLSVersion, req.CertID, req.DryRun, req.MustHTTPS, req.HTTP2)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) addAccount(c *gin.Context) {
	var req addAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Name == "" || req.ID == "" || req.RegionID == "" || req.AccessKeyID == "" || req.AccessKeySecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, id, region_id, access_key_id, access_key_secret are required"})
		return
	}

	account := &store.Account{
		Name:            req.Name,
		ID:              req.ID,
		RegionID:        req.RegionID,
		AccessKeyID:     req.AccessKeyID,
		AccessKeySecret: req.AccessKeySecret,
		AcceptLanguage:  req.AcceptLanguage,
	}

	if err := h.auth.GetStore().CreateAccount(account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account added successfully"})
}

func (h *Handler) getAllAccounts(c *gin.Context) {
	accounts, err := h.auth.GetStore().ListAccounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, 0, len(accounts))
	for _, acct := range accounts {
		result = append(result, gin.H{
			"id":              acct.ID,
			"name":            acct.Name,
			"region_id":       acct.RegionID,
			"access_key_id":   acct.AccessKeyID,
			"accept_language": acct.AcceptLanguage,
		})
	}

	c.JSON(http.StatusOK, gin.H{"accounts": result})
}

func (h *Handler) deleteAccount(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account id is required"})
		return
	}

	if err := h.auth.GetStore().DeleteAccount(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.service.RemoveAccountClient(id)

	c.JSON(http.StatusOK, gin.H{"message": "account deleted successfully"})
}

func (h *Handler) addGateway(c *gin.Context) {
	var req addGatewayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.AccountID == "" || req.Name == "" || req.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id, name, id are required"})
		return
	}

	gateway := &store.Gateway{
		AccountID: req.AccountID,
		Name:      req.Name,
		ID:        req.ID,
		BatchTLS:  req.BatchTLS,
	}

	if err := h.auth.GetStore().CreateGateway(gateway); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "gateway added successfully"})
}

func (h *Handler) getAllGateways(c *gin.Context) {
	gateways, err := h.auth.GetStore().ListGateways("")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := make([]gin.H, 0, len(gateways))
	for _, gw := range gateways {
		result = append(result, gin.H{
			"id":         gw.ID,
			"account_id": gw.AccountID,
			"name":       gw.Name,
			"batch_tls":  gw.BatchTLS,
		})
	}

	c.JSON(http.StatusOK, gin.H{"gateways": result})
}

func (h *Handler) deleteGateway(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "gateway id is required"})
		return
	}

	if err := h.auth.GetStore().DeleteGateway(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "gateway deleted successfully"})
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Update-Token")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
