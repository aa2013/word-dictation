// Package migration SQLite 版本化迁移器。
// 迁移文件为 migrations/ 下按 000N_名称.sql 命名的 SQL 脚本（go:embed 嵌入），
// 启动时按版本号顺序执行未应用的脚本，记录保存在 schema_migrations 表。
package migration

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type migrationFile struct {
	version    int
	name       string
	statements []string
}

var fileNameRe = regexp.MustCompile(`^(\d+)_(.+)\.sql$`)

func Run(db *gorm.DB) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TEXT NOT NULL
	)`).Error; err != nil {
		return fmt.Errorf("创建迁移记录表失败: %w", err)
	}

	files, err := loadFiles()
	if err != nil {
		return err
	}

	var current int
	if err := db.Raw(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&current).Error; err != nil {
		return err
	}

	for _, f := range files {
		if f.version <= current {
			continue
		}
		start := time.Now()
		err := db.Transaction(func(tx *gorm.DB) error {
			for _, stmt := range f.statements {
				if err := tx.Exec(stmt).Error; err != nil {
					return fmt.Errorf("执行 %s 失败: %w", f.name, err)
				}
			}
			return tx.Exec(`INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)`,
				f.version, f.name, time.Now().Format("2006-01-02 15:04:05")).Error
		})
		if err != nil {
			return err
		}
		log.Printf("[migration] %03d_%s 已应用 (%d 条语句, %dms)", f.version, f.name, len(f.statements), time.Since(start).Milliseconds())
	}
	return nil
}

func loadFiles() ([]migrationFile, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, err
	}
	var files []migrationFile
	for _, e := range entries {
		m := fileNameRe.FindStringSubmatch(e.Name())
		if m == nil {
			return nil, fmt.Errorf("迁移文件命名不合法: %s（应为 000N_名称.sql）", e.Name())
		}
		version, _ := strconv.Atoi(m[1])
		data, err := migrationFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, err
		}
		files = append(files, migrationFile{
			version:    version,
			name:       m[2],
			statements: SplitStatements(string(data)),
		})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].version < files[j].version })
	for i := 1; i < len(files); i++ {
		if files[i].version == files[i-1].version {
			return nil, fmt.Errorf("迁移版本号重复: %03d", files[i].version)
		}
	}
	return files, nil
}

// SplitStatements 引号感知的语句拆分，支持 '' 转义、双引号标识符与 -- 行注释。
func SplitStatements(script string) []string {
	var statements []string
	var sb strings.Builder

	flush := func() {
		s := strings.TrimSpace(sb.String())
		if s != "" {
			statements = append(statements, s)
		}
		sb.Reset()
	}

	inSingle, inDouble := false, false
	runes := []rune(script)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		switch {
		case inSingle:
			sb.WriteRune(ch)
			if ch == '\'' {
				if i+1 < len(runes) && runes[i+1] == '\'' {
					sb.WriteRune('\'')
					i++
				} else {
					inSingle = false
				}
			}
		case inDouble:
			sb.WriteRune(ch)
			if ch == '"' {
				inDouble = false
			}
		case ch == '-' && i+1 < len(runes) && runes[i+1] == '-':
			for ; i < len(runes) && runes[i] != '\n'; i++ {
			}
			if i < len(runes) {
				sb.WriteRune('\n')
			}
		case ch == '\'':
			inSingle = true
			sb.WriteRune(ch)
		case ch == '"':
			inDouble = true
			sb.WriteRune(ch)
		case ch == ';':
			flush()
		default:
			sb.WriteRune(ch)
		}
	}
	flush()
	return statements
}
