package wechat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestClient 构造指向 httptest server 的 Client（baseURL 可覆盖）。
func newTestClient(baseURL string) *Client {
	c := NewClient("test-appid", "test-secret")
	c.baseURL = baseURL
	return c
}

func TestGetAccessToken_FetchesAndCaches(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cgi-bin/token") {
			calls++
			w.Write([]byte(`{"access_token":"TOKEN_ABC","expires_in":7200}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)

	tok, err := c.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "TOKEN_ABC" {
		t.Errorf("expected TOKEN_ABC, got %q", tok)
	}
	if calls != 1 {
		t.Errorf("expected 1 fetch, got %d", calls)
	}

	// 第二次调用应命中缓存，不再请求
	tok2, _ := c.GetAccessToken(context.Background())
	if tok2 != "TOKEN_ABC" {
		t.Errorf("cached token mismatch: %q", tok2)
	}
	if calls != 1 {
		t.Errorf("expected cache hit (still 1 fetch), got %d", calls)
	}
}

func TestGetAccessToken_RefetchesAfterExpiry(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Write([]byte(`{"access_token":"TOKEN_X","expires_in":7200}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	if _, err := c.GetAccessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	// 手动把过期时间提前，模拟缓存失效
	c.expireAt = time.Now().Add(-time.Second)
	if _, err := c.GetAccessToken(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Errorf("expected refetch after expiry (2 calls), got %d", calls)
	}
}

func TestGetAccessToken_ErrcodeFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"errcode":40013,"errmsg":"invalid appid"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	if _, err := c.GetAccessToken(context.Background()); err == nil {
		t.Error("expected error on errcode, got nil")
	}
}

func TestSendSubscribeMessage_Success(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cgi-bin/token") {
			w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
			return
		}
		if strings.Contains(r.URL.Path, "/cgi-bin/message/subscribe/send") {
			buf := make([]byte, r.ContentLength)
			r.Body.Read(buf)
			gotBody = string(buf)
			w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	data := map[string]any{"thing1": map[string]string{"value": "小明"}}
	err := c.SendSubscribeMessage(context.Background(), "openid-123", "tpl-abc", data, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotBody, "openid-123") || !strings.Contains(gotBody, "tpl-abc") {
		t.Errorf("request body missing touser/template_id: %s", gotBody)
	}
	if strings.Contains(gotBody, "\"page\"") {
		t.Errorf("empty page should not appear in body: %s", gotBody)
	}
}

func TestSendSubscribeMessage_WithPage(t *testing.T) {
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cgi-bin/token") {
			w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
			return
		}
		if strings.Contains(r.URL.Path, "/cgi-bin/message/subscribe/send") {
			buf := make([]byte, r.ContentLength)
			r.Body.Read(buf)
			gotBody = string(buf)
			w.Write([]byte(`{"errcode":0,"errmsg":"ok"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	data := map[string]any{"thing1": map[string]string{"value": "小明"}}
	err := c.SendSubscribeMessage(context.Background(), "openid-123", "tpl-abc", data, "pages/notifications/index")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotBody, "\"page\"") || !strings.Contains(gotBody, "pages/notifications/index") {
		t.Errorf("request body missing page field: %s", gotBody)
	}
}

func TestSendSubscribeMessage_ErrcodeFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cgi-bin/token") {
			w.Write([]byte(`{"access_token":"TOK","expires_in":7200}`))
			return
		}
		w.Write([]byte(`{"errcode":43101,"errmsg":"user refuse to accept the msg"}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	err := c.SendSubscribeMessage(context.Background(), "openid-123", "tpl-abc", map[string]any{}, "")
	if err == nil {
		t.Error("expected error on errcode 43101, got nil")
	}
}
