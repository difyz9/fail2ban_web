package model

import (
	"time"

	"gorm.io/gorm"
)

// ApiResponse 统一API响应格式
type ApiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}, message ...string) ApiResponse {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return ApiResponse{
		Success: true,
		Data:    data,
		Message: msg,
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(error string, message ...string) ApiResponse {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	return ApiResponse{
		Success: false,
		Error:   error,
		Message: msg,
	}
}

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

// AccessLog 访问日志模型（统一存储 SSH 和 Nginx 访问日志）
type AccessLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ServiceType string    `json:"service_type" gorm:"index;not null"` // ssh, nginx, apache 等
	IPAddress   string    `json:"ip_address" gorm:"index;not null"`
	Username    string    `json:"username" gorm:"index"`              // SSH 用户名或 HTTP 认证用户
	Event       string    `json:"event"`                               // failed_password, accepted_password, http_request 等
	Status      string    `json:"status" gorm:"index"`                // success, failed, blocked, info
	Method      string    `json:"method"`                              // HTTP 方法 (GET, POST等) 或 SSH 事件类型
	Path        string    `json:"path"`                                // HTTP 请求路径或 SSH 连接信息
	StatusCode  int       `json:"status_code"`                         // HTTP 状态码
	UserAgent   string    `json:"user_agent"`                          // HTTP User-Agent
	Referer     string    `json:"referer"`                             // HTTP Referer
	BytesSent   int64     `json:"bytes_sent"`                          // 发送字节数
	RequestTime float64   `json:"request_time"`                        // 请求耗时（秒）
	Country     string    `json:"country" gorm:"index"`                // IP 所属国家
	City        string    `json:"city"`                                // IP 所属城市
	Latitude    float64   `json:"latitude"`                            // 纬度
	Longitude   float64   `json:"longitude"`                           // 经度
	IsThreat    bool      `json:"is_threat" gorm:"index;default:false"` // 是否威胁
	ThreatLevel string    `json:"threat_level"`                        // 威胁等级: low, medium, high, critical
	RawLog      string    `json:"raw_log" gorm:"type:text"`            // 原始日志行
	LogTime     time.Time `json:"log_time" gorm:"index;not null"`     // 日志时间
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
}

// IPStatistics IP 统计信息模型
type IPStatistics struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	IPAddress       string    `json:"ip_address" gorm:"uniqueIndex;not null"`
	TotalRequests   int       `json:"total_requests" gorm:"default:0"`      // 总请求数
	FailedRequests  int       `json:"failed_requests" gorm:"default:0"`     // 失败请求数
	SuccessRequests int       `json:"success_requests" gorm:"default:0"`    // 成功请求数
	SSHAttempts     int       `json:"ssh_attempts" gorm:"default:0"`        // SSH 尝试次数
	HTTPRequests    int       `json:"http_requests" gorm:"default:0"`       // HTTP 请求次数
	LastSeen        time.Time `json:"last_seen" gorm:"index"`               // 最后出现时间
	FirstSeen       time.Time `json:"first_seen" gorm:"index"`              // 首次出现时间
	Country         string    `json:"country" gorm:"index"`
	City            string    `json:"city"`
	IsBanned        bool      `json:"is_banned" gorm:"index;default:false"` // 是否被禁
	BanCount        int       `json:"ban_count" gorm:"default:0"`           // 被禁次数
	ThreatScore     int       `json:"threat_score" gorm:"default:0"`        // 威胁评分 0-100
	IsWhitelisted   bool      `json:"is_whitelisted" gorm:"default:false"`  // 是否白名单
	Notes           string    `json:"notes" gorm:"type:text"`               // 备注
	CreatedAt       time.Time `json:"created_at" gorm:"index"`
	UpdatedAt       time.Time `json:"updated_at"`
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