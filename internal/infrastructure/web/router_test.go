// internal/infrastructure/web/router_test.go
package web_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/controller"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/memory"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/system"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/web"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func TestRouter(t *testing.T) {
	repo := memory.NewTaskRepository()
	ctrl := controller.NewTaskController(
		usecase.NewCreateTask(repo, system.UUIDGenerator{}, system.Clock{}),
		usecase.NewGetTask(repo),
		usecase.NewListTasks(repo),
		usecase.NewCompleteTask(repo, system.Clock{}),
	)
	srv := httptest.NewServer(web.NewRouter(ctrl))
	defer srv.Close()

	tests := []struct {
		method, path string
		want         int
	}{
		{"POST", "/tasks", http.StatusCreated},
		{"GET", "/tasks", http.StatusOK},
		{"GET", "/tasks/no-such-task", http.StatusNotFound},
		{"DELETE", "/tasks", http.StatusMethodNotAllowed}, // メソッドが違えば405
		{"GET", "/users", http.StatusNotFound},            // パスがなければ404
	}
	for _, tt := range tests {
		req, _ := http.NewRequest(tt.method, srv.URL+tt.path, strings.NewReader(`{"title":"牛乳を買う"}`))
		req.Header.Set("X-User-ID", "user-1")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != tt.want {
			t.Errorf("%s %s: want %d, got %d", tt.method, tt.path, tt.want, res.StatusCode)
		}
	}
}
