// Package gorm 测试 GORM 自动配置逻辑
package gorm

import (
	"testing"

	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/data"
	"github.com/xudefa/go-boot/environment"
	"github.com/xudefa/go-boot/event"
)

// mockApplicationContext 模拟应用上下文用于测试
type mockApplicationContext struct {
	container core.Container
	env       *environment.Environment
}

func (m *mockApplicationContext) Container() core.Container {
	return m.container
}

func (m *mockApplicationContext) Environment() *environment.Environment {
	return m.env
}

func (m *mockApplicationContext) EventBus() interface {
	Publish(event event.ApplicationEvent)
} {
	return nil
}

func (m *mockApplicationContext) Register(name string, opts ...core.BuilderOption) error {
	return m.container.Register(name, opts...)
}

func (m *mockApplicationContext) Get(name string) (any, error) {
	return m.container.Get(name)
}

// TestGormAutoConfiguration_Configure_WithDefaults 测试使用默认配置
func TestGormAutoConfiguration_Configure_WithDefaults(t *testing.T) {
	t.Parallel()

	container := core.New()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		constants.GORMEnabled: "true",
	}))

	ctx := &mockApplicationContext{
		container: container,
		env:       env,
	}

	config := &GormAutoConfiguration{}
	err := config.Configure(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !container.Has(constants.GORMDBBeanID) {
		t.Fatal("expected gormDB bean to be registered")
	}

	bean, err := container.Get(constants.GORMDBBeanID)
	if err != nil {
		t.Fatalf("failed to get bean: %v", err)
	}

	_, ok := bean.(*DB)
	if !ok {
		t.Fatalf("expected *DB, got %T", bean)
	}

	if !container.Has(constants.DatabaseHealthIndicatorBeanID) {
		t.Fatal("expected databaseHealthIndicator bean to be registered")
	}
}

// TestGormAutoConfiguration_DBImplementsTransactor 验证自动配置的 DB 实现了 Transactor 接口
func TestGormAutoConfiguration_DBImplementsTransactor(t *testing.T) {
	t.Parallel()

	container := core.New()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		constants.GORMEnabled: "true",
	}))

	ctx := &mockApplicationContext{
		container: container,
		env:       env,
	}

	config := &GormAutoConfiguration{}
	err := config.Configure(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bean, err := container.Get(constants.GORMDBBeanID)
	if err != nil {
		t.Fatalf("failed to get bean: %v", err)
	}

	db, ok := bean.(*DB)
	if !ok {
		t.Fatalf("expected *DB, got %T", bean)
	}

	var _ data.Transactor = db
	transactor := db.Transactor()
	if transactor == nil {
		t.Fatal("transactor should not be nil")
	}
}

var _ boot.ApplicationContext = (*mockApplicationContext)(nil)
