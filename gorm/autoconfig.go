// Package gorm 提供 GORM ORM 的自动配置。
//
// 当 gorm.enabled=true 时自动启用，从 Environment 中读取 gorm.host、gorm.port、
// gorm.username、gorm.password、gorm.database、gorm.max-open-conns 等配置项，
// 创建并注册 GORM DB Bean 到 IoC 容器中（Bean ID: gormDB）。
//
// 同时会自动注册数据库健康指标（Bean ID: databaseHealthIndicator），
// 使用 PingContext 进行数据库连接检查。
package gorm

import (
	"context"
	"time"

	gormcore "github.com/xudefa/go-boot-gorm"

	"github.com/xudefa/go-boot/actuator"
	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
)

// init 注册 GORM 自动配置，由 gorm.enabled=true 条件控制。
func init() {
	boot.RegisterAutoConfig(&GormAutoConfiguration{},
		condition.OnProperty(constants.GORMEnabled, constants.ConditionTrue),
	)
}

// GormAutoConfiguration GORM ORM 的自动配置。
//
// 从 Environment 中读取 gorm.host、gorm.port、gorm.username、gorm.database 等配置项，
// 创建 MySQL 数据库连接并注册到 IoC 容器中。
// 启用条件：gorm.enabled=true
type GormAutoConfiguration struct{}

// Configure 执行自动配置逻辑，创建 GORM DB 连接并注册为 Bean。
//
// 同时注册数据库健康指标，用于监控数据库连接状态。
func (g *GormAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	opts := []gormcore.Option{
		gormcore.WithHost(env.GetString(constants.GORMHost, constants.DefaultGORMHost)),
		gormcore.WithPort(env.GetInt(constants.GORMPort, constants.DefaultGORMPort)),
		gormcore.WithUser(env.GetString(constants.GORMUsername, constants.DefaultGORMUsername)),
		gormcore.WithPassword(env.GetString(constants.GORMPassword, constants.DefaultGORMPassword)),
		gormcore.WithDBName(env.GetString(constants.GORMDatabase, constants.DefaultGORMDatabase)),
		gormcore.WithCharset(env.GetString(constants.GORMCharset, constants.DefaultGORMCharset)),
		gormcore.WithParseTime(true),
		gormcore.WithMaxOpenConns(env.GetInt(constants.GORMMaxOpenConns, constants.DefaultGORMMaxOpenConns)),
		gormcore.WithMaxIdleConns(env.GetInt(constants.GORMMaxIdleConns, constants.DefaultGORMMaxIdleConns)),
		gormcore.WithConnMaxLifetime(time.Duration(env.GetInt(constants.GORMConnMaxLifetime, constants.DefaultGORMConnMaxLifetime)) * time.Second),
	}
	if tz := env.GetString(constants.GORMTimezone, constants.DefaultGORMTimezone); tz != "" {
		opts = append(opts, gormcore.WithTimeZone(tz))
	}

	db, err := gormcore.OpenMySQL(opts...)
	if err != nil {
		panic(err)
	}

	if err := ctx.Register(constants.GORMDBBeanID,
		core.Bean(db),
		core.Singleton(),
	); err != nil {
		return err
	}

	dbHealthIndicator := actuator.NewDatabaseHealthIndicator(func(ctx context.Context) error {
		sqlDB, err := db.DB().DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	})

	if err := ctx.Register(constants.DatabaseHealthIndicatorBeanID,
		core.Bean(dbHealthIndicator),
		core.Singleton(),
	); err != nil {
		return err
	}

	return nil
}
