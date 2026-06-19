// Package gorm 基于 GORM 提供数据库访问层实现。
//
// 该包将 GORM 与 go-boot 数据访问层接口集成，
// 支持多种数据库、事务和 Repository 模式。
//
// 定义：
//
//   - DB: 数据库连接实现了 data.Transactor 接口
//   - Transaction: 事务实现了 data.Transaction 接口
//   - Repository[T]: 泛型 Repository 实现了 data.Repository[T] 接口
//   - Option: 数据库配置选项
//
// 支持的数据库类型：
//
//   - MySQL: OpenMySQL()
//   - PostgreSQL: OpenPostgreSQL()
//   - SQLServer: OpenSQLServer()
//   - SQLite: OpenSQLite() (默认)
//
// 快速开始:
//
//	// 创建数据库连接
//	db, _ := gorm.OpenMySQL(
//	    gorm.WithHost("localhost"),
//	    gorm.WithPort(3306),
//	    gorm.WithUser("gate"),
//	    gorm.WithPassword("123456"),
//	    gorm.WithDBName("gate"),
//	)
//
//	// 创建 Repository
//	repo := gorm.NewRepository[User](db.DB())
//
//	// CRUD 操作
//	user := &User{Name: "John"}
//	repo.Create(user)
package gorm

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/xudefa/go-boot/data"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// DB 是 GORM 数据库连接，实现了 data.Transactor 接口。
//
// 字段说明:
//   - db: GORM 数据库实例
type DB struct {
	db *gorm.DB
}

// Transactor 返回 data.Transactor 接口
func (d *DB) Transactor() data.Transactor {
	return d
}

