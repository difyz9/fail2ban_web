package handler

import (
	"log"
	"net/http"

	"fail2ban-web/internal/service"

	"github.com/gin-gonic/gin"
)

type IntelligentHandler struct {
	intelligentService *service.IntelligentScanService
}

func NewIntelligentHandler(intelligentService *service.IntelligentScanService) *IntelligentHandler {
	return &IntelligentHandler{
		intelligentService: intelligentService,
	}
}

// GetCurrentThreats 获取当前威胁
func (h *IntelligentHandler) GetCurrentThreats(c *gin.Context) {
	threats := h.intelligentService.GetCurrentThreats()
	
	c.JSON(http.StatusOK, gin.H{
		"threats": threats,
		"total":   len(threats),
	})
}

// GetScanResult 获取扫描结果
func (h *IntelligentHandler) GetScanResult(c *gin.Context) {
	result := h.intelligentService.GetScanResult()
	
	c.JSON(http.StatusOK, gin.H{
		"result": result,
	})
}

// ManualBanIP 手动封禁IP
func (h *IntelligentHandler) ManualBanIP(c *gin.Context) {
	var req struct {
		IP     string `json:"ip" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if req.Reason == "" {
		req.Reason = "手动封禁"
	}

	if err := h.intelligentService.ManualBanIP(req.IP, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed_to_ban_ip",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "IP已被手动封禁",
		"ip":      req.IP,
		"reason":  req.Reason,
	})
}

// GetThreatStats 获取威胁统计
func (h *IntelligentHandler) GetThreatStats(c *gin.Context) {
	threats := h.intelligentService.GetCurrentThreats()
	
	stats := struct {
		TotalThreats   int `json:"total_threats"`
		HighRisk       int `json:"high_risk"`
		MediumRisk     int `json:"medium_risk"`
		LowRisk        int `json:"low_risk"`
		AutoBanned     int `json:"auto_banned"`
		SSHThreats     int `json:"ssh_threats"`
		NginxThreats   int `json:"nginx_threats"`
	}{}
	
	for _, threat := range threats {
		stats.TotalThreats++
		
		if threat.AutoBanned {
			stats.AutoBanned++
		}
		
		if threat.SSHAttempts > 0 {
			stats.SSHThreats++
		}
		
		if threat.NginxAttempts > 0 {
			stats.NginxThreats++
		}
		
		if threat.ThreatScore >= 80 {
			stats.HighRisk++
		} else if threat.ThreatScore >= 50 {
			stats.MediumRisk++
		} else {
			stats.LowRisk++
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// AnalyzeLogFile 分析日志文件
func (h *IntelligentHandler) AnalyzeLogFile(c *gin.Context) {
	var req struct {
		LogFilePath string `json:"log_file_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "log_file_path is required",
		})
		return
	}

	// 异步分析日志文件
	go func() {
		if err := h.intelligentService.AnalyzeLogFile(req.LogFilePath); err != nil {
			log.Printf("日志分析失败: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message":       "日志分析已开始",
		"log_file_path": req.LogFilePath,
		"status":        "processing",
	})
}

// AnalyzeAccessLog 分析access.log文件
func (h *IntelligentHandler) AnalyzeAccessLog(c *gin.Context) {
	// 异步分析access.log文件
	go func() {
		if err := h.intelligentService.AnalyzeAccessLog(); err != nil {
			log.Printf("access.log分析失败: %v", err)
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "access.log自动分析已开始",
		"status":  "processing",
	})
}