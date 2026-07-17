package posts

import (
	"reflect"
	"testing"
)

func TestPostStruct_Fields(t *testing.T) {
	p := Post{}
	typ := reflect.TypeOf(p)

	expected := map[string]string{
		"Title":        "string",
		"Content":      "string",
		"Cover":        "string",
		"Status":       "posts.PostStatus",
		"AuthorID":     "uint",
		"ViewCount":    "int",
		"LikeCount":    "int",
		"CommentCount": "int",
	}

	for fieldName, expectedType := range expected {
		f, ok := typ.FieldByName(fieldName)
		if !ok {
			t.Errorf("expected Post to have field %s", fieldName)
			continue
		}
		actualType := f.Type.String()
		if actualType != expectedType {
			t.Errorf("expected Post.%s to be %s, got %s", fieldName, expectedType, actualType)
		}
	}
}

func TestPostStruct_GormModelEmbedded(t *testing.T) {
	_, ok := reflect.TypeOf(Post{}).FieldByName("Model")
	if !ok {
		t.Error("expected Post to embed gorm.Model")
	}
}

func TestPost_TableName(t *testing.T) {
	p := Post{}
	if p.TableName() != "posts" {
		t.Errorf("expected table name 'posts', got %s", p.TableName())
	}
}

func TestPostStatus_Constants(t *testing.T) {
	if PostStatusDraft != "DRAFT" {
		t.Errorf("expected PostStatusDraft = 'DRAFT', got %s", PostStatusDraft)
	}
	if PostStatusPublished != "PUBLISHED" {
		t.Errorf("expected PostStatusPublished = 'PUBLISHED', got %s", PostStatusPublished)
	}
}

func TestPostImageStruct_Fields(t *testing.T) {
	pi := PostImage{}
	typ := reflect.TypeOf(pi)

	expected := map[string]string{
		"PostID":    "uint",
		"URL":       "string",
		"SortOrder": "int",
	}

	for fieldName, expectedType := range expected {
		f, ok := typ.FieldByName(fieldName)
		if !ok {
			t.Errorf("expected PostImage to have field %s", fieldName)
			continue
		}
		actualType := f.Type.String()
		if actualType != expectedType {
			t.Errorf("expected PostImage.%s to be %s, got %s", fieldName, expectedType, actualType)
		}
	}
}

func TestPostImage_TableName(t *testing.T) {
	pi := PostImage{}
	if pi.TableName() != "post_images" {
		t.Errorf("expected table name 'post_images', got %s", pi.TableName())
	}
}

func TestTopicStruct_Fields(t *testing.T) {
	tp := Topic{}
	typ := reflect.TypeOf(tp)

	f, ok := typ.FieldByName("Name")
	if !ok {
		t.Fatal("expected Topic to have field Name")
	}
	if f.Type.String() != "string" {
		t.Errorf("expected Topic.Name to be string, got %s", f.Type.String())
	}
}

func TestTopic_TableName(t *testing.T) {
	tp := Topic{}
	if tp.TableName() != "topics" {
		t.Errorf("expected table name 'topics', got %s", tp.TableName())
	}
}

func TestIsValidCategory(t *testing.T) {
	cases := []struct {
		cat  string
		want bool
	}{
		{"", true},         // empty = uncategorized, valid
		{"story", true},
		{"daily", true},
		{"tech", true},
		{"food", true},
		{"travel", true},
		{"invalid", false},
		{"Story", false}, // case-sensitive
		{"故事", false},
	}
	for _, c := range cases {
		if got := IsValidCategory(c.cat); got != c.want {
			t.Errorf("IsValidCategory(%q) = %v, want %v", c.cat, got, c.want)
		}
	}
}

func TestFixedCategories(t *testing.T) {
	if len(FixedCategories) != 5 {
		t.Fatalf("expected 5 fixed categories, got %d", len(FixedCategories))
	}
	// 顺序与 value/label 对应
	expect := []CategoryInfo{
		{Value: "story", Label: "故事"},
		{Value: "daily", Label: "日常"},
		{Value: "tech", Label: "技术"},
		{Value: "food", Label: "美食"},
		{Value: "travel", Label: "旅游"},
	}
	for i, e := range expect {
		if FixedCategories[i] != e {
			t.Errorf("FixedCategories[%d] = %+v, want %+v", i, FixedCategories[i], e)
		}
	}
	// 所有 value 都应通过 IsValidCategory
	for _, c := range FixedCategories {
		if !IsValidCategory(c.Value) {
			t.Errorf("FixedCategories value %q should be valid", c.Value)
		}
	}
}

func TestPost_CategoryField(t *testing.T) {
	typ := reflect.TypeOf(Post{})
	f, ok := typ.FieldByName("Category")
	if !ok {
		t.Fatal("expected Post to have field Category")
	}
	if f.Type.String() != "string" {
		t.Errorf("expected Post.Category to be string, got %s", f.Type.String())
	}
}
