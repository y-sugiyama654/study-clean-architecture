// internal/infrastructure/logger/logger.go
package logger

import (
	"context"
	"io"
	"log/slog"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/middleware"
)

// New はJSON形式で構造化ログを出力するロガーを作る
// context を渡してログを出すと（InfoContext など）、リクエストIDが自動で付く
func New(w io.Writer, level slog.Level) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	return slog.New(contextHandler{h})
}

// contextHandler は、context に入っているリクエストIDをログに付け足す slog.Handler
type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := middleware.RequestIDFrom(ctx); id != "" {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}
