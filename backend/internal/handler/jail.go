package handler

import (
	"net/http"

	"fail2ban-web/internal/model"
	"fail2ban-web/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JailHandler struct {
	jailService *service.JailService
}

func NewJailHandler(jailService *service.JailService) *JailHandler {
	return &JailHandler{
		jailService: jailService,
	}
}

// CreateJail 创建新的jail配置
func (h *JailHandler) CreateJail(c *gin.Context) {
	var jail model.Fail2banJail
	if err := c.ShouldBindJSON(&jail); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	if err := h.jailService.CreateJail(&jail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jail_creation_failed",
			"message": "Failed to create jail configuration",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Jail configuration created successfully",
		"jail":    jail,
	})
}

// GetJails 获取jail配置列表
func (h *JailHandler) GetJails(c *gin.Context) {
	jails, err := h.jailService.GetAllJails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jails_fetch_failed",
			"message": "Failed to fetch jail configurations",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"jails": jails,
		"total": len(jails),
	})
}

// GetJail 获取指定jail配置
func (h *JailHandler) GetJail(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_jail_name",
			"message": "Jail name is required",
		})
		return
	}

	jail, err := h.jailService.GetJailByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "jail_not_found",
				"message": "Jail configuration not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jail_fetch_failed",
			"message": "Failed to fetch jail configuration",
		})
		return
	}

	c.JSON(http.StatusOK, jail)
}

// UpdateJail 更新jail配置
func (h *JailHandler) UpdateJail(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_jail_name",
			"message": "Jail name is required",
		})
		return
	}

	var updateData model.Fail2banJail
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	jail, err := h.jailService.GetJailByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "jail_not_found",
				"message": "Jail configuration not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jail_fetch_failed",
			"message": "Failed to fetch jail configuration",
		})
		return
	}

	// 更新字段
	jail.Enabled = updateData.Enabled
	jail.Port = updateData.Port
	jail.Protocol = updateData.Protocol
	jail.Filter = updateData.Filter
	jail.LogPath = updateData.LogPath
	jail.MaxRetry = updateData.MaxRetry
	jail.FindTime = updateData.FindTime
	jail.BanTime = updateData.BanTime
	jail.Action = updateData.Action

	if err := h.jailService.UpdateJail(jail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jail_update_failed",
			"message": "Failed to update jail configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Jail configuration updated successfully",
		"jail":    jail,
	})
}

// DeleteJail 删除jail配置
func (h *JailHandler) DeleteJail(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_jail_name",
			"message": "Jail name is required",
		})
		return
	}

	jail, err := h.jailService.GetJailByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "jail_not_found",
				"message": "Jail configuration not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jail_fetch_failed",
			"message": "Failed to fetch jail configuration",
		})
		return
	}

	if err := h.jailService.DeleteJail(jail.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jail_delete_failed",
			"message": "Failed to delete jail configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Jail configuration deleted successfully",
		"name":    name,
	})
}

// ToggleJail 启用/禁用jail
func (h *JailHandler) ToggleJail(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing_jail_name",
			"message": "Jail name is required",
		})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": err.Error(),
		})
		return
	}

	jail, err := h.jailService.GetJailByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "jail_not_found",
				"message": "Jail configuration not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jail_fetch_failed",
			"message": "Failed to fetch jail configuration",
		})
		return
	}

	jail.Enabled = req.Enabled
	if err := h.jailService.UpdateJail(jail); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "jail_toggle_failed",
			"message": "Failed to toggle jail status",
		})
		return
	}

	status := "disabled"
	if req.Enabled {
		status = "enabled"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Jail " + status + " successfully",
		"name":    name,
		"enabled": req.Enabled,
	})
}