// Open 打开数据库连接。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *DB: 数据库连接实例
//   - error: 连接错误
//
// 示例:
//
//	db, err := gorm.Open(
//	    gorm.WithHost("localhost"),
//	    gorm.WithPort(3306),
//	    gorm.WithUser("gate"),
//	    gorm.WithPassword("123456"),
//	    gorm.WithDBName("gate"),
//	)
func Open(opts ...Option) (*DB, error) {
	cfg := &Config{Type: string(SQLite), DSN: ":memory:"}
	for _, opt := range opts {
		opt(cfg)
	}

	var dialector gorm.Dialector
	switch cfg.Type {
	case string(MySQL):
		dialector = mysql.Open(cfg.DSN)
	case string(PostgreSQL):
		dialector = postgres.Open(cfg.DSN)
	case string(SQLServer):
		dialector = sqlserver.Open(cfg.DSN)
	case string(SQLite):
		dialector = sqlite.Open(cfg.DSN)
	default:
		dialector = mysql.Open(cfg.DSN)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	applyPoolConfig(db, cfg)
	return &DB{db: db}, nil
}

// Config 定义数据库连接配置。
//
// 字段说明:
//   - Type: 数据库类型 (mysql/postgres/sqlserver/sqlite)
//   - DSN: 数据源名称 (自定义连接字符串)
//   - Host: 主机地址
//   - Port: 端口号
//   - User: 用户名
//   - Password: 密码
//   - DBName: 数据库名称
//   - MaxIdleConns: 最大空闲连接数
//   - MaxOpenConns: 最大打开连接数
//   - ConnMaxLifetime: 连接最大生命周期
//   - TimeZone: 时区
//   - SslMode: SSL 模式
//   - Charset: 字符集
//   - ParseTime: 是否解析时间
type Config struct {
	Type            string
	DSN             string
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
	TimeZone        string
	SslMode         string
	Charset         string
	ParseTime       bool
}

// DSNForMySQL 生成 MySQL 数据源名称。
//
// 返回值:
//   - string: MySQL DSN 字符串
func (c *Config) DSNForMySQL() string {
	if c.DSN != "" {
		return c.DSN
	}
	dsn := c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + itoa(c.Port) + ")/" + c.DBName + "?charset=" + c.Charset + "&parseTime=" + btoa(c.ParseTime) + "&loc=" + c.TimeZone
	return dsn
}

// DSNForPostgres 生成 PostgreSQL 数据源名称。
//
// 返回值:
//   - string: PostgreSQL DSN 字符串
func (c *Config) DSNForPostgres() string {
	if c.DSN != "" {
		return c.DSN
	}
	return "host=" + c.Host + " user=" + c.User + " password=" + c.Password + " dbname=" + c.DBName + " port=" + itoa(c.Port) + " sslmode=" + c.SslMode + " TimeZone=" + c.TimeZone
}

// DSNForSQLServer 生成 SQL Server 数据源名称。
//
// 返回值:
//   - string: SQL Server DSN 字符串
func (c *Config) DSNForSQLServer() string {
	if c.DSN != "" {
		return c.DSN
	}
	return "sqlserver://" + c.User + ":" + c.Password + "@" + c.Host + ":" + itoa(c.Port) + "?database=" + c.DBName
}

// DSNForSQLite 生成 SQLite 数据源名称。
//
// 返回值:
//   - string: SQLite DSN 字符串
func (c *Config) DSNForSQLite() string {
	if c.DSN != "" {
		return c.DSN
	}
	return c.DBName
}

// DBType 定义数据库类型。
type DBType string

const (
	MySQL      DBType = "mysql"     // MySQL 数据库
	PostgreSQL DBType = "postgres"  // PostgreSQL 数据库
	SQLServer  DBType = "sqlserver" // SQL Server 数据库
	SQLite     DBType = "sqlite"    // SQLite 数据库
)

// Option 是数据库配置选项函数。
type Option func(*Config)

// WithDSN 设置自定义数据源名称。
//
// 参数:
//   - dsn: 数据源名称字符串
//
// 返回值:
//   - Option: 配置选项函数
func WithDSN(dsn string) Option {
	return func(c *Config) {
		c.DSN = dsn
	}
}

// WithDBType 设置数据库类型。
//
// 参数:
//   - t: 数据库类型
//
// 返回值:
//   - Option: 配置选项函数
func WithDBType(t DBType) Option {
	return func(c *Config) {
		c.Type = string(t)
	}
}

// WithHost 设置数据库主机地址。
//
// 参数:
//   - host: 主机地址
//
// 返回值:
//   - Option: 配置选项函数
func WithHost(host string) Option {
	return func(c *Config) {
		c.Host = host
	}
}

// WithPort 设置数据库端口号。
//
// 参数:
//   - port: 端口号
//
// 返回值:
//   - Option: 配置选项函数
func WithPort(port int) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// WithUser 设置数据库用户名。
//
// 参数:
//   - user: 用户名
//
// 返回值:
//   - Option: 配置选项函数
func WithUser(user string) Option {
	return func(c *Config) {
		c.User = user
	}
}

// WithPassword 设置数据库密码。
//
// 参数:
//   - password: 密码
//
// 返回值:
//   - Option: 配置选项函数
func WithPassword(password string) Option {
	return func(c *Config) {
		c.Password = password
	}
}

// WithDBName 设置数据库名称。
//
// 参数:
//   - dbname: 数据库名称
//
// 返回值:
//   - Option: 配置选项函数
func WithDBName(dbname string) Option {
	return func(c *Config) {
		c.DBName = dbname
	}
}

// WithMaxIdleConns 设置最大空闲连接数。
//
// 参数:
//   - n: 空闲连接数
//
// 返回值:
//   - Option: 配置选项函数
func WithMaxIdleConns(n int) Option {
	return func(c *Config) {
		c.MaxIdleConns = n
	}
}

// WithMaxOpenConns 设置最大打开连接数。
//
// 参数:
//   - n: 打开连接数
//
// 返回值:
//   - Option: 配置选项函数
func WithMaxOpenConns(n int) Option {
	return func(c *Config) {
		c.MaxOpenConns = n
	}
}

// WithConnMaxLifetime 设置连接最大生命周期。
//
// 参数:
//   - d: 生命周期时长
//
// 返回值:
//   - Option: 配置选项函数
func WithConnMaxLifetime(d time.Duration) Option {
	return func(c *Config) {
		c.ConnMaxLifetime = d
	}
}

// applyPoolConfig 将连接池配置应用到数据库连接。
func applyPoolConfig(db *gorm.DB, cfg *Config) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
}

// WithSSLMode 设置 SSL 模式。
//
// 参数:
//   - mode: SSL 模式
//
// 返回值:
//   - Option: 配置选项函数
func WithSSLMode(mode string) Option {
	return func(c *Config) {
		c.SslMode = mode
	}
}

