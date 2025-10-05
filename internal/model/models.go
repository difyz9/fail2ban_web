package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // 不在JSON中返回密码
	Email     string         `json:"email" gorm:"uniqueIndex"`
	Role      string         `json:"role" gorm:"default:user"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BannedIP 被禁IP模型
type BannedIP struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	IPAddress   string    `json:"ip_address" gorm:"not null"`
	Jail        string    `json:"jail" gorm:"not null"`
	BanTime     time.Time `json:"ban_time"`
	UnbanTime   time.Time `json:"unban_time"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Fail2banJail jail 配置模型
type Fail2banJail struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"uniqueIndex;not null"`
	Enabled     bool   `json:"enabled" gorm:"default:true"`
	Port        string `json:"port"`
	Protocol    string `json:"protocol" gorm:"default:tcp"`
	Filter      string `json:"filter"`
	LogPath     string `json:"log_path"`
	MaxRetry    int    `json:"max_retry" gorm:"default:5"`
	FindTime    int    `json:"find_time" gorm:"default:600"`
	BanTime     int    `json:"ban_time" gorm:"default:3600"`
	Action      string `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user"`
	ExpiresAt int64  `json:"expires_at"`
}

// StatsResponse 统计响应
type StatsResponse struct {
	BannedCount   int    `json:"banned_count"`
	TodayBlocks   int    `json:"today_blocks"`
	ActiveRules   int    `json:"active_rules"`
	SystemStatus  string `json:"system_status"`
}

// SystemInfoResponse 系统信息响应
type SystemInfoResponse struct {
	Version       string `json:"version"`
	Uptime        int64  `json:"uptime"`
	BannedIPs     int    `json:"banned_ips"`
	ActiveJails   int    `json:"active_jails"`
}

// BannedIPResponse 被禁IP响应
type BannedIPResponse struct {
	Address       string `json:"address"`
	Jail          string `json:"jail"`
	BanTime       time.Time `json:"ban_time"`
	RemainingTime int64  `json:"remaining_time"`
}

// BannedIPsResponse 被禁IP列表响应
type BannedIPsResponse struct {
	IPs   []BannedIPResponse `json:"ips"`
	Total int                `json:"total"`
}

// UnbanRequest 解禁请求
type UnbanRequest struct {
	IP   string `json:"ip" binding:"required"`
	Jail string `json:"jail" binding:"required"`
}