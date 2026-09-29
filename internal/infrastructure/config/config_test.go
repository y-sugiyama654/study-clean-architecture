// internal/infrastructure/config/config_test.go
package config_test

import (
	"log/slog"
	"testing"

	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/config"
)

func TestLoad(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://app@localhost:5432/tasks")
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
}

func TestLoad_Errors(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := config.Load(); err == nil {
		t.Error("DATABASE_URL がないのにエラーにならなかった")
	}

	t.Setenv("DATABASE_URL", "postgres://app@localhost:5432/tasks")
	t.Setenv("LOG_LEVEL", "verbose")
	if _, err := config.Load(); err == nil {
		t.Error("不正な LOG_LEVEL がエラーにならなかった")
	}
}
