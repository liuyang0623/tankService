package appconfig

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-service/internal/users"
	"go-service/pkg/response"
)

// openidLookup 抽象「按 userID 反查 openid」,便于中间件注入 fake 进行测试。
type openidLookup interface {
	OpenidByUserID(userID uint) (string, error)
}

// gormOpenidLookup 用 *gorm.DB 实现 openidLookup,查询 users 表的 openid 列。
type gormOpenidLookup struct {
	db *gorm.DB
}

func (g *gormOpenidLookup) OpenidByUserID(userID uint) (string, error) {
	var u users.User
	if err := g.db.Select("openid").First(&u, userID).Error; err != nil {
		return "", err
	}
	return u.Openid, nil
}

// newAdminOnly 用给定的 openid 反查器与白名单构造管理员鉴权中间件。
// 中间件依赖上游 JWTMiddleware 已注入 "userID";未注入或不在白名单一律 403。
func newAdminOnly(lookup openidLookup, adminOpenids []string) gin.HandlerFunc {
	whitelist := make(map[string]struct{}, len(adminOpenids))
	for _, id := range adminOpenids {
		whitelist[id] = struct{}{}
	}
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		if userID == 0 {
			response.Error(c, http.StatusForbidden, "admin only")
			c.Abort()
			return
		}
		openid, err := lookup.OpenidByUserID(userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, http.StatusForbidden, "admin only")
			} else {
				response.InternalError(c, "server error")
			}
			c.Abort()
			return
		}
		if _, ok := whitelist[openid]; !ok {
			response.Error(c, http.StatusForbidden, "admin only")
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminOnlyMiddleware 是生产装配入口:用 *gorm.DB 反查 openid,按白名单校验管理员。
func AdminOnlyMiddleware(db *gorm.DB, adminOpenids []string) gin.HandlerFunc {
	return newAdminOnly(&gormOpenidLookup{db: db}, adminOpenids)
}
