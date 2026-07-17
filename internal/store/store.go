package store

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Role string

const (
	RoleRoot    Role = "root"
	RoleAdmin   Role = "admin"
	RoleWatcher Role = "watcher"
)

type User struct {
	ID           int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	Name         string         `gorm:"uniqueIndex;size:100"`
	Phone        string         `gorm:"uniqueIndex;size:20"`
	Email        string         `gorm:"uniqueIndex;size:100"`
	PasswordHash string         `gorm:"size:255"`
	Role         Role           `gorm:"size:20;default:watcher"`
}

type ResetToken struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	UserID    int64          `gorm:"uniqueIndex"`
	Token     string         `gorm:"uniqueIndex;size:64"`
	ExpiresAt int64          `gorm:"index"`
}

type PasswordResetRequestLog struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UserID    int64  `gorm:"index"`
	Email     string `gorm:"size:100;index"`
	IP        string `gorm:"size:64"`
	Device    string `gorm:"size:512"`
}

type PasswordSecurityAuditLog struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	UserID    int64  `gorm:"index"`
	Email     string `gorm:"size:100;index"`
	Event     string `gorm:"size:50;index"`
	IP        string `gorm:"size:64"`
	Device    string `gorm:"size:512"`
	Detail    string `gorm:"size:1000"`
}

