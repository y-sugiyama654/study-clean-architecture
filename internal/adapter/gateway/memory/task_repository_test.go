// internal/adapter/gateway/memory/task_repository_test.go
package memory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/memory"
	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func TestTaskRepository(t *testing.T) {
	ctx := context.Background()
	repo := memory.NewTaskRepository()
	now := time.Date(2025, 8, 1, 9, 0, 0, 0, time.UTC)
	title, _ := domain.NewTitle("牛乳を買う")
	task, _ := domain.NewTask("task-1", "user-1", title, domain.Description{}, now)

	if err := repo.Save(ctx, task); err != nil {
		t.Fatal(err)
	}

	// 取り出したタスクを書き換えても、Saveするまでは保存内容は変わらない
	found, err := repo.FindByID(ctx, "task-1")
	if err != nil {
		t.Fatal(err)
	}
	found.Complete(now)
	again, _ := repo.FindByID(ctx, "task-1")
	if again.Status() != domain.StatusTodo {
		t.Error("Saveしていない変更が保存されてしまった")
	}

	if _, err := repo.FindByID(ctx, "nope"); !errors.Is(err, usecase.ErrTaskNotFound) {
		t.Errorf("want ErrTaskNotFound, got %v", err)
	}
	list, _ := repo.ListByOwner(ctx, "user-1")
	if len(list) != 1 {
		t.Errorf("want 1件, got %d件", len(list))
	}
}
