// Package DBMigration 提供基于版本号的自增迁移能力。
//
// 迁移文件放在本包 migrations/ 子目录，文件名为 {序号}_{名称}.sql，例如 0001_baseline.sql、0002_add_xxx.sql。
// 已应用版本记录在数据库 schema_migrations 表；启动时自动把未应用的迁移按序号从小到大执行。
// 新增数据库变更时，添加新的序号文件即可，不要修改已提交的旧文件。
package DBMigration

import (
	"embed"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var migrationFilePattern = regexp.MustCompile(`^(\d+)_(.+)\.sql$`)

// schemaMigration 对应 schema_migrations 表。
type schemaMigration struct {
	Version   int64
	Name      string
	AppliedAt string
}

func (schemaMigration) TableName() string {
	return "schema_migrations"
}

// MigrationError 迁移失败时包装错误信息（只包含版本号，不包含 SQL 内容，避免日志泄漏表结构细节）。
type MigrationError struct {
	Version int64
	Name    string
	Err     error
}

func (e *MigrationError) Error() string {
	return fmt.Sprintf("database migration %d (%s) failed: %v", e.Version, e.Name, e.Err)
}

func (e *MigrationError) Unwrap() error {
	return e.Err
}

// Migration 表示一个已解析的迁移文件。
type Migration struct {
	Version int64
	Name    string
	SQL     string
}

// loadMigrations 从 embedded FS 读取并按版本号升序返回所有迁移。
func loadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("读取迁移目录失败: %w", err)
	}
	migrations := make([]Migration, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		match := migrationFilePattern.FindStringSubmatch(entry.Name())
		if match == nil {
			return nil, fmt.Errorf("非法迁移文件名 %q，应为 {序号}_{名称}.sql", entry.Name())
		}
		version, err := strconv.ParseInt(match[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("解析迁移版本号 %q 失败: %w", match[1], err)
		}
		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("读取迁移文件 %q 失败: %w", entry.Name(), err)
		}
		migrations = append(migrations, Migration{
			Version: version,
			Name:    match[2],
			SQL:     string(content),
		})
	}
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})
	return migrations, nil
}

// appliedVersions 返回数据库已应用的迁移版本集合。
func appliedVersions(db *gorm.DB) (map[int64]bool, error) {
	var rows []schemaMigration
	if err := db.Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("读取 schema_migrations 失败: %w", err)
	}
	versions := make(map[int64]bool, len(rows))
	for _, row := range rows {
		versions[row.Version] = true
	}
	return versions, nil
}

// Migrate 把数据库迁移到最新版本：
//
//   - 自动创建 schema_migrations 版本表（幂等）。
//   - 对比 embedded 迁移文件和已应用版本，找出未应用的迁移。
//   - 在事务中逐个执行迁移 SQL 并写入版本记录；单个迁移失败会回滚该迁移并返回错误，已成功的迁移不受影响。
//
// 迁移文件使用纯 SQL，支持多语句；不要在迁移中手动操作 schema_migrations 表。
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&schemaMigration{}); err != nil {
		return fmt.Errorf("创建 schema_migrations 表失败: %w", err)
	}
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}
	applied, err := appliedVersions(db)
	if err != nil {
		return err
	}
	for _, migration := range migrations {
		if applied[migration.Version] {
			continue
		}
		if err := applyOne(db, migration); err != nil {
			return err
		}
	}
	return nil
}

func applyOne(db *gorm.DB, migration Migration) error {
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(migration.SQL).Error; err != nil {
			return err
		}
		record := schemaMigration{Version: migration.Version, Name: migration.Name}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return &MigrationError{Version: migration.Version, Name: migration.Name, Err: err}
	}
	fmt.Printf("[automigration] applied migration %d_%s\n", migration.Version, migration.Name)
	return nil
}
