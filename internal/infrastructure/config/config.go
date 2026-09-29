// internal/infrastructure/config/config.go
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Config はアプリケーションの設定
type Config struct {
	HTTPAddr         string            // HTTPサーバーが待ち受けるアドレス
	Storage          string            // データの保存先: "postgres" または "memory"
	DatabaseURL      string            // PostgreSQLの接続URL（Storage が "postgres" のとき）
	LogLevel         slog.Level        // 出力するログの最低レベル
	APITokens        map[string]string // APIトークン → ユーザーID
	NotifyWebhookURL string            // タスクの完了を知らせるWebhookのURL（空なら通知しない）
}

// Load は環境変数から設定を読み込む
// 接続先などの環境ごとに変わる値や、パスワードを含む値はコードに書かず、環境変数で渡す
func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:         getenv("HTTP_ADDR", ":8080"),
		Storage:          getenv("STORAGE", "postgres"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		LogLevel:         slog.LevelInfo,
		NotifyWebhookURL: os.Getenv("NOTIFY_WEBHOOK_URL"),
	}
	switch cfg.Storage {
	case "postgres":
		if cfg.DatabaseURL == "" {
			return Config{}, errors.New("config: 環境変数 DATABASE_URL を設定してください")
		}
	case "memory":
	default:
		return Config{}, fmt.Errorf("config: STORAGE は postgres か memory を指定してください: %q", cfg.Storage)
	}
	tokens, err := parseTokens(os.Getenv("API_TOKENS"))
	if err != nil {
		return Config{}, err
	}
	cfg.APITokens = tokens
	if s := os.Getenv("LOG_LEVEL"); s != "" {
		if err := cfg.LogLevel.UnmarshalText([]byte(s)); err != nil {
			return Config{}, fmt.Errorf("config: LOG_LEVEL が不正です: %w", err)
		}
	}
	return cfg, nil
}

// parseTokens は「トークン:ユーザーID」をカンマでつないだ文字列を読む（例: "t0ken-a:alice,t0ken-b:bob"）
// トークンは秘密の値なので、コードやリポジトリには書かず、環境変数で渡す
func parseTokens(s string) (map[string]string, error) {
	if s == "" {
		return nil, errors.New("config: 環境変数 API_TOKENS を設定してください")
	}
	tokens := map[string]string{}
	for _, pair := range strings.Split(s, ",") {
		token, user, ok := strings.Cut(strings.TrimSpace(pair), ":")
		if !ok || token == "" || user == "" {
			return nil, errors.New("config: API_TOKENS は「トークン:ユーザーID」をカンマでつないだ形式で指定してください")
		}
		tokens[token] = user
	}
	return tokens, nil
}

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
