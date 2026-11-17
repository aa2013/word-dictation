package util

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type quietGormLogger struct {
	gormlogger.Interface
}

// Trace 屏蔽 ErrRecordNotFound 的错误日志（预期内分支）
func (l quietGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	l.Interface.Trace(ctx, begin, fc, err)
}

func GormLogger() gormlogger.Interface {
	return quietGormLogger{gormlogger.Default.LogMode(gormlogger.Warn)}
}
