package handler

import (
	"net/http"
	"os"

	"fail2ban-web/core"
	"fail2ban-web/internal/middleware"
	"fail2ban-web/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	BaseHandler
}

func NewAuthHandler(app *core.AppServer, db *gorm.DB) *AuthHandler {
	return &AuthHandler{
		BaseHandler: BaseHandler{
			App: app,
			DB:  db,
		},
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			"invalid_request",
			err.Error(),
		))
		return
	}

	// 验证用户名和密码（临时使用环境变量或默认值）
	adminUser := getEnvOrDefault("ADMIN_USERNAME", "admin")
	adminPass := getEnvOrDefault("ADMIN_PASSWORD", "admin")
	
	if req.Username != adminUser || req.Password != adminPass {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			"invalid_credentials",
			"Invalid username or password",
		))
		return
	}

	// 生成JWT token
	token, expiresAt, err := middleware.GenerateToken(1, req.Username, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			"token_generation_failed",
			"Failed to generate token",
		))
		return
	}

	// 创建用户对象（用于响应）
	adminEmail := getEnvOrDefault("ADMIN_EMAIL", "admin@example.com")
	user := model.User{
		ID:       1,
		Username: adminUser,
		Email:    adminEmail,
		Role:     "admin",
		IsActive: true,
	}

	response := model.AuthResponse{
		Token:     token,
		User:      user,
		ExpiresAt: expiresAt,
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(response, "Login successful"))
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			"unauthorized",
			"User not authenticated",
		))
		return
	}

	// 返回管理员用户信息
	adminEmail := getEnvOrDefault("ADMIN_EMAIL", "admin@example.com")
	user := model.User{
		ID:       1,
		Username: username.(string),
		Email:    adminEmail,
		Role:     "admin",
		IsActive: true,
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user, "Profile retrieved successfully"))
}

// RefreshToken 刷新token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			"unauthorized",
			"User not authenticated",
		))
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			"unauthorized",
			"User information not found",
		))
		return
	}

	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
			"unauthorized",
			"User role not found",
		))
		return
	}

	// 生成新的JWT token
	token, expiresAt, err := middleware.GenerateToken(userID.(uint), username.(string), role.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			"token_generation_failed",
			"Failed to generate token",
		))
		return
	}

	data := gin.H{
		"token":      token,
		"expires_at": expiresAt,
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(data, "Token refreshed successfully"))
}

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
