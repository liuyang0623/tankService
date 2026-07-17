package users

import (
	"reflect"
	"strings"
	"testing"

	"gorm.io/gorm"
)

func TestUserStruct_Fields(t *testing.T) {
	user := User{}
	rt := reflect.TypeOf(user)

	// gorm.Model embeds ID, CreatedAt, UpdatedAt, DeletedAt
	// Verify gorm.Model is embedded
	_, hasGormModel := rt.FieldByName("Model")
	if !hasGormModel {
		t.Error("User should embed gorm.Model")
	}

	expectedFields := []struct {
		name       string
		kind       reflect.Kind
		tagContains string
	}{
		{"Openid", reflect.String, "uniqueIndex"},
		{"Unionid", reflect.String, "varchar(100)"},
		{"SessionKey", reflect.String, "varchar(100)"},
		{"Nickname", reflect.String, "not null"},
		{"Avatar", reflect.String, "varchar(500)"},
		{"Phone", reflect.String, "varchar(20)"},
		{"Bio", reflect.String, "default:''"},
		{"Gender", reflect.Int, "default:0"},
	}

	for _, ef := range expectedFields {
		field, ok := rt.FieldByName(ef.name)
		if !ok {
			t.Errorf("User should have field %s", ef.name)
			continue
		}
		if field.Type.Kind() != ef.kind {
			t.Errorf("Field %s should be kind %s, got %s", ef.name, ef.kind, field.Type.Kind())
		}
		tagVal := field.Tag.Get("gorm")
		if !strings.Contains(tagVal, ef.tagContains) {
			t.Errorf("Field %s gorm tag should contain %q, got %q", ef.name, ef.tagContains, tagVal)
		}
	}
}

func TestUserStruct_GormModelEmbedded(t *testing.T) {
	// Verify that User embeds gorm.Model (has ID, CreatedAt, UpdatedAt, DeletedAt via embedding)
	user := User{}
	rt := reflect.TypeOf(user)

	modelField, ok := rt.FieldByName("Model")
	if !ok {
		t.Fatal("User must embed gorm.Model")
	}
	if modelField.Type != reflect.TypeOf(gorm.Model{}) {
		t.Errorf("Model field type should be gorm.Model, got %v", modelField.Type)
	}
}

func TestUser_TableName(t *testing.T) {
	u := User{}
	if u.TableName() != "users" {
		t.Errorf("TableName() should return 'users', got %q", u.TableName())
	}
}
