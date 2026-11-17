package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Email struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Sender   string `yaml:"sender"`
	Debug    bool   `yaml:"debug"`
}

type JWT struct {
	Secret       string `yaml:"secret"`
	TokenDays    int    `yaml:"tokenDays"`
	RememberDays int    `yaml:"rememberDays"`
}

type Storage struct {
	Dir string `yaml:"dir"`
}

type DB struct {
	Path string `yaml:"path"`
}

type Config struct {
	Port    int     `yaml:"port"`
	DB      DB      `yaml:"db"`
	Storage Storage `yaml:"storage"`
	JWT     JWT     `yaml:"jwt"`
	Email   Email   `yaml:"email"`
}

func defaultConfig() *Config {
	return &Config{
		Port:    8099,
		DB:      DB{Path: "data/word-dictation.db"},
		Storage: Storage{Dir: "data/storage"},
		JWT: JWT{
			Secret:       randomSecret(),
			TokenDays:    7,
			RememberDays: 30,
		},
		Email: Email{
			Host:  "smtp.qq.com",
			Port:  587,
			Debug: true,
		},
	}
}

func randomSecret() string {
	buf := make([]byte, 48)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := Save(path, defaultConfig()); err != nil {
			return nil, fmt.Errorf("写入默认配置失败: %w", err)
		}
		fmt.Println("[config] 未找到配置文件，已生成默认配置:", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := defaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	if cfg.JWT.Secret == "" {
		return nil, fmt.Errorf("配置 jwt.secret 不能为空")
	}
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func dirOf(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[:i]
		}
	}
	return "."
}
