// internal/infrastructure/config/config.go
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
)

// Config はアプリケーションの設定
type Config struct {
	HTTPAddr    string     // HTTPサーバーが待ち受けるアドレス
	DatabaseURL string     // PostgreSQLの接続URL
	LogLevel    slog.Level // 出力するログの最低レベル
}

// Load は環境変数から設定を読み込む
// 接続先などの環境ごとに変わる値や、パスワードを含む値はコードに書かず、環境変数で渡す
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:    getenv("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		LogLevel:    slog.LevelInfo,
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("config: 環境変数 DATABASE_URL を設定してください")
	}
	if s := os.Getenv("LOG_LEVEL"); s != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(s)); err != nil {
			return Config{}, fmt.Errorf("config: LOG_LEVEL が不正です: %w", err)
		}
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
