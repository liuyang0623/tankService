package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-service/internal/users"

	"gorm.io/gorm"
)

// --- fake DB that implements dbQuerier ---

type fakeDB struct {
	// controls what First returns
	firstUser  *users.User
	firstError error
	// controls Create/Save behavior
	createError error
	saveError   error
	// records calls
	lastCreated *users.User
	lastSaved   *users.User
}

func (f *fakeDB) First(dest interface{}, conds ...interface{}) error {
	if f.firstError != nil {
		return f.firstError
	}
	u := dest.(*users.User)
	if f.firstUser != nil {
		*u = *f.firstUser
	}
	return nil
}

func (f *fakeDB) Create(dest interface{}) error {
	u := dest.(*users.User)
	if f.createError == nil {
		u.ID = 42 // simulate auto-increment
	}
	f.lastCreated = u
	return f.createError
}

func (f *fakeDB) Save(dest interface{}) error {
	u := dest.(*users.User)
	f.lastSaved = u
	return f.saveError
}

// --- helpers ---

func newTestServer(respBody string, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		w.Write([]byte(respBody)) //nolint:errcheck
	}))
}

func newServiceWithURL(db dbQuerier, baseURL string) *AuthService {
	s := newAuthServiceFromQuerier(db, "test-secret", "myAppID", "myAppSecret")
	s.wechatBaseURL = baseURL
	return s
}

// --- tests ---

func TestWechatLogin_Success_ExistingUser(t *testing.T) {
	wechatResp := `{"openid":"wx-openid-123","session_key":"sk","unionid":"u1"}`
	srv := newTestServer(wechatResp, http.StatusOK)
	defer srv.Close()

	fdb := &fakeDB{
		firstUser: &users.User{
			Model:  gorm.Model{ID: 7},
			Openid: "wx-openid-123",
		},
	}

	svc := newServiceWithURL(fdb, srv.URL)
	result, err := svc.WechatLogin(context.Background(), "some-code", "Nick", "https://img.url")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected non-empty token")
	}
	// JWT must be 3 dot-separated parts
	parts := strings.Split(result.Token, ".")
	if len(parts) != 3 {
		t.Errorf("expected JWT with 3 parts, got %d parts", len(parts))
	}
	// User info should be returned
	if result.User.ID != 7 {
		t.Errorf("expected user ID 7, got %d", result.User.ID)
	}
	// The user should have been saved with updated fields
	if fdb.lastSaved == nil {
		t.Fatal("expected Save to be called")
	}
	if fdb.lastSaved.Nickname != "Nick" {
		t.Errorf("expected nickname 'Nick', got %q", fdb.lastSaved.Nickname)
	}
	if fdb.lastSaved.Avatar != "https://img.url" {
		t.Errorf("expected avatar 'https://img.url', got %q", fdb.lastSaved.Avatar)
	}
}

func TestWechatLogin_Success_NewUser(t *testing.T) {
	wechatResp := `{"openid":"wx-new-user","session_key":"sk"}`
	srv := newTestServer(wechatResp, http.StatusOK)
	defer srv.Close()

	fdb := &fakeDB{
		firstError: gorm.ErrRecordNotFound,
	}

	svc := newServiceWithURL(fdb, srv.URL)
	result, err := svc.WechatLogin(context.Background(), "new-user-code", "NewUser", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if fdb.lastCreated == nil {
		t.Fatal("expected Create to be called")
	}
	if fdb.lastCreated.Openid != "wx-new-user" {
		t.Errorf("expected openid 'wx-new-user', got %q", fdb.lastCreated.Openid)
	}
	if fdb.lastCreated.Nickname != "NewUser" {
		t.Errorf("expected nickname 'NewUser', got %q", fdb.lastCreated.Nickname)
	}
}

func TestWechatLogin_WechatAPIError(t *testing.T) {
	wechatResp := `{"errcode":40029,"errmsg":"invalid code"}`
	srv := newTestServer(wechatResp, http.StatusOK)
	defer srv.Close()

	fdb := &fakeDB{}
	svc := newServiceWithURL(fdb, srv.URL)
	_, err := svc.WechatLogin(context.Background(), "bad-code", "", "")
	if err == nil {
		t.Fatal("expected error for wechat errcode != 0")
	}
	if !strings.Contains(err.Error(), "40029") && !strings.Contains(err.Error(), "invalid code") {
		t.Errorf("error should mention wechat error, got: %v", err)
	}
}

func TestWechatLogin_WechatHTTPError(t *testing.T) {
	// Server returns 500
	srv := newTestServer(`internal error`, http.StatusInternalServerError)
	defer srv.Close()

	fdb := &fakeDB{}
	svc := newServiceWithURL(fdb, srv.URL)
	_, err := svc.WechatLogin(context.Background(), "code", "", "")
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestWechatLogin_EmptyOpenid(t *testing.T) {
	// WeChat returns 200 but no openid (and no errcode)
	wechatResp := `{"session_key":"sk"}`
	srv := newTestServer(wechatResp, http.StatusOK)
	defer srv.Close()

	fdb := &fakeDB{}
	svc := newServiceWithURL(fdb, srv.URL)
	_, err := svc.WechatLogin(context.Background(), "code", "", "")
	if err == nil {
		t.Fatal("expected error for empty openid")
	}
}

func TestWechatLogin_DBError(t *testing.T) {
	wechatResp := `{"openid":"wx-openid-999","session_key":"sk"}`
	srv := newTestServer(wechatResp, http.StatusOK)
	defer srv.Close()

	fdb := &fakeDB{
		firstError: gorm.ErrInvalidDB,
	}
	svc := newServiceWithURL(fdb, srv.URL)
	_, err := svc.WechatLogin(context.Background(), "code", "", "")
	if err == nil {
		t.Fatal("expected error when DB fails")
	}
	if !strings.Contains(err.Error(), "database error") {
		t.Errorf("error should mention database error, got: %v", err)
	}
}

func TestWechatLogin_RequestContainsCorrectParams(t *testing.T) {
	var capturedURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		resp := map[string]string{"openid": "wx-param-test", "session_key": "sk"}
		json.NewEncoder(w).Encode(resp) //nolint:errcheck
	}))
	defer srv.Close()

	fdb := &fakeDB{
		firstUser: &users.User{Model: gorm.Model{ID: 1}, Openid: "wx-param-test"},
	}
	svc := newServiceWithURL(fdb, srv.URL)
	_, err := svc.WechatLogin(context.Background(), "mycode123", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedURL, "appid=myAppID") {
		t.Errorf("request URL missing appid, got: %s", capturedURL)
	}
	if !strings.Contains(capturedURL, "secret=myAppSecret") {
		t.Errorf("request URL missing secret, got: %s", capturedURL)
	}
	if !strings.Contains(capturedURL, "js_code=mycode123") {
		t.Errorf("request URL missing js_code, got: %s", capturedURL)
	}
	if !strings.Contains(capturedURL, "grant_type=authorization_code") {
		t.Errorf("request URL missing grant_type, got: %s", capturedURL)
	}
}
