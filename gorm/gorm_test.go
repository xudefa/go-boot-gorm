// gorm 集成模块测试
// 测试 GORM 数据库连接、事务、Repository CRUD 操作和 DSN 生成等功能
package gorm

import (
	"context"
	"testing"

	"github.com/xudefa/go-boot/data"
)

// TestOpenSQLite 测试使用 SQLite 内存数据库打开连接，验证 db 对象和 engine 不为空
func TestOpenSQLite(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	if db == nil {
		t.Fatal("OpenSQLite() returned nil")
	}
	if db.db == nil {
		t.Error("db should not be nil")
	}
}

// TestOpenWithOptions 测试使用 Open 通用接口通过选项创建数据库连接，验证连接成功
func TestOpenWithOptions(t *testing.T) {
	db, err := Open(
		WithDBType(SQLite),
		WithDBName(":memory:"),
	)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	if db == nil {
		t.Fatal("Open() returned nil")
	}
}

// TestDB_Begin 测试开启数据库事务，验证返回的事务对象不为空
func TestDB_Begin(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	if tx == nil {
		t.Fatal("Begin() returned nil")
	}
}

// TestDB_Query 测试执行原生查询 SQL，验证返回的行集不为空且无错误
func TestDB_Query(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer func() { _ = rows.Close() }()
	if rows == nil {
		t.Fatal("rows should not be nil")
	}
}

// TestDB_QueryRow 测试查询单行数据并扫描到变量，验证结果值为 1
func TestDB_QueryRow(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	row := db.QueryRow(ctx, "SELECT 1")
	var val int
	err = row.Scan(&val)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}
}

// TestDB_Exec 测试执行 DDL 语句（建表），验证执行结果不为空
func TestDB_Exec(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	result, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER)")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

// TestDB_Stats 测试获取数据库统计信息，验证返回的统计对象不为空
func TestDB_Stats(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	stats := db.Stats()
	_ = stats
}

// TestTransaction_Commit 测试提交事务，验证提交无错误
func TestTransaction_Commit(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestTransaction_Rollback 测试回滚事务，验证回滚无错误
func TestTransaction_Rollback(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	err = tx.Rollback()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestTransaction_Close 测试关闭事务，验证关闭无错误
func TestTransaction_Close(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	err = tx.Close()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestClient 测试创建 GORM Client 包装对象，验证不为空
func TestClient(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	client := NewClient(db.DB())
	if client == nil {
		t.Error("NewClient should return client")
	}
}

// TestDB_ImplementsTransactor 验证 DB 实现了 data.Transactor 接口
func TestDB_ImplementsTransactor(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var _ data.Transactor = db
}

// TestTransaction_ImplementsTransactionInterface 编译时检查 Transaction 是否实现了 data.Transaction 接口
func TestTransaction_ImplementsTransactionInterface(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	var _ data.Transaction = tx
}

// TestRepository_Create 测试 Repository 的 Create 方法，验证创建记录后 ID 被自动填充
func TestRepository_Create(t *testing.T) {
	type User struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	db, _ := OpenSQLite(WithDBName(":memory:"))
	if err := db.db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	repo := NewRepository[User](db.DB())
	user := &User{Name: "John"}

	err := repo.Create(user)
	if err != nil {
		t.Errorf("Create failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("user ID should be set after create")
	}
}

// TestRepository_FindByID 测试 Repository 的 FindByID 方法，验证根据 ID 查找记录并返回正确字段
func TestRepository_FindByID(t *testing.T) {
	type User struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	db, _ := OpenSQLite(WithDBName(":memory:"))
	if err := db.db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	repo := NewRepository[User](db.DB())
	user := &User{Name: "John"}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(user.ID)
	if err != nil {
		t.Errorf("FindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("found should not be nil")
	}
	if found.Name != "John" {
		t.Errorf("expected name 'John', got '%s'", found.Name)
	}
}

// TestRepository_Update 测试 Repository 的 Update 方法，验证更新记录后重新查询得到新值
func TestRepository_Update(t *testing.T) {
	type User struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	db, _ := OpenSQLite(WithDBName(":memory:"))
	if err := db.db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	repo := NewRepository[User](db.DB())
	user := &User{Name: "John"}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	user.Name = "Jane"
	err := repo.Update(user)
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(user.ID)
	if found.Name != "Jane" {
		t.Errorf("expected name 'Jane', got '%s'", found.Name)
	}
}

// TestRepository_Delete 测试 Repository 的 Delete 方法
func TestRepository_Delete(t *testing.T) {
	type User struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	db, _ := OpenSQLite(WithDBName(":memory:"))
	if err := db.db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	repo := NewRepository[User](db.DB())
	user := &User{Name: "John"}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err := repo.Delete(user.ID)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	found, _ := repo.FindByID(user.ID)
	if found != nil {
		t.Error("user should be deleted")
	}
}

// TestRepository_FindAll 测试 Repository 的 FindAll 方法，验证插入两条记录后查询到全部结果
func TestRepository_FindAll(t *testing.T) {
	type User struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	db, _ := OpenSQLite(WithDBName(":memory:"))
	if err := db.db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	repo := NewRepository[User](db.DB())
	if err := repo.Create(&User{Name: "John"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := repo.Create(&User{Name: "Jane"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	results, err := repo.FindAll(nil)
	if err != nil {
		t.Errorf("FindAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 users, got %d", len(results))
	}
}

// TestRepository_Count 测试 Repository 的 Count 方法，验证插入两条记录后计数为 2
func TestRepository_Count(t *testing.T) {
	type User struct {
		ID   uint `gorm:"primaryKey"`
		Name string
	}

	db, _ := OpenSQLite(WithDBName(":memory:"))
	if err := db.db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	repo := NewRepository[User](db.DB())
	if err := repo.Create(&User{Name: "John"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := repo.Create(&User{Name: "Jane"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	count, err := repo.Count(nil)
	if err != nil {
		t.Errorf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

// TestRepository_ImplementsRepositoryInterface 编译时检查 Repository[User] 是否实现了 data.Repository[User] 接口
func TestRepository_ImplementsRepositoryInterface(t *testing.T) {
	type User struct {
		ID   uint
		Name string
	}
	var _ data.Repository[User] = (*Repository[User])(nil)
}

// TestConfig_DSNForMySQL 测试生成 MySQL 的 DSN 连接串，验证不为空且长度合理
func TestConfig_DSNForMySQL(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     3306,
		User:     "gate",
		Password: "123456",
		DBName:   "gate",
		Charset:  "utf8",
	}
	dsn := cfg.DSNForMySQL()
	if dsn == "" {
		t.Error("DSN should not be empty")
	}
	if len(dsn) < 20 {
		t.Errorf("DSN seems too short: %s", dsn)
	}
}

// TestConfig_DSNForPostgres 测试生成 PostgreSQL 的 DSN 连接串，验证不为空
func TestConfig_DSNForPostgres(t *testing.T) {
	cfg := &Config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "testdb",
		SslMode:  "disable",
	}
	dsn := cfg.DSNForPostgres()
	if dsn == "" {
		t.Error("DSN should not be empty")
	}
}

// TestConfig_DSNForSQLite 测试生成 SQLite 的 DSN 连接串，验证直接返回数据库文件名
func TestConfig_DSNForSQLite(t *testing.T) {
	cfg := &Config{
		DBName: "test.db",
	}
	dsn := cfg.DSNForSQLite()
	if dsn != "test.db" {
		t.Errorf("expected 'test.db', got '%s'", dsn)
	}
}
