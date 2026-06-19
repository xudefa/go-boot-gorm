package gorm

import src "github.com/xudefa/go-boot"

// Starter 实现 gorm 数据库连接启动器。
//
// 用于在应用启动时初始化数据库连接并运行迁移。
type Starter struct {
	db          *DB
	autoMigrate []any
}

// NewStarter 创建新的数据库启动器。
//
// 参数:
//   - db: 数据库连接
//   - models: 要迁移的模型
//
// 返回值:
//   - *Starter: 启动器实例
func NewStarter(db *DB, models []any) *Starter {
	return &Starter{
		db:          db,
		autoMigrate: models,
	}
}

// Starter 实现 boot.Starter 接口。
func (s *Starter) Starter() error {
	if s.db == nil {
		return nil
	}
	if len(s.autoMigrate) > 0 {
		return s.db.db.AutoMigrate(s.autoMigrate...)
	}
	return nil
}

// WithStarter 创建数据库启动器的选项。
func WithStarter(db *DB, models ...any) *Starter {
	return NewStarter(db, models)
}

// NewAutoMigrateStarter 创建自动迁移启动器。
func NewAutoMigrateStarter(db *DB, models ...any) src.Starter {
	return &Starter{
		db:          db,
		autoMigrate: models,
	}
}
