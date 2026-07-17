package auth

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"go-service/pkg/response"
)

// authServiceIface abstracts AuthService for handler injection and testing.
type authServiceIface interface {
	WechatLogin(ctx context.Context, code, nickName, avatarURL string) (*LoginResult, error)
}

// WechatLoginRequest is the JSON body for the WeChat login endpoint.
type WechatLoginRequest struct {
	Code      string `json:"code"`
	NickName  string `json:"nickName"`
	AvatarURL string `json:"avatarUrl"`
}

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	service authServiceIface
}

// NewAuthHandler creates a new AuthHandler with the given service.
func NewAuthHandler(service authServiceIface) *AuthHandler {
	return &AuthHandler{service: service}
}

// WechatLogin handles POST /auth/wechat/login
// @Summary      WeChat mini-program login
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  WechatLoginRequest  true  "WeChat code and optional user info"
// @Success      200   {object}  map[string]interface{}
// @Failure      400   {object}  map[string]interface{}
// @Failure      500   {object}  map[string]interface{}
// @Router       /auth/wechat/login [post]
func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if req.Code == "" {
		response.BadRequest(c, "code is required")
		return
	}

	result, err := h.service.WechatLogin(c.Request.Context(), req.Code, req.NickName, req.AvatarURL)
	if err != nil {
		response.InternalError(c, fmt.Sprintf("login failed: %s", err.Error()))
		return
	}

	response.Success(c, result)
}
