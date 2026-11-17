package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"worddictation/server/config"
	"worddictation/server/global"
	"worddictation/server/migration"
	"worddictation/server/router"
	"worddictation/server/util"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	global.Config = cfg

	if err := os.MkdirAll(filepath.Dir(cfg.DB.Path), 0o755); err != nil {
		log.Fatalf("创建数据目录失败: %v", err)
	}
	if err := os.MkdirAll(cfg.Storage.Dir, 0o755); err != nil {
		log.Fatalf("创建存储目录失败: %v", err)
	}

	dsn := cfg.DB.Path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: util.GormLogger(),
	})
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	global.DB = db

	if err := migration.Run(db); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	r := router.Setup()
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("word-dictation 已启动: http://127.0.0.1:%d", cfg.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
