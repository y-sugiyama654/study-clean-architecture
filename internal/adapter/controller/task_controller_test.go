// internal/adapter/controller/task_controller_test.go
package controller_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/controller"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/memory"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/middleware"
	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2025, 8, 1, 9, 0, 0, 0, time.UTC) }

// seqIDs は task-1, task-2, ... と順にIDを払い出す
type seqIDs struct{ n int }

func (g *seqIDs) NewTaskID() domain.TaskID {
	g.n++
	return domain.TaskID(fmt.Sprintf("task-%d", g.n))
}

func newController() *controller.TaskController {
	store := memory.NewStore()
	tx := memory.NewTransactor(store)
	repo := memory.NewTaskRepository(store)
	activities := memory.NewActivityRepository(store)
	return controller.NewTaskController(
		usecase.NewCreateTask(tx, repo, activities, &seqIDs{}, fixedClock{}),
		usecase.NewGetTask(repo),
		usecase.NewListTasks(repo),
		usecase.NewCompleteTask(tx, repo, activities, fixedClock{}),
	)
}

// do はハンドラを直接呼び出して、レスポンスを返す
func do(h http.HandlerFunc, method, path, userID, body string, pathID string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if userID != "" {
		// 本来は認証ミドルウェアが入れる、認証済みのユーザーIDを context に入れる
		req = req.WithContext(middleware.WithUserID(req.Context(), domain.UserID(userID)))
	}
	if pathID != "" {
		req.SetPathValue("id", pathID) // ルーターを通さないので、パスパラメータを自分で設定する
	}
	rec := httptest.NewRecorder()
	h(rec, req)
	return rec
}

func TestTaskController_CreateGetComplete(t *testing.T) {
	c := newController()

	rec := do(c.Create, "POST", "/tasks", "user-1", `{"title":"牛乳を買う","description":"低脂肪のもの"}`, "")
	if rec.Code != http.StatusCreated {
		t.Fatalf("作成: want 201, got %d: %s", rec.Code, rec.Body)
	}
	var created map[string]any
	json.Unmarshal(rec.Body.Bytes(), &created)
	if created["id"] != "task-1" || created["title"] != "牛乳を買う" || created["status"] != "todo" {
		t.Errorf("作成のレスポンスが不正: %s", rec.Body)
	}

	rec = do(c.Get, "GET", "/tasks/task-1", "user-1", "", "task-1")
	if rec.Code != http.StatusOK {
		t.Fatalf("取得: want 200, got %d: %s", rec.Code, rec.Body)
	}

	rec = do(c.Complete, "POST", "/tasks/task-1/complete", "user-1", "", "task-1")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"done"`) {
		t.Fatalf("完了: want 200 & done, got %d: %s", rec.Code, rec.Body)
	}

	rec = do(c.List, "GET", "/tasks", "user-1", "", "")
	var list []map[string]any
	json.Unmarshal(rec.Body.Bytes(), &list)
	if rec.Code != http.StatusOK || len(list) != 1 {
		t.Errorf("一覧: want 200 & 1件, got %d: %s", rec.Code, rec.Body)
	}
}

func TestTaskController_ErrorStatus(t *testing.T) {
	c := newController()
	do(c.Create, "POST", "/tasks", "user-1", `{"title":"牛乳を買う"}`, "")
	do(c.Complete, "POST", "/tasks/task-1/complete", "user-1", "", "task-1")

	tests := []struct {
		name string
		rec  *httptest.ResponseRecorder
		want int
	}{
		{"JSONが壊れている", do(c.Create, "POST", "/tasks", "user-1", `{"title":`, ""), http.StatusBadRequest},
		{"タイトルが空", do(c.Create, "POST", "/tasks", "user-1", `{"title":""}`, ""), http.StatusBadRequest},
		{"存在しないタスク", do(c.Get, "GET", "/tasks/nope", "user-1", "", "nope"), http.StatusNotFound},
		{"他人のタスク", do(c.Get, "GET", "/tasks/task-1", "user-2", "", "task-1"), http.StatusNotFound},
		{"完了済みのタスクを完了", do(c.Complete, "POST", "/tasks/task-1/complete", "user-1", "", "task-1"), http.StatusConflict},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.rec.Code != tt.want {
				t.Errorf("want %d, got %d: %s", tt.want, tt.rec.Code, tt.rec.Body)
			}
			t.Logf("%d %s", tt.rec.Code, strings.TrimSpace(tt.rec.Body.String()))
		})
	}
}
