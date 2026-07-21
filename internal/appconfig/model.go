package appconfig

// singletonID 是 app_config 单行配置的固定主键。
// 全局仅一行，读写都锁定该 ID，避免出现多行歧义。
const singletonID uint = 1

// AppConfig 全局应用配置（固定单行）。
// 运行时可通过更新该行的 audit_mode 列切换审核模式，无需发版。
type AppConfig struct {
	ID uint `gorm:"primaryKey"`
	// AuditMode 审核模式开关。true=审核模式（隐藏互动板块），false=正常模式。
	AuditMode bool `gorm:"column:audit_mode;not null;default:false"`
}

// TableName 返回表名。
func (AppConfig) TableName() string { return "app_config" }

// AppConfigResponse 是应用配置 API 响应。
type AppConfigResponse struct {
	AuditMode bool `json:"auditMode"`
}
