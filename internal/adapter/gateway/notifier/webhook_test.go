// internal/adapter/gateway/notifier/webhook_test.go
package notifier_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/notifier"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func TestWebhook(t *testing.T) {
	var got map[string]any
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(status)
	}))
	defer srv.Close()

	var logs bytes.Buffer
	n := notifier.NewWebhook(srv.URL, srv.Client(), slog.New(slog.NewJSONHandler(&logs, nil)))
	completedAt := time.Date(2025, 8, 1, 10, 0, 0, 0, time.UTC)
	task := usecase.TaskOutput{ID: "task-1", OwnerID: "alice", Title: "牛乳を買う", Status: "done", CompletedAt: &completedAt}

	n.TaskCompleted(context.Background(), task)
	if got["event"] != "task.completed" || got["task_id"] != "task-1" || got["title"] != "牛乳を買う" {
		t.Errorf("送られた内容が不正: %v", got)
	}

	// 通知先がエラーを返しても、呼び出し側には何も返さず、ログに残す
	status = http.StatusInternalServerError
	n.TaskCompleted(context.Background(), task)
	if !strings.Contains(logs.String(), "通知先がエラーを返しました") {
		t.Errorf("失敗がログに残っていない: %s", logs.String())
	}
}