// WithTimeZone 设置时区。
//
// 参数:
//   - tz: 时区字符串
//
// 返回值:
//   - Option: 配置选项函数
func WithTimeZone(tz string) Option {
	return func(c *Config) {
		c.TimeZone = tz
	}
}

// WithCharset 设置字符集。
//
// 参数:
//   - charset: 字符集
//
// 返回值:
//   - Option: 配置选项函数
func WithCharset(charset string) Option {
	return func(c *Config) {
		c.Charset = charset
	}
}

// WithParseTime 设置是否解析时间。
//
// 参数:
//   - parseTime: 是否解析时间
//
// 返回值:
//   - Option: 配置选项函数
func WithParseTime(parseTime bool) Option {
	return func(c *Config) {
		c.ParseTime = parseTime
	}
}

// OpenMySQL 打开 MySQL 数据库连接。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *DB: 数据库连接实例
//   - error: 连接错误
//
// 示例:
//
//	db, err := gorm.OpenMySQL(
//	    gorm.WithHost("localhost"),
//	    gorm.WithPort(3306),
//	    gorm.WithUser("gate"),
//	    gorm.WithPassword("123456"),
//	    gorm.WithDBName("gate"),
//	)
func OpenMySQL(opts ...Option) (*DB, error) {
	cfg := &Config{Type: string(MySQL)}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.DSN = cfg.DSNForMySQL()

	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	gormDB := &DB{db: db}
	applyPoolConfig(db, cfg)
	return gormDB, nil
}

// OpenPostgreSQL 打开 PostgreSQL 数据库连接。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *DB: 数据库连接实例
//   - error: 连接错误
func OpenPostgreSQL(opts ...Option) (*DB, error) {
	cfg := &Config{Type: string(PostgreSQL)}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.DSN = cfg.DSNForPostgres()

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	gormDB := &DB{db: db}
	applyPoolConfig(db, cfg)
	return gormDB, nil
}

