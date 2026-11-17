package migration

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开数据库失败: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
	})
	return db
}

func count(t *testing.T, db *gorm.DB, table string) int64 {
	t.Helper()
	var n int64
	if err := db.Raw("SELECT COUNT(*) FROM " + table).Scan(&n).Error; err != nil {
		t.Fatalf("统计 %s 失败: %v", table, err)
	}
	return n
}

func TestSplitStatements(t *testing.T) {
	script := `
-- 建表
CREATE TABLE a (x TEXT);  -- 行尾注释
INSERT INTO a VALUES (1, 'it''s ok; fine');
INSERT INTO a VALUES (2, "quoted; ident");
`
	got := SplitStatements(script)
	if len(got) != 3 {
		t.Fatalf("期望 3 条语句, 得到 %d: %v", len(got), got)
	}
	if got[0] != "CREATE TABLE a (x TEXT)" {
		t.Fatalf("第一条语句错误: %q", got[0])
	}
	if got[1] != "INSERT INTO a VALUES (1, 'it''s ok; fine')" {
		t.Fatalf("带转义引号的语句被错误拆分: %q", got[1])
	}
}

func TestRunSeedsData(t *testing.T) {
	db := openTestDB(t)
	if err := Run(db); err != nil {
		t.Fatalf("迁移失败: %v", err)
	}
	cases := map[string]int64{
		"lib":                 9,
		"word":                787,
		"word_explain":        7391,
		"lib_word":            1602,
		"plan":                13,
		"plan_word":           425,
		"user":                4,
		"user_lib":            19,
		"word_explain_custom": 0,
	}
	for table, want := range cases {
		if got := count(t, db, table); got != want {
			t.Errorf("表 %s 行数 = %d, 期望 %d", table, got, want)
		}
	}
	var applied int64
	db.Raw("SELECT COUNT(*) FROM schema_migrations").Scan(&applied)
	if applied != 2 {
		t.Errorf("迁移记录 = %d, 期望 2", applied)
	}

	// 重复执行应跳过已应用迁移
	if err := Run(db); err != nil {
		t.Fatalf("重复迁移失败: %v", err)
	}
	if got := count(t, db, "lib"); got != 9 {
		t.Errorf("重复迁移后 lib 行数 = %d, 期望 9", got)
	}

	// 转义数据完整性：o'clock 等含转义引号的行应被正确导入
	var wordText string
	if err := db.Raw("SELECT word FROM word WHERE id = 226").Scan(&wordText).Error; err != nil {
		t.Fatal(err)
	}
	if wordText != "o'clock" {
		t.Errorf("转义引号导入错误: %q", wordText)
	}
}
