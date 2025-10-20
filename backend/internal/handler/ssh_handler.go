package handler

import (
	"net/http"
	"strconv"

	"fail2ban-web/internal/service"

	"github.com/gin-gonic/gin"
)

type SSHHandler struct {
	sshService *service.SSHService
	defaultSSHService *service.DefaultSSHService
}

func NewSSHHandler(sshService *service.SSHService, defaultSSHService *service.DefaultSSHService) *SSHHandler {
	return &SSHHandler{
		sshService: sshService,
		defaultSSHService: defaultSSHService,
	}
}

// GetSSHStats 获取SSH统计信息
func (h *SSHHandler) GetSSHStats(c *gin.Context) {
	stats, err := h.sshService.GetSSHStats()
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

// GetSSHLogs 获取SSH日志
func (h *SSHHandler) GetSSHLogs(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	logs, err := h.sshService.GetSSHLogs(limit)
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

// GetSSHJailStatus 获取SSH jail状态
func (h *SSHHandler) GetSSHJailStatus(c *gin.Context) {
	status, err := h.sshService.GetSSHJailStatus()
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

// BanSSHIP 手动禁止SSH IP
func (h *SSHHandler) BanSSHIP(c *gin.Context) {
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

	if err := h.sshService.BanSSHIP(req.IP, req.Jail); err != nil {
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

// UnbanSSHIP 解禁SSH IP
func (h *SSHHandler) UnbanSSHIP(c *gin.Context) {
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

	if err := h.sshService.UnbanSSHIP(req.IP, req.Jail); err != nil {
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

// GetSSHDefaults 获取SSH默认配置
func (h *SSHHandler) GetSSHDefaults(c *gin.Context) {
	jails := h.defaultSSHService.GetDefaultSSHJails()
	filters := h.defaultSSHService.GetSSHFilterTemplates()
	config := h.defaultSSHService.GetSSHJailConfig()
	practices := h.defaultSSHService.GetSSHBestPractices()
	tips := h.defaultSSHService.GetSSHSecurityTips()

	c.JSON(http.StatusOK, gin.H{
		"jails":    jails,
		"filters":  filters,
		"config":   config,
		"practices": practices,
		"tips":     tips,
	})
}

// InstallSSHDefaults 安装SSH默认配置
func (h *SSHHandler) InstallSSHDefaults(c *gin.Context) {
	if err := h.defaultSSHService.InstallSSHDefaults(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_install_defaults",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "SSH default configurations installed successfully",
	})
}