type AuditLog struct {
	ID        int64 `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time
	AccountID string `gorm:"size:100;index"`
	Action    string `gorm:"size:50;index"`

	GatewayID       string `gorm:"size:100;index"`
	Domain          string `gorm:"size:255"`
	CertBefore      string `gorm:"size:255"`
	CertAfter       string `gorm:"size:255"`
	Protocol        string `gorm:"size:20"`
	ProtocolBefore  string `gorm:"size:20"`
	ProtocolAfter   string `gorm:"size:20"`
	TLSVersion      string `gorm:"size:20"`
	TLSBefore       string `gorm:"size:20"`
	TLSAfter        string `gorm:"size:20"`
	MustHTTPS       bool
	MustHTTPSBefore *bool
	MustHTTPSAfter  *bool
	HTTP2           string `gorm:"size:20"`
	HTTP2Before     string `gorm:"size:20"`
	HTTP2After      string `gorm:"size:20"`
	DryRun          bool
	Result          string `gorm:"size:20;index"`
	Message         string `gorm:"size:1000"`
}

type Account struct {
	ID              string `gorm:"primaryKey;size:100"`
	Name            string `gorm:"size:100"`
	RegionID        string `gorm:"size:50"`
	AccessKeyID     string `gorm:"size:100"`
	AccessKeySecret string `gorm:"size:200"`
	AcceptLanguage  string `gorm:"size:20;default:zh-CN"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"index"`
}

type Gateway struct {
	ID        string `gorm:"primaryKey;size:100"`
	AccountID string `gorm:"size:100;index"`
	Name      string `gorm:"size:100"`
	BatchTLS  string `gorm:"size:20"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Store struct {
	db *gorm.DB
}

func New(dsn string) (*Store, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database failed: %w", err)
	}

	if err := db.AutoMigrate(&User{}, &ResetToken{}, &PasswordResetRequestLog{}, &PasswordSecurityAuditLog{}, &AuditLog{}, &Account{}, &Gateway{}); err != nil {
		return nil, fmt.Errorf("migrate database failed: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) CreateUser(user *User) error {
	return s.db.Create(user).Error
}

func (s *Store) GetUserByID(id int64) (*User, error) {
	var user User
	if err := s.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByPhone(phone string) (*User, error) {
	var user User
	if err := s.db.Where("phone = ?", phone).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByName(name string) (*User, error) {
	var user User
	if err := s.db.Where("name = ?", name).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) GetUserByRole(role Role) (*User, error) {
	var user User
	if err := s.db.Where("role = ?", role).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Store) UpdateUserRole(id int64, role Role) error {
	return s.db.Model(&User{}).Where("id = ?", id).Update("role", role).Error
}

func (s *Store) DeleteUser(id int64) error {
	return s.db.Delete(&User{}, id).Error
}

func (s *Store) UpdateProfile(id int64, updates map[string]interface{}) error {
	return s.db.Model(&User{}).Where("id = ?", id).Updates(updates).Error
}

func (s *Store) ListUsers() ([]User, error) {
	var users []User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Store) SaveResetToken(token *ResetToken) error {
	return s.db.Create(token).Error
}

func (s *Store) GetResetToken(token string) (*ResetToken, error) {
	var rt ResetToken
	if err := s.db.Where("token = ?", token).First(&rt).Error; err != nil {
		return nil, err
	}
	return &rt, nil
}

func (s *Store) DeleteResetToken(userID int64) error {
	return s.db.Unscoped().Where("user_id = ?", userID).Delete(&ResetToken{}).Error
}

func (s *Store) DeleteResetTokenByToken(token string) error {
	return s.db.Unscoped().Where("token = ?", token).Delete(&ResetToken{}).Error
}

func (s *Store) UpdatePassword(userID int64, hash string) error {
	return s.db.Model(&User{}).Where("id = ?", userID).Update("password_hash", hash).Error
}

func (s *Store) CreatePasswordResetRequestLog(log *PasswordResetRequestLog) error {
	return s.db.Create(log).Error
}

func (s *Store) CountPasswordResetRequestsSince(userID int64, since time.Time) (int64, error) {
	var count int64
	if err := s.db.Model(&PasswordResetRequestLog{}).
		Where("user_id = ? AND created_at >= ?", userID, since).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) CreatePasswordSecurityAuditLog(log *PasswordSecurityAuditLog) error {
	return s.db.Create(log).Error
}

func (s *Store) WriteAuditLog(log *AuditLog) error {
	return s.db.Create(log).Error
}

func (s *Store) QueryAuditLogs(accountID, gatewayID, action, result string, page, pageSize int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64

	query := s.db.Model(&AuditLog{}).Order("created_at DESC")

	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if gatewayID != "" {
		query = query.Where("gateway_id = ?", gatewayID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if result != "" {
		query = query.Where("result = ?", result)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (s *Store) UpsertAccount(account *Account) error {
	if account.CreatedAt.IsZero() {
		account.CreatedAt = time.Now()
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "region_id", "access_key_id", "access_key_secret", "accept_language", "updated_at"}),
	}).Create(account).Error
}

func (s *Store) ListAccounts() ([]Account, error) {
	var accounts []Account
	if err := s.db.Where("deleted_at IS NULL").Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

func (s *Store) GetAccountByID(id string) (*Account, error) {
	var account Account
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (s *Store) UpsertGateway(gateway *Gateway) error {
	if gateway.CreatedAt.IsZero() {
		gateway.CreatedAt = time.Now()
	}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"account_id", "name", "batch_tls", "updated_at"}),
	}).Create(gateway).Error
}

func (s *Store) ListGateways(accountID string) ([]Gateway, error) {
	var gateways []Gateway
	query := s.db.Where("deleted_at IS NULL")
	if accountID != "" {
		query = query.Where("account_id = ?", accountID)
	}
	if err := query.Find(&gateways).Error; err != nil {
		return nil, err
	}
	return gateways, nil
}

func (s *Store) GetGatewayByID(id string) (*Gateway, error) {
	var gateway Gateway
	if err := s.db.Where("id = ? AND deleted_at IS NULL", id).First(&gateway).Error; err != nil {
		return nil, err
	}
	return &gateway, nil
}

func (s *Store) CreateAccount(account *Account) error {
	return s.db.Create(account).Error
}

func (s *Store) DeleteAccount(id string) error {
	return s.db.Delete(&Account{}, "id = ?", id).Error
}

func (s *Store) CreateGateway(gateway *Gateway) error {
	return s.db.Create(gateway).Error
}

func (s *Store) DeleteGateway(id string) error {
	return s.db.Delete(&Gateway{}, "id = ?", id).Error
}
