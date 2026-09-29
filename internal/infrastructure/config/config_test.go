// internal/infrastructure/config/config_test.go
package config_test

import (
	"log/slog"
	"testing"

	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/config"
)

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://app@localhost:5432/tasks")
	t.Setenv("API_TOKENS", "test-token-a:alice, test-token-b:bob")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("HTTP_ADDR の既定値: want :8080, got %s", cfg.HTTPAddr)
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LOG_LEVEL: want DEBUG, got %s", cfg.LogLevel)
	}
	if len(cfg.APITokens) != 2 || cfg.APITokens["test-token-b"] != "bob" {
		t.Errorf("API_TOKENS の読み込みが不正: %v", cfg.APITokens)
	}
}

func TestLoad_Storage(t *testing.T) {
	// メモリに保存する場合は、DATABASE_URL はなくてよい
	t.Setenv("STORAGE", "memory")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("API_TOKENS", "test-token-a:alice")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Storage != "memory" {
		t.Errorf("want memory, got %s", cfg.Storage)
	}

	t.Setenv("STORAGE", "mysql")
	if _, err := config.Load(); err == nil {
		t.Error("未知の STORAGE がエラーにならなかった")
	}
}

func TestLoad_Errors(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := config.Load(); err == nil {
		t.Error("DATABASE_URL がないのにエラーにならなかった")
	}

	t.Setenv("DATABASE_URL", "postgres://app@localhost:5432/tasks")
	t.Setenv("API_TOKENS", "no-colon")
	if _, err := config.Load(); err == nil {
		t.Error("不正な API_TOKENS がエラーにならなかった")
	}

	t.Setenv("API_TOKENS", "test-token-a:alice")
	t.Setenv("LOG_LEVEL", "verbose")
	if _, err := config.Load(); err == nil {
		t.Error("不正な LOG_LEVEL がエラーにならなかった")
	}
}
