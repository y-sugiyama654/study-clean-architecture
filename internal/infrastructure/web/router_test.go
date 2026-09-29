// internal/infrastructure/web/router_test.go
package web_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/controller"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/memory"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/auth"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/system"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/web"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func TestRouter(t *testing.T) {
	store := memory.NewStore()
	tx := memory.NewTransactor(store)
	repo := memory.NewTaskRepository(store)
	activities := memory.NewActivityRepository(store)
	ctrl := controller.NewTaskController(
		usecase.NewCreateTask(tx, repo, activities, system.UUIDGenerator{}, system.Clock{}),
		usecase.NewGetTask(repo),
		usecase.NewListTasks(repo),
		usecase.NewCompleteTask(tx, repo, activities, system.Clock{}),
	)
	verifier := auth.NewStaticTokenVerifier(map[string]string{"test-token-1": "user-1"})
	srv := httptest.NewServer(web.NewRouter(ctrl, verifier, slog.New(slog.DiscardHandler)))
	defer srv.Close()

	tests := []struct {
		method, path, token string
		want                int
	}{
		{"POST", "/tasks", "test-token-1", http.StatusCreated},
		{"GET", "/tasks", "test-token-1", http.StatusOK},
		{"GET", "/tasks/no-such-task", "test-token-1", http.StatusNotFound},
		{"DELETE", "/tasks", "test-token-1", http.StatusMethodNotAllowed}, // メソッドが違えば405
		{"GET", "/users", "test-token-1", http.StatusNotFound},            // パスがなければ404
		{"GET", "/tasks", "", http.StatusUnauthorized},                    // トークンがなければ401
		{"GET", "/tasks", "wrong-token", http.StatusUnauthorized},         // トークンが違えば401
	}
	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, srv.URL+tt.path, strings.NewReader(`{"title":"牛乳を買う"}`))
		if tt.token != "" {
			req.Header.Set("Authorization", "Bearer "+tt.token)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != tt.want {
			t.Errorf("%s %s (token=%q): want %d, got %d", tt.method, tt.path, tt.token, tt.want, res.StatusCode)
		}
		if res.Header.Get("X-Request-ID") == "" {
			t.Errorf("%s %s: X-Request-ID ヘッダがない", tt.method, tt.path)
		}
	}
}
