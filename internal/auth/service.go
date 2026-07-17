package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go-service/internal/users"
	"go-service/pkg/middleware"

	"gorm.io/gorm"
)

// dbQuerier abstracts database operations needed by AuthService.
// This interface enables injecting a fake implementation in tests.
type dbQuerier interface {
	First(dest interface{}, conds ...interface{}) error
	Create(dest interface{}) error
	Save(dest interface{}) error
}

// gormQuerier adapts *gorm.DB to the dbQuerier interface.
type gormQuerier struct {
	db *gorm.DB
}

func (g *gormQuerier) First(dest interface{}, conds ...interface{}) error {
	return g.db.Where(conds[0]).First(dest).Error
}

func (g *gormQuerier) Create(dest interface{}) error {
	return g.db.Create(dest).Error
}

func (g *gormQuerier) Save(dest interface{}) error {
	return g.db.Save(dest).Error
}

// wechatSessionResponse is the JSON payload returned by the WeChat jscode2session API.
type wechatSessionResponse struct {
	Openid     string `json:"openid"`
	Unionid    string `json:"unionid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// LoginResult holds the result of a successful login.
type LoginResult struct {
	Token string         `json:"token"`
	User  LoginUserInfo  `json:"user"`
}

// LoginUserInfo is a subset of the user returned in the login response.
type LoginUserInfo struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Phone    string `json:"phone,omitempty"`
}

// AuthService handles WeChat Mini Program authentication.
type AuthService struct {
	db            dbQuerier
	jwtSecret     string
	appID         string
	appSecret     string
	httpClient    *http.Client
	wechatBaseURL string // overridable for tests
}

// NewAuthService creates an AuthService backed by a real *gorm.DB.
func NewAuthService(db *gorm.DB, jwtSecret, appID, appSecret string) *AuthService {
	return &AuthService{
		db:            &gormQuerier{db: db},
		jwtSecret:     jwtSecret,
		appID:         appID,
		appSecret:     appSecret,
		httpClient:    &http.Client{},
		wechatBaseURL: "https://api.weixin.qq.com",
	}
}

// newAuthServiceFromQuerier is used internally and in tests to inject a dbQuerier directly.
func newAuthServiceFromQuerier(q dbQuerier, jwtSecret, appID, appSecret string) *AuthService {
	return &AuthService{
		db:            q,
		jwtSecret:     jwtSecret,
		appID:         appID,
		appSecret:     appSecret,
		httpClient:    &http.Client{},
		wechatBaseURL: "https://api.weixin.qq.com",
	}
}

// WechatLogin exchanges a WeChat code for a JWT token.
// It calls the WeChat jscode2session API, looks up or creates the user by openid,
// updates user info if nickName/avatarUrl are provided, and returns a signed JWT
// along with the user info.
func (s *AuthService) WechatLogin(ctx context.Context, code, nickName, avatarURL string) (*LoginResult, error) {
	wechatData, err := s.fetchWechatSession(ctx, code)
	if err != nil {
		return nil, err
	}

	// Look up user by openid
	user := &users.User{}
	err = s.db.First(user, users.User{Openid: wechatData.Openid})

	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = &users.User{
			Openid:     wechatData.Openid,
			Unionid:    wechatData.Unionid,
			SessionKey: wechatData.SessionKey,
			Nickname:   nickName,
			Avatar:     avatarURL,
		}
		if err := s.db.Create(user); err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	} else {
		// Update existing user's session key and optional fields
		user.SessionKey = wechatData.SessionKey
		if wechatData.Unionid != "" {
			user.Unionid = wechatData.Unionid
		}
		if nickName != "" {
			user.Nickname = nickName
		}
		if avatarURL != "" {
			user.Avatar = avatarURL
		}
		if err := s.db.Save(user); err != nil {
			return nil, fmt.Errorf("update user: %w", err)
		}
	}

	token, err := middleware.GenerateToken(user.ID, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("token generation failed: %w", err)
	}

	return &LoginResult{
		Token: token,
		User: LoginUserInfo{
			ID:       user.ID,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Phone:    user.Phone,
		},
	}, nil
}

// fetchWechatSession calls the WeChat jscode2session API and returns the full session response.
func (s *AuthService) fetchWechatSession(ctx context.Context, code string) (*wechatSessionResponse, error) {
	apiURL := s.wechatBaseURL + "/sns/jscode2session"

	params := url.Values{}
	params.Set("appid", s.appID)
	params.Set("secret", s.appSecret)
	params.Set("js_code", code)
	params.Set("grant_type", "authorization_code")

	fullURL := apiURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build wechat request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wechat API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("wechat API returned HTTP %d", resp.StatusCode)
	}

	var result wechatSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode wechat response: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error %d: %s", result.ErrCode, result.ErrMsg)
	}

	if result.Openid == "" {
		return nil, fmt.Errorf("wechat returned empty openid")
	}

	return &result, nil
}
