package appconfig

import (
	"context"

	"github.com/gin-gonic/gin"

	"go-service/pkg/response"
)

// appConfigServiceIface 抽象 AppConfigService 供 handler 注入与测试。
type appConfigServiceIface interface {
	GetConfig(ctx context.Context) (*AppConfigResponse, error)
}

// AppConfigHandler 处理应用配置 HTTP 请求。
type AppConfigHandler struct {
	service appConfigServiceIface
}

// NewAppConfigHandler 构造。
func NewAppConfigHandler(service *AppConfigService) *AppConfigHandler {
	return &AppConfigHandler{service: service}
}

// GetConfig 返回全局应用配置（免鉴权，登录前/启动阶段调用）。
//
// @Summary  获取应用配置
// @Tags     app-config
// @Produce  json
// @Success  200 {object} response.responseBody
// @Router   /app-config [get]
func (h *AppConfigHandler) GetConfig(c *gin.Context) {
	cfg, err := h.service.GetConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, cfg)
}
