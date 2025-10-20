package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type DefaultConfigHandler struct {
}

func NewDefaultConfigHandler() *DefaultConfigHandler {
	return &DefaultConfigHandler{}
}

// GetDefaultConfigInfo 获取默认配置信息
func (h *DefaultConfigHandler) GetDefaultConfigInfo(c *gin.Context) {
	info := map[string]interface{}{
		"title":       "Fail2Ban Default Configurations",
		"description": "Pre-configured security rules for common services",
		"version":     "1.0.0",
		"categories": map[string]interface{}{
			"ssh": map[string]interface{}{
				"name":        "SSH Protection",
				"jails":       []string{"sshd", "sshd-ddos", "sshd-aggressive"},
				"description": "Protects against SSH brute force and DDoS attacks",
			},
			"nginx": map[string]interface{}{
				"name":        "Nginx Web Protection",
				"jails":       []string{"nginx-http-auth", "nginx-botsearch", "nginx-bad-request", "nginx-limit-req"},
				"description": "Comprehensive web application security for Nginx",
			},
		},
		"installation_steps": []string{
			"1. Use /api/ssh/defaults/install for SSH protection",
			"2. Use /api/nginx/defaults/install for basic Nginx protection", 
			"3. Use /api/nginx/defaults/advanced/install for advanced Nginx protection",
			"4. Adjust configurations as needed for your environment",
		},
		"log_requirements": map[string]interface{}{
			"ssh": map[string]string{
				"log_path": "/var/log/auth.log",
				"format":   "Standard syslog format",
			},
			"nginx": map[string]string{
				"access_log": "/var/log/nginx/access.log",
				"error_log":  "/var/log/nginx/error.log",
				"format":     "Default Nginx log format",
			},
		},
		"recommendations": []string{
			"Start with conservative settings and adjust as needed",
			"Monitor fail2ban.log for any issues",
			"Consider implementing rate limiting in Nginx itself",
			"Use whitelist for trusted IPs",
			"Test configurations in non-production environment first",
		},
	}

	c.JSON(http.StatusOK, info)
}

// InstallNginxDefaults 安装默认的 Nginx 配置
func (h *DefaultConfigHandler) InstallNginxDefaults(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 /api/nginx/defaults/install 接口安装Nginx默认配置",
		"redirect": "/api/nginx/defaults/install",
	})
}

// GetNginxFilterTemplates 获取 Nginx 过滤器模板
func (h *DefaultConfigHandler) GetNginxFilterTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 /api/nginx/defaults 接口获取Nginx过滤器模板",
		"redirect": "/api/nginx/defaults",
	})
}

// GetNginxJailConfig 获取 Nginx jail 配置
func (h *DefaultConfigHandler) GetNginxJailConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 /api/nginx/defaults 接口获取Nginx配置",
		"redirect": "/api/nginx/defaults",
	})
}

// ExportNginxConfig 导出 Nginx 配置
func (h *DefaultConfigHandler) ExportNginxConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "请使用 /api/nginx/defaults/advanced 接口获取完整配置",
		"redirect": "/api/nginx/defaults/advanced",
	})
}