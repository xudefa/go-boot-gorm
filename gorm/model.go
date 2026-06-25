package gorm

// BaseModel GORM 模块的基础模型
//
// 提供通用的主键、创建时间、更新时间和软删除字段，
// 业务模型可嵌入此结构体以获得标准审计字段。
type BaseModel struct {
	ID        uint   `gorm:"primaryKey" json:"id"`             // 主键
	CreatedAt int64  `gorm:"autoCreateTime" json:"created_at"` // 创建时间（Unix 时间戳）
	UpdatedAt int64  `gorm:"autoUpdateTime" json:"updated_at"` // 更新时间（Unix 时间戳）
	DeletedAt *int64 `gorm:"index" json:"deleted_at"`          // 软删除时间，nil 表示未删除
}
