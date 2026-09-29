// internal/adapter/middleware/observability.go
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/presenter"
)

type requestIDKey struct{}

// RequestIDFrom はリクエストIDを context から取り出す（なければ空文字）
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// RequestID はリクエストごとにIDを振り、context とレスポンスヘッダに入れる
// 呼び出し元が X-Request-ID を付けてきた場合はそれを引き継ぐ（サービスをまたいで処理を追えるようにする）
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			b := make([]byte, 8)
			rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

// statusRecorder は、書き込まれたステータスコードを覚えておく ResponseWriter
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(status int) {
	rec.status = status
	rec.ResponseWriter.WriteHeader(status)
}

// AccessLog はリクエストごとに、メソッド・パス・ステータスコード・処理時間をログに出す
func AccessLog(lg *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			lg.InfoContext(r.Context(), "HTTPリクエスト",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// Recover はハンドラ内で起きたパニックを捕まえてログに出し、500を返す
// これがないと、1つのリクエストのバグでサーバー全体が落ちてしまう
func Recover(lg *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					lg.ErrorContext(r.Context(), "パニックが発生しました", slog.Any("panic", p))
					presenter.JSON(w, http.StatusInternalServerError, presenter.ErrorResponse{
						Error: presenter.ErrorBody{Code: "internal", Message: presenter.InternalErrorMessage},
					})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
