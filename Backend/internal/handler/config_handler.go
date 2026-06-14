package handler

import (
	"net/http"

	"mentalchat/internal/config"
	"mentalchat/internal/storage"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	storage *storage.Storage
}

func NewConfigHandler(storage *storage.Storage) *ConfigHandler {
	return &ConfigHandler{storage: storage}
}

// GetSyncConfig возвращает настройки синхронизации (только если multi-device).
// GET /api/v1/config/sync
func (h *ConfigHandler) GetSyncConfig(c *gin.Context) {
	cfg := config.LoadConfig()
	userID := c.MustGet("user_id").(uint)

	deviceCount, _ := h.storage.GetUserDeviceCount(userID)
	multiDevice := deviceCount > 1

	c.JSON(http.StatusOK, gin.H{
		"sync_interval_seconds": cfg.Sync.IntervalSeconds,
		"multi_device":          multiDevice,
		"devices":               deviceCount,
	})
}
