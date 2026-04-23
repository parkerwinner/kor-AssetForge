package models

import (
	"time"

	"gorm.io/gorm"
)

// UserRole represents user roles for RBAC
type UserRole string

const (
	RoleUser      UserRole = "user"
	RoleAdmin     UserRole = "admin"
	RoleModerator UserRole = "moderator"
)

// User represents a platform user
type User struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	StellarAddress    string         `gorm:"uniqueIndex;not null" json:"stellar_address"`
	Email             string         `gorm:"uniqueIndex" json:"email"`
	Username          string         `gorm:"uniqueIndex" json:"username"`
	PasswordHash      string         `gorm:"not null" json:"-"` // Never expose in JSON
	Role              UserRole       `gorm:"type:varchar(20);default:'user'" json:"role"`
	EmailVerified     bool           `gorm:"default:false" json:"email_verified"`
	EmailToken        string         `gorm:"index" json:"-"` // Email verification token
	EmailTokenExpires time.Time      `json:"-"`             // Token expiration
	PasswordResetToken string        `gorm:"index" json:"-"` // Password reset token
	PasswordResetExpires time.Time   `json:"-"`             // Reset token expiration
	KYCVerified       bool           `gorm:"default:false" json:"kyc_verified"`
	AccreditedInvestor bool           `gorm:"default:false" json:"accredited_investor"`
	LastLoginAt       *time.Time     `json:"last_login_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserSession represents active user sessions
type UserSession struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"not null" json:"user_id"`
	User         User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SessionToken string    `gorm:"uniqueIndex;not null" json:"session_token"`
	IPAddress    string    `gorm:"not null" json:"ip_address"`
	UserAgent    string    `json:"user_agent"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// UserBalance represents a user's token balance
type UserBalance struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         uint      `gorm:"not null" json:"user_id"`
	User           User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	AssetID        uint      `gorm:"not null" json:"asset_id"`
	Asset          Asset     `gorm:"foreignKey:AssetID" json:"asset,omitempty"`
	Balance        int64     `gorm:"not null;default:0" json:"balance"`
	LockedBalance  int64     `gorm:"not null;default:0" json:"locked_balance"` // For active listings
	UpdatedAt      time.Time `json:"updated_at"`
}
