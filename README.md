# go-boot-gorm

[![Go Version](https://img.shields.io/github/go-mod/go-version/xudefa/go-boot-gorm)](https://go.dev/) [![License](https://img.shields.io/github/license/xudefa/go-boot-gorm)](./LICENSE) [![Build Status](https://img.shields.io/github/actions/workflow/status/xudefa/go-boot-gorm/test.yml?branch=master)](https://github.com/xudefa/go-boot-gorm/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/xudefa/go-boot-gorm.svg)](https://pkg.go.dev/github.com/xudefa/go-boot-gorm) [![Go Report Card](https://goreportcard.com/badge/github.com/xudefa/go-boot-gorm)](https://goreportcard.com/report/github.com/xudefa/go-boot-gorm)

基于 [go-boot](https://github.com/xudefa/go-boot) 的 GORM ORM 集成模块。将 GORM 无缝集成到 go-boot 的 IoC 容器和自动配置体系中，提供自动配置、泛型 Repository、事务管理和数据库健康检查能力。

> 设计理念：遵循 go-boot 的开发规范，将 GORM 作为 `data.Transactor` 接口的实现，通过自动配置实现零代码初始化数据库连接。

## 整体架构

```
┌───────────────────────────────────────────────────────────────────────┐
│                    go-boot ApplicationContext                         │
│  ┌───────────┐ ┌──────────────┐ ┌───────────┐ ┌───────────┐           │
│  │ Container │ │  Environment │ │ Lifecycle │ │ EventBus  │           │
│  └───────────┘ └──────────────┘ └───────────┘ └───────────┘           │
│                       ┌─────────────────────┐                         │
│                       │ AutoConfig Registry │                         │
│                       └─────────────────────┘                         │
└───────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │    go-boot-gorm Starter       │
                    │  ┌─────────────────────────┐  │
                    │  │ GormDB Bean             │  │
                    │  │ (data.Transactor)       │  │
                    │  │ Repository[T]           │  │
                    │  │ Health Indicator        │  │
                    │  └─────────────────────────┘  │
                    └───────────────────────────────┘
```

## 目录

- [快速开始](#快速开始)
- [功能特性](#功能特性)
- [使用示例](#使用示例)
- [配置选项](#配置选项)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [贡献](#贡献)
- [许可证](#许可证)

## 快速开始

### 安装

```bash
# 安装核心框架
go get github.com/xudefa/go-boot

# 安装 GORM 集成模块
go get github.com/xudefa/go-boot-gorm
```

### 最小示例

```go
package main

import (
    "context"
    "fmt"

    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot-gorm"
)

type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:256"`
    Age  int
}

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-app"),
        boot.WithVersion("1.0.0"),
        boot.WithProperty("gorm.enabled", "true"),
        boot.WithProperty("gorm.host", "localhost"),
        boot.WithProperty("gorm.port", "3306"),
        boot.WithProperty("gorm.username", "root"),
        boot.WithProperty("gorm.password", "123456"),
        boot.WithProperty("gorm.database", "mydb"),
    )
    if err != nil {
        panic(err)
    }
    defer app.Stop()

    app.Start()

    // 从容器获取 GORM DB
    db := app.Container().Get("gormDB").(*gorm.DB)

    // 自动迁移表结构
    db.DB().AutoMigrate(&User{})

    // 创建泛型 Repository
    repo := gorm.NewRepository[User](db.DB())

    // 创建实体
    user := &User{Name: "John", Age: 30}
    if err := repo.Create(user); err != nil {
        panic(err)
    }
    fmt.Printf("Created user: %s\n", user.Name)

    // 查询实体
    found, err := repo.FindByID(user.ID)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Found user: %s, Age: %d\n", found.Name, found.Age)

    app.WaitForSignal()
}
```

## 功能特性

| 特性 | 说明 |
|------|------|
| GORM 集成 | 将 GORM DB 注册为 Bean，支持依赖注入 |
| data.Transactor 实现 | GORM DB 实现 go-boot 的 `data.Transactor` 接口 |
| 自动配置 | 通过 `gorm.enabled=true` 自动初始化数据库连接 |
| 泛型 Repository | 提供 `Repository[T]` 泛型 CRUD 操作 |
| 事务管理 | 支持 `data.Transaction` 接口的事务操作 |
| 多数据库支持 | 支持 MySQL、PostgreSQL、SQLServer、SQLite |
| 健康检查 | 自动注册数据库健康指标 |
| 连接池管理 | 支持连接池配置（最大连接数、空闲连接数、生命周期） |

## 使用示例

### 直接操作数据库

```go
db := app.Container().Get("gormDB").(*gorm.DB)

// 执行原生 SQL
rows, err := db.Query(context.Background(), "SELECT id, name FROM users WHERE age > ?", 18)
if err != nil {
    panic(err)
}
defer rows.Close()

for rows.Next() {
    var id int
    var name string
    if err := rows.Scan(&id, &name); err != nil {
        panic(err)
    }
    fmt.Printf("User: %d, %s\n", id, name)
}
```

### 使用泛型 Repository

```go
repo := gorm.NewRepository[User](db.DB())

// 创建
user := &User{Name: "Alice", Age: 25}
repo.Create(user)

// 批量创建
users := []User{
    {Name: "Bob", Age: 30},
    {Name: "Charlie", Age: 35},
}
repo.CreateBatch(users)

// 查询
found, _ := repo.FindByID(user.ID)
all, _ := repo.FindAll(nil)
count, _ := repo.Count("age > ?", 20)

// 更新
found.Age = 26
repo.Update(found)

// 删除
repo.Delete(user.ID)
```

### 事务操作

```go
db := app.Container().Get("gormDB").(*gorm.DB)

// 开始事务
tx, err := db.Begin(context.Background())
if err != nil {
    panic(err)
}
defer tx.Close()

// 在事务中执行操作
repo := gorm.NewRepositoryWithTx[User](tx.(*gorm.Transaction).DB())
user := &User{Name: "Transactional User", Age: 28}
if err := repo.Create(user); err != nil {
    tx.Rollback()
    return
}

// 提交事务
if err := tx.Commit(); err != nil {
    panic(err)
}
```

### 手动创建数据库连接

```go
// MySQL
db, err := gorm.OpenMySQL(
    gorm.WithHost("localhost"),
    gorm.WithPort(3306),
    gorm.WithUser("root"),
    gorm.WithPassword("123456"),
    gorm.WithDBName("mydb"),
    gorm.WithCharset("utf8mb4"),
    gorm.WithParseTime(true),
    gorm.WithMaxOpenConns(100),
    gorm.WithMaxIdleConns(10),
    gorm.WithConnMaxLifetime(3600 * time.Second),
)

// PostgreSQL
db, err := gorm.OpenPostgreSQL(
    gorm.WithHost("localhost"),
    gorm.WithPort(5432),
    gorm.WithUser("postgres"),
    gorm.WithPassword("123456"),
    gorm.WithDBName("mydb"),
    gorm.WithSSLMode("disable"),
)

// SQLite
db, err := gorm.OpenSQLite(
    gorm.WithDBName("mydb.sqlite"),
)
```

## 配置选项

通过 `boot.WithProperty()` 或配置文件设置：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `gorm.enabled` | `false` | 是否启用 GORM |
| `gorm.host` | `localhost` | 数据库主机地址 |
| `gorm.port` | `3306` | 数据库端口 |
| `gorm.username` | `root` | 数据库用户名 |
| `gorm.password` | `123456` | 数据库密码 |
| `gorm.database` | `test` | 数据库名称 |
| `gorm.charset` | `utf8mb4` | 字符集 |
| `gorm.timezone` | `Local` | 时区 |
| `gorm.max-open-conns` | `100` | 最大打开连接数 |
| `gorm.max-idle-conns` | `10` | 最大空闲连接数 |
| `gorm.conn-max-lifetime` | `3600` | 连接最大生命周期（秒） |

### 示例配置

```yaml
# application.yml
gorm:
  enabled: true
  host: localhost
  port: 3306
  username: root
  password: 123456
  database: mydb
  charset: utf8mb4
  timezone: Asia/Shanghai
  max-open-conns: 100
  max-idle-conns: 10
  conn-max-lifetime: 3600
```

## 项目结构

```
go-boot-gorm/
├── gorm.go              # GORM 核心实现（DB、Repository、Transaction）
├── gorm_starter.go      # GORM 启动器
├── autoconfig.go        # GORM 自动配置
├── model.go             # 数据模型辅助
├── gorm_test.go         # 单元测试
├── README.md
├── LICENSE
└── go.mod
```

## 开发指南

### 构建

```bash
go build ./...
```

### 测试

```bash
go test ./...
go test -cover ./...       # 带覆盖率
go test -race ./...        # 数据竞争检测
```

### 代码规范

```bash
go fmt ./...
golangci-lint run
```

## 贡献

欢迎提交 Issue 和 Pull Request！详细贡献指南请参阅 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 许可证

本项目采用 MIT 许可证 — 详情请参阅 [LICENSE](./LICENSE) 文件。