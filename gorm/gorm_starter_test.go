// gorm_starter 测试
package gorm

import (
	"testing"

	src "github.com/xudefa/go-boot"
)

// TestStarter_Starter_WithNilEngine 测试 nil engine 时不报错
func TestStarter_Starter_WithNilEngine(t *testing.T) {
	t.Parallel()

	starter := &Starter{
		db:          nil,
		autoMigrate: nil,
	}

	err := starter.Starter()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// TestStarter_Starter_WithEmptyModels 测试空模型列表时不执行迁移
func TestStarter_Starter_WithEmptyModels(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	starter := &Starter{
		db:          db,
		autoMigrate: []any{},
	}

	err = starter.Starter()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// TestStarter_Starter_WithModels 测试有模型时调用 AutoMigrate
func TestStarter_Starter_WithModels(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	type TestStarterModel struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	starter := &Starter{
		db:          db,
		autoMigrate: []any{&TestStarterModel{}},
	}

	err = starter.Starter()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// TestNewStarter 测试创建 Starter
func TestNewStarter(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	starter := NewStarter(db, nil)
	if starter == nil {
		t.Fatal("NewStarter should return starter")
	}
	if starter.db != db {
		t.Error("starter db should match")
	}
}

// TestNewAutoMigrateStarter 测试创建自动迁移 Starter
func TestNewAutoMigrateStarter(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	type TestModel struct {
		ID uint `gorm:"primaryKey"`
	}

	starter := NewAutoMigrateStarter(db, &TestModel{})
	if starter == nil {
		t.Fatal("NewAutoMigrateStarter should return starter")
	}

	_, ok := starter.(src.Starter)
	if !ok {
		t.Error("starter should implement src.Starter")
	}
}

// TestStarter_ImplementsStarterInterface 验证 Starter 实现了 boot.Starter 接口
func TestStarter_ImplementsStarterInterface(t *testing.T) {
	t.Parallel()

	var _ src.Starter = (*Starter)(nil)
}
