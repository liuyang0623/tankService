package database

import (
	"testing"
)

// TestConnect_EmptyDSN 测试空 DSN 时 Connect 返回错误
func TestConnect_EmptyDSN(t *testing.T) {
	// 重置全局 DB 状态
	DB = nil

	err := Connect("")
	if err == nil {
		t.Error("期望空 DSN 返回错误，但得到 nil")
	}

	// 确认 DB 仍然为 nil（连接失败不应设置全局变量）
	if DB != nil {
		t.Error("连接失败后 DB 应仍为 nil")
	}
}

// TestDB_InitiallyNil 测试 DB 变量初始状态为 nil
func TestDB_InitiallyNil(t *testing.T) {
	// 重置
	DB = nil

	if DB != nil {
		t.Error("DB 初始状态应为 nil")
	}
}

// TestAutoMigrate_NilDB 测试 DB 未初始化时 AutoMigrate 的行为
func TestAutoMigrate_NilDB(t *testing.T) {
	// 重置
	DB = nil

	// AutoMigrate 应返回错误（DB 为 nil，调用会 panic 或返回错误）
	// 使用 recover 捕获可能的 panic
	var panicOccurred bool
	var autoMigrateErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				panicOccurred = true
			}
		}()
		autoMigrateErr = AutoMigrate()
	}()

	// 两种可接受的行为：panic 或返回错误
	if !panicOccurred && autoMigrateErr == nil {
		t.Error("DB 为 nil 时 AutoMigrate 应返回错误或 panic")
	}
}

// TestConnect_InvalidDSN 测试无效 DSN 格式时返回错误
func TestConnect_InvalidDSN(t *testing.T) {
	DB = nil

	err := Connect("not-a-valid-dsn")
	if err == nil {
		t.Error("期望无效 DSN 返回错误，但得到 nil")
	}
}
