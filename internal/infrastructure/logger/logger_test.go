// internal/infrastructure/logger/logger_test.go
package logger_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/middleware"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/logger"
)

func TestLogger_AddsRequestID(t *testing.T) {
	var buf bytes.Buffer
	lg := logger.New(&buf, slog.LevelInfo)

	h := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lg.InfoContext(r.Context(), "処理中")
	}))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "req-123")
	h.ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(buf.String(), `"request_id":"req-123"`) {
		t.Errorf("ログにリクエストIDが付いていない: %s", buf.String())
	}
}
