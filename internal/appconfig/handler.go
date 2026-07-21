package appconfig

import (
	"context"

	"github.com/gin-gonic/gin"

	"go-service/pkg/response"
)

// appConfigServiceIface 抽象 AppConfigService 供 handler 注入与测试。
type appConfigServiceIface interface {
	GetConfig(ctx context.Context) (*AppConfigResponse, error)
	UpdateAuditMode(ctx context.Context, auditMode bool) error
}

// updateConfigRequest 是 PUT /app-config 的请求体。
// auditMode 用指针区分「未传」与「传 false」,未传时视为缺参返回 400。
type updateConfigRequest struct {
	AuditMode *bool `json:"auditMode"`
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

// UpdateConfig 更新全局应用配置（需管理员白名单鉴权，由中间件保护）。
//
// @Summary  更新应用配置
// @Tags     app-config
// @Accept   json
// @Produce  json
// @Param    body body updateConfigRequest true "配置项"
// @Success  200 {object} response.responseBody
// @Failure  400 {object} response.responseBody
// @Router   /app-config [put]
func (h *AppConfigHandler) UpdateConfig(c *gin.Context) {
	var req updateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if req.AuditMode == nil {
		response.BadRequest(c, "auditMode is required")
		return
	}
	if err := h.service.UpdateAuditMode(c.Request.Context(), *req.AuditMode); err != nil {
		response.InternalError(c, "server error")
		return
	}
	cfg, err := h.service.GetConfig(c.Request.Context())
	if err != nil {
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, cfg)
}
