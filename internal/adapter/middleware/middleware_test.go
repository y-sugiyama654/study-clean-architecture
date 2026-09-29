// internal/adapter/middleware/middleware_test.go
package middleware_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/middleware"
	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// stubVerifier は "good-token" だけを alice のトークンとして受け付ける
type stubVerifier struct{}

func (stubVerifier) Verify(_ context.Context, token string) (domain.UserID, error) {
	if token == "good-token" {
		return "alice", nil
	}
	return "", middleware.ErrInvalidToken
}

func TestAuthenticate(t *testing.T) {
	var gotUser domain.UserID
	h := middleware.Authenticate(stubVerifier{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser = middleware.UserID(r.Context())
	}))

	tests := []struct {
		name, header string
		wantStatus   int
		wantUser     domain.UserID
	}{
		{"正しいトークン", "Bearer good-token", http.StatusOK, "alice"},
		{"ヘッダなし", "", http.StatusUnauthorized, ""},
		{"Bearerでない", "Basic good-token", http.StatusUnauthorized, ""},
		{"違うトークン", "Bearer bad-token", http.StatusUnauthorized, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser = ""
			req := httptest.NewRequest("GET", "/tasks", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus || gotUser != tt.wantUser {
				t.Errorf("want (%d, %q), got (%d, %q): %s", tt.wantStatus, tt.wantUser, rec.Code, gotUser, rec.Body)
			}
		})
	}
}

func TestRequestID(t *testing.T) {
	var inCtx string
	h := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inCtx = middleware.RequestIDFrom(r.Context())
	}))

	// 付いていなければ新しく振る
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if inCtx == "" || rec.Header().Get("X-Request-ID") != inCtx {
		t.Errorf("リクエストIDが振られていない: ctx=%q header=%q", inCtx, rec.Header().Get("X-Request-ID"))
	}

	// 付いていれば引き継ぐ
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "from-upstream")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if inCtx != "from-upstream" {
		t.Errorf("呼び出し元のリクエストIDを引き継いでいない: %q", inCtx)
	}
}

func TestAccessLogAndRecover(t *testing.T) {
	var buf bytes.Buffer
	lg := slog.New(slog.NewJSONHandler(&buf, nil))
	h := middleware.Recover(lg)(middleware.AccessLog(lg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/panic" {
			panic("バグ")
		}
		w.WriteHeader(http.StatusTeapot)
	})))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/teapot", nil))
	if !strings.Contains(buf.String(), `"status":418`) {
		t.Errorf("アクセスログにステータスが出ていない: %s", buf.String())
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/panic", nil))
	if rec.Code != http.StatusInternalServerError || !strings.Contains(buf.String(), "パニックが発生しました") {
		t.Errorf("パニックが500とログになっていない: %d %s", rec.Code, buf.String())
	}
}
