package subscribepush

import (
	"context"
	"errors"
	"testing"
	"time"
)

// --- fakes ---

type fakeSender struct {
	called  bool
	openid  string
	tplID   string
	data    map[string]any
	page    string
	sendErr error
}

func (f *fakeSender) SendSubscribeMessage(ctx context.Context, openid, tplID string, data map[string]any, page string) error {
	f.called = true
	f.openid = openid
	f.tplID = tplID
	f.data = data
	f.page = page
	return f.sendErr
}

type fakeStore struct {
	openid     string
	quota      int
	lookupErr  error
	decrErr    error
	decrCalled bool
	nickname   string
}

func (f *fakeStore) GetSubscribeTarget(ctx context.Context, userID uint) (openid string, quota int, err error) {
	return f.openid, f.quota, f.lookupErr
}
func (f *fakeStore) DecrSubscribeQuota(ctx context.Context, userID uint) error {
	f.decrCalled = true
	return f.decrErr
}
func (f *fakeStore) GetNickname(ctx context.Context, userID uint) (string, error) {
	return f.nickname, nil
}

// pushFollowSync 是 PushFollow 的同步核心（测试直接调用，避免 goroutine 时序）。

func TestPush_WithQuota_Sends(t *testing.T) {
	sender := &fakeSender{}
	store := &fakeStore{openid: "openid-a", quota: 2, nickname: "b"}
	p := New(sender, store, "tpl-123")

	p.pushFollowSync(context.Background(), 1, 2)

	if !sender.called {
		t.Error("expected SendSubscribeMessage to be called")
	}
	if sender.openid != "openid-a" || sender.tplID != "tpl-123" {
		t.Errorf("wrong openid/tpl: %s / %s", sender.openid, sender.tplID)
	}
	if sender.page != "pages/notifications/index" {
		t.Errorf("expected jump page pages/notifications/index, got %q", sender.page)
	}
	if !store.decrCalled {
		t.Error("expected quota decremented after send")
	}
}

func TestPush_LongNickname_Truncated(t *testing.T) {
	sender := &fakeSender{}
	long := "这是一个非常非常非常非常非常非常长的昵称超过二十个字符了" // >20 runes
	store := &fakeStore{openid: "openid-a", quota: 1, nickname: long}
	p := New(sender, store, "tpl-123")

	p.pushFollowSync(context.Background(), 1, 2)

	thing1, _ := sender.data["thing1"].(map[string]string)
	if got := []rune(thing1["value"]); len(got) > 20 {
		t.Errorf("thing1 not truncated to 20 runes: got %d", len(got))
	}
}

func TestPush_NoQuota_DoesNotSend(t *testing.T) {
	sender := &fakeSender{}
	store := &fakeStore{openid: "openid-a", quota: 0}
	p := New(sender, store, "tpl-123")

	p.pushFollowSync(context.Background(), 1, 2)

	if sender.called {
		t.Error("expected no send when quota is 0")
	}
	if store.decrCalled {
		t.Error("expected no quota decrement when quota is 0")
	}
}

func TestPush_EmptyOpenid_DoesNotSend(t *testing.T) {
	sender := &fakeSender{}
	store := &fakeStore{openid: "", quota: 3}
	p := New(sender, store, "tpl-123")

	p.pushFollowSync(context.Background(), 1, 2)

	if sender.called {
		t.Error("expected no send when openid is empty")
	}
}

func TestPush_SendFails_NoDecr(t *testing.T) {
	sender := &fakeSender{sendErr: errors.New("wechat 43101")}
	store := &fakeStore{openid: "openid-a", quota: 1, nickname: "b"}
	p := New(sender, store, "tpl-123")

	p.pushFollowSync(context.Background(), 1, 2)

	if !sender.called {
		t.Error("expected send attempted")
	}
	if store.decrCalled {
		t.Error("expected no quota decrement when send fails")
	}
}

func TestPush_LookupFails_DoesNotSend(t *testing.T) {
	sender := &fakeSender{}
	store := &fakeStore{lookupErr: errors.New("db down")}
	p := New(sender, store, "tpl-123")

	p.pushFollowSync(context.Background(), 1, 2)

	if sender.called {
		t.Error("expected no send when target lookup fails")
	}
}

// PushFollow (async) 不应阻塞调用方；这里只验证它能被调用且很快返回。
func TestPushFollow_Async_ReturnsImmediately(t *testing.T) {
	sender := &fakeSender{}
	store := &fakeStore{openid: "openid-a", quota: 1, nickname: "b"}
	p := New(sender, store, "tpl-123")

	done := make(chan struct{})
	go func() {
		p.PushFollow(context.Background(), 1, 2)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("PushFollow did not return promptly")
	}
}
