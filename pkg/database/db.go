package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接实例
var DB *gorm.DB

// Connect 初始化数据库连接
func Connect(dsn string) error {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}
	DB = db
	return nil
}

// AutoMigrate 同步所有数据模型表结构
// models 参数接收所有需要迁移的模型指针
func AutoMigrate(models ...interface{}) error {
	return DB.AutoMigrate(models...)
}
