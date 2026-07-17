package posts

import (
	"testing"
)

// Note: The PostService now directly uses *gorm.DB for full transaction support.
// Integration tests require a real or in-memory database connection.
// Unit tests for the service layer will be added when a test DB fixture is available.

func TestPostStatus_Values(t *testing.T) {
	if PostStatusDraft != "DRAFT" {
		t.Errorf("expected PostStatusDraft = 'DRAFT', got %s", PostStatusDraft)
	}
	if PostStatusPublished != "PUBLISHED" {
		t.Errorf("expected PostStatusPublished = 'PUBLISHED', got %s", PostStatusPublished)
	}
}

func TestCreatePostInput_Defaults(t *testing.T) {
	input := CreatePostInput{}
	if input.Title != "" {
		t.Errorf("expected empty title by default")
	}
	if input.Status != "" {
		t.Errorf("expected empty status by default")
	}
}
