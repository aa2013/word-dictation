package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// LocalTime 统一 "yyyy-MM-dd HH:mm:ss" 文本格式，与种子数据（原 MySQL dump）的 datetime 兼容。
type LocalTime struct {
	time.Time
}

const timeLayout = "2006-01-02 15:04:05"

func NewLocalTime(t time.Time) LocalTime { return LocalTime{Time: t} }

func Now() LocalTime { return LocalTime{Time: time.Now()} }

func ParseLocalTime(s string) (LocalTime, error) {
	t, err := time.ParseInLocation(timeLayout, s, time.Local)
	return LocalTime{Time: t}, err
}

func (t LocalTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Format(timeLayout), nil
}

func (t *LocalTime) Scan(v any) error {
	if v == nil {
		t.Time = time.Time{}
		return nil
	}
	switch val := v.(type) {
	case []byte:
		return t.parse(string(val))
	case string:
		return t.parse(val)
	case time.Time:
		t.Time = val
		return nil
	default:
		return fmt.Errorf("无法将 %T 解析为 LocalTime", v)
	}
}

func (t *LocalTime) parse(s string) error {
	for _, layout := range []string{timeLayout, "2006-01-02T15:04:05Z07:00", "2006-01-02"} {
		if parsed, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			t.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("无法解析时间: %s", s)
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.Format(timeLayout) + `"`), nil
}

func (t *LocalTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		t.Time = time.Time{}
		return nil
	}
	if len(s) >= 2 && s[0] == '"' {
		s = s[1 : len(s)-1]
	}
	return t.parse(s)
}
