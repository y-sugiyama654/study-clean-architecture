// internal/infrastructure/logger/logger.go
package logger

import (
	"io"
	"log/slog"
)

// New はJSON形式で構造化ログを出力するロガーを作る
func New(w io.Writer, level slog.Level) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level}))
}
