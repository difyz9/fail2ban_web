package handler

import (
	"net/http"
	"strconv"

	"fail2ban-web/internal/service"

	"github.com/gin-gonic/gin"
)

type NginxHandler struct {
	nginxService *service.NginxService
	defaultNginxService *service.DefaultNginxService
	defaultNginxAdvancedService *service.DefaultNginxAdvancedService
}

func NewNginxHandler(nginxService *service.NginxService, defaultNginxService *service.DefaultNginxService, defaultNginxAdvancedService *service.DefaultNginxAdvancedService) *NginxHandler {
	return &NginxHandler{
		nginxService: nginxService,
		defaultNginxService: defaultNginxService,
		defaultNginxAdvancedService: defaultNginxAdvancedService,
	}
}

// GetNginxStats 获取Nginx统计信息
func (h *NginxHandler) GetNginxStats(c *gin.Context) {
	stats, err := h.nginxService.GetNginxStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_get_stats",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// GetNginxLogs 获取Nginx日志
func (h *NginxHandler) GetNginxLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	logs, err := h.nginxService.GetNginxLogs(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_get_logs",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs": logs,
		"total": len(logs),
	})
}

// GetNginxJailStatus 获取Nginx jail状态
func (h *NginxHandler) GetNginxJailStatus(c *gin.Context) {
	status, err := h.nginxService.GetNginxJailStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_get_status",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}

// BanNginxIP 手动禁止Nginx IP
func (h *NginxHandler) BanNginxIP(c *gin.Context) {
	var req struct {
		IP   string `json:"ip" binding:"required"`
		Jail string `json:"jail"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.nginxService.BanNginxIP(req.IP, req.Jail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_ban_ip",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IP banned successfully",
		"ip":      req.IP,
		"jail":    req.Jail,
	})
}

// UnbanNginxIP 解禁Nginx IP
func (h *NginxHandler) UnbanNginxIP(c *gin.Context) {
	var req struct {
		IP   string `json:"ip" binding:"required"`
		Jail string `json:"jail"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.nginxService.UnbanNginxIP(req.IP, req.Jail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_unban_ip",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IP unbanned successfully",
		"ip":      req.IP,
		"jail":    req.Jail,
	})
}

// GetNginxDefaults 获取Nginx默认配置
func (h *NginxHandler) GetNginxDefaults(c *gin.Context) {
	jails := h.defaultNginxService.GetDefaultNginxJails()
	filters := h.defaultNginxService.GetNginxFilterTemplates()

	c.JSON(http.StatusOK, gin.H{
		"jails":   jails,
		"filters": filters,
	})
}

// GetNginxAdvancedDefaults 获取Nginx高级默认配置
func (h *NginxHandler) GetNginxAdvancedDefaults(c *gin.Context) {
	jails := h.defaultNginxAdvancedService.GetAdvancedNginxJails()
	filters := h.defaultNginxAdvancedService.GetAdvancedNginxFilterTemplates()
	config := h.defaultNginxAdvancedService.GetNginxSecurityConfig()
	practices := h.defaultNginxAdvancedService.GetNginxBestPractices()

	c.JSON(http.StatusOK, gin.H{
		"jails":     jails,
		"filters":   filters,
		"config":    config,
		"practices": practices,
	})
}

// InstallNginxDefaults 安装Nginx默认配置
func (h *NginxHandler) InstallNginxDefaults(c *gin.Context) {
	if err := h.defaultNginxService.InstallNginxDefaults(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_install_defaults",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Nginx default configurations installed successfully",
	})
}

// InstallNginxAdvancedDefaults 安装Nginx高级默认配置
func (h *NginxHandler) InstallNginxAdvancedDefaults(c *gin.Context) {
	if err := h.defaultNginxAdvancedService.InstallAdvancedNginxDefaults(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_install_advanced_defaults",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Nginx advanced configurations installed successfully",
	})
}