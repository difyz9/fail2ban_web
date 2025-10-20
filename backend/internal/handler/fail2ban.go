package handler

import (
	"net/http"
	"strconv"

	"fail2ban-web/internal/model"
	"fail2ban-web/internal/service"

	"github.com/gin-gonic/gin"
)

type Fail2BanHandler struct {
	fail2banService *service.Fail2BanService
}

func NewFail2BanHandler(fail2banService *service.Fail2BanService) *Fail2BanHandler {
	return &Fail2BanHandler{
		fail2banService: fail2banService,
	}
}

// GetStats 获取统计信息
func (h *Fail2BanHandler) GetStats(c *gin.Context) {
	stats, err := h.fail2banService.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "stats_fetch_failed",
			"message": "Failed to fetch statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetBannedIPs 获取被禁IP列表
func (h *Fail2BanHandler) GetBannedIPs(c *gin.Context) {
	// 获取查询参数
	limitStr := c.DefaultQuery("limit", "0")
	limit, _ := strconv.Atoi(limitStr)

	bannedIPs, err := h.fail2banService.GetBannedIPs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "banned_ips_fetch_failed",
			"message": "Failed to fetch banned IPs",
		})
		return
	}

	// 如果设置了limit，则限制返回数量
	if limit > 0 && len(bannedIPs) > limit {
		bannedIPs = bannedIPs[:limit]
	}

	response := model.BannedIPsResponse{
		IPs:   bannedIPs,
		Total: len(bannedIPs),
	}

	c.JSON(http.StatusOK, response)
}

// UnbanIP 解禁IP
func (h *Fail2BanHandler) UnbanIP(c *gin.Context) {
	var req model.UnbanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.fail2banService.UnbanIP(req.Jail, req.IP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "unban_failed",
			"message": "Failed to unban IP",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IP successfully unbanned",
		"ip":      req.IP,
		"jail":    req.Jail,
	})
}

// BanIP 手动禁止IP
func (h *Fail2BanHandler) BanIP(c *gin.Context) {
	var req model.UnbanRequest // 使用相同的结构
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.fail2banService.BanIP(req.Jail, req.IP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "ban_failed",
			"message": "Failed to ban IP",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IP successfully banned",
		"ip":      req.IP,
		"jail":    req.Jail,
	})
}

// GetSystemInfo 获取系统信息
func (h *Fail2BanHandler) GetSystemInfo(c *gin.Context) {
	info, err := h.fail2banService.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "system_info_fetch_failed",
			"message": "Failed to fetch system information",
		})
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetJails 获取jail列表
func (h *Fail2BanHandler) GetJails(c *gin.Context) {
	jails, err := h.fail2banService.GetJails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jails_fetch_failed",
			"message": "Failed to fetch jails",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jails": jails,
		"total": len(jails),
	})
}

// GetJailStatus 获取指定jail的状态
func (h *Fail2BanHandler) GetJailStatus(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_jail_name",
			"message": "Jail name is required",
		})
		return
	}

	status, err := h.fail2banService.GetJailStatus(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jail_status_fetch_failed",
			"message": "Failed to fetch jail status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jail":   name,
		"status": status,
	})
}

// GetLogs 获取日志
func (h *Fail2BanHandler) GetLogs(c *gin.Context) {
	// 获取查询参数
	filePath := c.DefaultQuery("file", "/var/log/fail2ban.log")
	linesStr := c.DefaultQuery("lines", "100")
	search := c.Query("search")

	lines, _ := strconv.Atoi(linesStr)
	if lines <= 0 {
		lines = 100
	}
	if lines > 1000 {
		lines = 1000 // 限制最大行数
	}

	var logLines []string
	var err error

	if search != "" {
		logLines, err = h.fail2banService.SearchLogs(filePath, search, lines)
	} else {
		logLines, err = h.fail2banService.ParseLogFile(filePath, lines)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "logs_fetch_failed",
			"message": "Failed to fetch logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logLines,
		"total": len(logLines),
		"file":  filePath,
	})
}

// GetVersion 获取Fail2Ban版本
func (h *Fail2BanHandler) GetVersion(c *gin.Context) {
	version, err := h.fail2banService.GetVersion()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "version_fetch_failed",
			"message": "Failed to fetch version",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"version": version,
	})
}

// HealthCheck 健康检查
func (h *Fail2BanHandler) HealthCheck(c *gin.Context) {
	// 尝试获取状态来检查fail2ban是否运行
	_, err := h.fail2banService.GetStatus()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "error",
			"message": "Fail2Ban service is not available",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Fail2Ban Web Panel is running",
	})
}