// OpenSQLServer 打开 SQL Server 数据库连接。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *DB: 数据库连接实例
//   - error: 连接错误
func OpenSQLServer(opts ...Option) (*DB, error) {
	cfg := &Config{Type: string(SQLServer)}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.DSN = cfg.DSNForSQLServer()

	db, err := gorm.Open(sqlserver.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	gormDB := &DB{db: db}
	applyPoolConfig(db, cfg)
	return gormDB, nil
}

// OpenSQLite 打开 SQLite 数据库连接。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *DB: 数据库连接实例
//   - error: 连接错误
func OpenSQLite(opts ...Option) (*DB, error) {
	cfg := &Config{Type: string(SQLite)}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.DSN = cfg.DSNForSQLite()

	db, err := gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	gormDB := &DB{db: db}
	applyPoolConfig(db, cfg)
	return gormDB, nil
}

// DB 返回 GORM 数据库实例。
func (d *DB) DB() *gorm.DB {
	return d.db
}

// Query 执行查询并返回多行结果。
func (d *DB) Query(ctx context.Context, query string, args ...any) (data.Rows, error) {
	rows, err := d.db.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

// QueryRow 执行查询并返回单行结果。
func (d *DB) QueryRow(ctx context.Context, query string, args ...any) data.Row {
	row := d.db.WithContext(ctx).Raw(query, args...).Row()
	return &Row{row: row}
}

// Exec 执行 SQL 并返回结果。
func (d *DB) Exec(ctx context.Context, query string, args ...any) (data.Result, error) {
	tx := d.db.WithContext(ctx).Exec(query, args...)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &Result{tx: tx}, nil
}

// Begin 开始一个新事务。
func (d *DB) Begin(ctx context.Context) (data.Transaction, error) {
	tx := d.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &Transaction{tx: tx}, nil
}

// Stats 返回数据库统计信息。
func (d *DB) Stats() data.DBStats {
	sqlDB, err := d.db.DB()
	if err != nil {
		return data.DBStats{}
	}
	stats := sqlDB.Stats()
	return data.DBStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
	}
}

// Close 关闭数据库连接。
func (d *DB) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Rows 包装 sql.Rows 实现 data.Rows 接口。
type Rows struct {
	rows *sql.Rows
}

// Next 准备下一行结果供读取。
//
// 返回值:
//   - bool: 是否有下一行
func (r *Rows) Next() bool {
	return r.rows.Next()
}

// Scan 将当前行数据扫描到目标变量中。
//
// 参数:
//   - dest: 目标变量列表
//
// 返回值:
//   - error: 扫描错误
func (r *Rows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

// Close 关闭 Rows 迭代器，释放资源。
//
// 返回值:
//   - error: 关闭错误
func (r *Rows) Close() error {
	return r.rows.Close()
}

// Err 返回迭代过程中的错误。
//
// 返回值:
//   - error: 迭代错误
func (r *Rows) Err() error {
	return r.rows.Err()
}

// Row 包装 sql.Row 实现 data.Row 接口。
type Row struct {
	row *sql.Row
}

// Scan 将单行结果扫描到目标变量中。
//
// 参数:
//   - dest: 目标变量列表
//
// 返回值:
//   - error: 扫描错误
func (r *Row) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

// Result 包装 GORM 的执行结果实现 data.Result 接口。
type Result struct {
	tx *gorm.DB
}

// LastInsertId 返回插入操作生成的自增 ID。
//
// 返回值:
//   - int64: 自增 ID（GORM 不直接暴露此值，始终返回 0）
//   - error: 错误
func (r *Result) LastInsertId() (int64, error) {
	return 0, nil // GORM doesn't expose this easily
}

// RowsAffected 返回受操作影响的行数。
//
// 返回值:
//   - int64: 受影响的行数
//   - error: 错误
func (r *Result) RowsAffected() (int64, error) {
	return r.tx.RowsAffected, nil
}

// Transaction 是 GORM 事务实现了 data.Transaction 接口。
//
// 字段说明:
//   - tx: GORM 事务实例
type Transaction struct {
	tx *gorm.DB
}

// Query 在事务中执行查询。
func (t *Transaction) Query(ctx context.Context, query string, args ...any) (data.Rows, error) {
	rows, err := t.tx.WithContext(ctx).Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	return &Rows{rows: rows}, nil
}

// QueryRow 在事务中执行查询并返回单行。
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...any) data.Row {
	row := t.tx.WithContext(ctx).Raw(query, args...).Row()
	return &Row{row: row}
}

// Exec 在事务中执行 SQL。
func (t *Transaction) Exec(ctx context.Context, query string, args ...any) (data.Result, error) {
	tx := t.tx.WithContext(ctx).Exec(query, args...)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &Result{tx: tx}, nil
}

// Begin 在事务中开始嵌套事务。
func (t *Transaction) Begin(ctx context.Context) (data.Transaction, error) {
	tx := t.tx.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &Transaction{tx: tx}, nil
}

// Stats 返回事务统计信息。
func (t *Transaction) Stats() data.DBStats {
	return data.DBStats{}
}

// Close 关闭事务。
func (t *Transaction) Close() error {
	return t.tx.Rollback().Error
}

// Commit 提交事务。
func (t *Transaction) Commit() error {
	return t.tx.Commit().Error
}

// Rollback 回滚事务。
func (t *Transaction) Rollback() error {
	return t.tx.Rollback().Error
}

var _ data.Transaction = (*Transaction)(nil)

// Repository 是 GORM 泛型 Repository。
//
// 字段说明:
//   - db: GORM 数据库实例
//
// 类型参数:
//   - T: 实体类型
type Repository[T any] struct {
	db *gorm.DB
}

// NewRepository 创建新的泛型 Repository。
//
// 参数:
//   - db: GORM 数据库实例
//
// 返回值:
//   - *Repository[T]: Repository 实例
func NewRepository[T any](db *gorm.DB) *Repository[T] {
	return &Repository[T]{db: db}
}

// NewRepositoryWithTx 使用事务创建新的泛型 Repository。
//
// 参数:
//   - tx: GORM 数据库实例（事务）
//
// 返回值:
//   - *Repository[T]: Repository 实例
func NewRepositoryWithTx[T any](tx *gorm.DB) *Repository[T] {
	return &Repository[T]{db: tx}
}

// Create 创建一个新实体。
//
// 参数:
//   - entity: 实体对象指针
//
// 返回值:
//   - error: 错误
func (r *Repository[T]) Create(entity *T) error {
	return r.db.Create(entity).Error
}

// CreateBatch 批量创建实体。
//
// 参数:
//   - entities: 实体切片
//
// 返回值:
//   - error: 错误
func (r *Repository[T]) CreateBatch(entities []T) error {
	if len(entities) == 0 {
		return nil
	}
	return r.db.Create(&entities).Error
}

// Delete 根据 ID 删除实体。
//
// 参数:
//   - id: 实体 ID
//
// 返回值:
//   - error: 错误
func (r *Repository[T]) Delete(id any) error {
	var model T
	result := r.db.Where("id = ?", id).Delete(&model)
	if result.Error != nil {
		return result.Error
	}
	// 如果没有找到记录，GORM 也不会报错，但应该返回错误
	if result.RowsAffected == 0 {
		return fmt.Errorf("entity with ID %v not found", id)
	}
	return nil
}

// DeleteByCondition 根据条件删除实体。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - error: 错误
func (r *Repository[T]) DeleteByCondition(where any, args ...any) error {
	return r.db.Where(where, args...).Delete(new(T)).Error
}

// Update 更新实体。
//
// 参数:
//   - entity: 实体对象指针
//
// 返回值:
//   - error: 错误
func (r *Repository[T]) Update(entity *T) error {
	return r.db.Save(entity).Error
}

// UpdateByCondition 根据条件更新实体。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - int64: 受影响的行数
//   - error: 错误
func (r *Repository[T]) UpdateByCondition(where any, args ...any) (int64, error) {
	result := r.db.Where(where, args...).Updates(new(T))
	return result.RowsAffected, result.Error
}

// FindByID 根据 ID 查询实体。
//
// 参数:
//   - id: 实体 ID
//
// 返回值:
//   - *T: 实体指针
//   - error: 错误
func (r *Repository[T]) FindByID(id any) (*T, error) {
	var result T
	err := r.db.First(&result, id).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// FindOne 根据条件查询单个实体。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - *T: 实体指针
//   - error: 错误
func (r *Repository[T]) FindOne(where any, args ...any) (*T, error) {
	var result T
	err := r.db.Where(where, args...).First(&result).Error
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// FindAll 根据条件查询所有实体。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - []T: 实体切片
//   - error: 错误
func (r *Repository[T]) FindAll(where any, args ...any) ([]T, error) {
	var results []T
	err := r.db.Where(where, args...).Find(&results).Error
	return results, err
}

// Count 统计符合条件 的实体数量。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - int64: 数量
//   - error: 错误
func (r *Repository[T]) Count(where any, args ...any) (int64, error) {
	var count int64
	err := r.db.Model(new(T)).Where(where, args...).Count(&count).Error
	return count, err
}

// Raw 执行原生 SQL 查询。
//
// 参数:
//   - sql: SQL 语句
//   - args: 查询参数
//
// 返回值:
//   - []T: 结果切片
//   - error: 错误
func (r *Repository[T]) Raw(sql string, args ...any) ([]T, error) {
	var results []T
	err := r.db.Raw(sql, args...).Find(&results).Error
	return results, err
}

// Client 是 GORM 数据库客户端。
//
// 字段说明:
//   - db: GORM 数据库实例
type Client struct {
	db *gorm.DB
}

// NewClient 创建新的数据库客户端。
//
// 参数:
//   - db: GORM 数据库实例
//
// 返回值:
//   - *Client: 客户端实例
func NewClient(db *gorm.DB) *Client {
	return &Client{db: db}
}

// DB 返回 GORM 数据库实例。
func (c *Client) DB() *gorm.DB {
	return c.db
}

// Begin 开始一个新事务。
//
// 返回值:
//   - data.Transaction: 事务实例
//   - error: 错误
func (c *Client) Begin() (data.Transaction, error) {
	tx := c.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &Transaction{tx: tx}, nil
}

// BeginTx 在事务中执行回调函数。
//
// 参数:
//   - ctx: 上下文
//   - db: GORM 数据库实例
//   - fc: 回调函数
//
// 返回值:
//   - error: 错误
func BeginTx(ctx context.Context, db *gorm.DB, fc func(tx *gorm.DB) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fc(tx)
	})
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func btoa(b bool) string {
	return strconv.FormatBool(b)
}
