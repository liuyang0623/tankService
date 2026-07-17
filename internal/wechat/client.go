// Package wechat 封装微信小程序服务端能力：access_token 获取（带缓存）与订阅消息发送。
package wechat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Client 持有微信调用凭据与 access_token 缓存。并发安全。
type Client struct {
	appID     string
	appSecret string
	http      *http.Client
	baseURL   string // 可覆盖，测试用

	mu       sync.Mutex
	token    string
	expireAt time.Time
}

// NewClient 创建微信客户端。
func NewClient(appID, appSecret string) *Client {
	return &Client{
		appID:     appID,
		appSecret: appSecret,
		http:      &http.Client{Timeout: 10 * time.Second},
		baseURL:   "https://api.weixin.qq.com",
	}
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// GetAccessToken 返回有效的 access_token；缓存未过期则复用，否则重新获取。
// 缓存有效期取 expires_in - 300s，留 5 分钟余量。加锁防并发重复取。
func (c *Client) GetAccessToken(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.expireAt) {
		return c.token, nil
	}

	params := url.Values{}
	params.Set("grant_type", "client_credential")
	params.Set("appid", c.appID)
	params.Set("secret", c.appSecret)
	fullURL := c.baseURL + "/cgi-bin/token?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	if tr.ErrCode != 0 {
		return "", fmt.Errorf("wechat token error %d: %s", tr.ErrCode, tr.ErrMsg)
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("wechat returned empty access_token")
	}

	c.token = tr.AccessToken
	c.expireAt = time.Now().Add(time.Duration(tr.ExpiresIn-300) * time.Second)
	return c.token, nil
}

type sendResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// SendSubscribeMessage 发送一条订阅消息给 openid 对应用户。
// data 按微信模板字段结构组装，如 {"thing1": {"value": "xxx"}}。
func (c *Client) SendSubscribeMessage(ctx context.Context, openid, tplID string, data map[string]any) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	body := map[string]any{
		"touser":      openid,
		"template_id": tplID,
		"data":        data,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal subscribe body: %w", err)
	}

	fullURL := c.baseURL + "/cgi-bin/message/subscribe/send?access_token=" + url.QueryEscape(token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build subscribe request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("subscribe request: %w", err)
	}
	defer resp.Body.Close()

	var sr sendResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return fmt.Errorf("decode subscribe response: %w", err)
	}
	if sr.ErrCode != 0 {
		return fmt.Errorf("wechat subscribe error %d: %s", sr.ErrCode, sr.ErrMsg)
	}
	return nil
}
