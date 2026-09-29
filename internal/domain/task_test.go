// internal/domain/task_test.go
package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

var now = time.Date(2025, 8, 1, 9, 0, 0, 0, time.UTC)

func newTestTask(t *testing.T) *domain.Task {
	t.Helper()
	title, _ := domain.NewTitle("牛乳を買う")
	desc, _ := domain.NewDescription("低脂肪のもの")
	task, err := domain.NewTask("task-1", "user-1", title, desc, now)
	if err != nil {
		t.Fatal(err)
	}
	return task
}

func TestNewTask(t *testing.T) {
	task := newTestTask(t)

	if task.Status() != domain.StatusTodo {
		t.Errorf("新しいタスクは未完了のはず: got %s", task.Status())
	}
	if !task.CreatedAt().Equal(now) {
		t.Errorf("作成日時は引数の時刻になるはず: got %v", task.CreatedAt())
	}
	if task.CompletedAt() != nil {
		t.Errorf("未完了のタスクに完了日時はないはず: got %v", task.CompletedAt())
	}
	if !task.IsOwnedBy("user-1") || task.IsOwnedBy("user-2") {
		t.Error("所有者の判定が不正")
	}
}

func TestNewTask_Invalid(t *testing.T) {
	title, _ := domain.NewTitle("牛乳を買う")
	tests := []struct {
		name    string
		id      domain.TaskID
		owner   domain.UserID
		title   domain.Title
		wantErr error
	}{
		{name: "IDが空", id: "", owner: "user-1", title: title, wantErr: domain.ErrTaskIDRequired},
		{name: "所有者が空", id: "task-1", owner: "", title: title, wantErr: domain.ErrUserIDRequired},
		{name: "タイトルがゼロ値", id: "task-1", owner: "user-1", title: domain.Title{}, wantErr: domain.ErrTitleRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewTask(tt.id, tt.owner, tt.title, domain.Description{}, now)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestTask_Complete(t *testing.T) {
	task := newTestTask(t)
	completedAt := now.Add(time.Hour)

	if err := task.Complete(completedAt); err != nil {
		t.Fatal(err)
	}
	if task.Status() != domain.StatusDone {
		t.Errorf("完了後の状態: want done, got %s", task.Status())
	}
	if task.CompletedAt() == nil || !task.CompletedAt().Equal(completedAt) {
		t.Errorf("完了日時が記録されていない: got %v", task.CompletedAt())
	}

	// 完了済みのタスクはもう一度完了にはできない
	if err := task.Complete(completedAt.Add(time.Hour)); !errors.Is(err, domain.ErrTaskAlreadyCompleted) {
		t.Errorf("2回目の完了はエラーになるはず: got %v", err)
	}
	if !task.CompletedAt().Equal(completedAt) {
		t.Error("2回目の完了で完了日時が変わってしまった")
	}
}
