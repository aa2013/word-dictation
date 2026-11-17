// Package storage 封面图片等本地文件存储。
package storage

import (
	_ "embed"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"worddictation/server/global"
	"worddictation/server/util"
)

//go:embed default_cover.jpg
var defaultCover []byte

// EnsureDefaultCover 启动时把内置默认封面落到存储目录，供 /covers/default.jpg 访问
func EnsureDefaultCover() error {
	path := filepath.Join(global.Config.Storage.Dir, "default.jpg")
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return os.WriteFile(path, defaultCover, 0o644)
}

func UploadCover(libId int, b64Data string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(util.StripBase64Prefix(b64Data))
	if err != nil {
		return "", fmt.Errorf("封面 base64 解码失败: %w", err)
	}
	rel := fmt.Sprintf("wordDictation/cover%d.jpg", libId)
	abs := filepath.Join(global.Config.Storage.Dir, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		return "", err
	}
	return "/covers/" + rel, nil
}
