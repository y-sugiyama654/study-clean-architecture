// internal/usecase/usecase_test.go
package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func TestCreateTask(t *testing.T) {
	repo := newFakeTaskRepository()
	activities := &fakeActivityRepository{}
	tx := &fakeTransactor{}
	uc := usecase.NewCreateTask(tx, repo, activities, fixedIDGenerator{id: "task-1"}, fixedClock{now: testNow})

	out, err := uc.Execute(context.Background(), usecase.CreateTaskInput{
		UserID: "user-1", Title: "  牛乳を買う ", Description: "低脂肪のもの",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := usecase.TaskOutput{
		ID: "task-1", OwnerID: "user-1", Title: "牛乳を買う", Description: "低脂肪のもの",
		Status: "todo", CreatedAt: testNow,
	}
	if out != want {
		t.Errorf("出力が不正:\nwant %+v\ngot  %+v", want, out)
	}
	if _, ok := repo.tasks["task-1"]; !ok {
		t.Error("タスクが保存されていない")
	}
	wantActivity := domain.Activity{TaskID: "task-1", ActorID: "user-1", Action: domain.ActivityCreated, OccurredAt: testNow}
	if len(activities.activities) != 1 || activities.activities[0] != wantActivity {
		t.Errorf("作成の履歴が記録されていない: %+v", activities.activities)
	}
	if tx.calls != 1 {
		t.Errorf("保存はトランザクションの中で行うはず: WithinTxの呼び出し %d回", tx.calls)
	}
}

func TestCreateTask_Errors(t *testing.T) {
	tests := []struct {
		name    string
		input   usecase.CreateTaskInput
		saveErr error
		addErr  error
		wantErr error
	}{
		{name: "タイトルが空", input: usecase.CreateTaskInput{UserID: "user-1", Title: ""}, wantErr: domain.ErrTitleRequired},
		{name: "ユーザーIDが空", input: usecase.CreateTaskInput{UserID: "", Title: "牛乳を買う"}, wantErr: domain.ErrUserIDRequired},
		{name: "保存に失敗", input: usecase.CreateTaskInput{UserID: "user-1", Title: "牛乳を買う"}, saveErr: errDB, wantErr: errDB},
		{name: "履歴の記録に失敗", input: usecase.CreateTaskInput{UserID: "user-1", Title: "牛乳を買う"}, addErr: errDB, wantErr: errDB},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newFakeTaskRepository()
			repo.saveErr = tt.saveErr
			activities := &fakeActivityRepository{addErr: tt.addErr}
			uc := usecase.NewCreateTask(&fakeTransactor{}, repo, activities, fixedIDGenerator{id: "task-1"}, fixedClock{now: testNow})

			_, err := uc.Execute(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("want %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestGetTask(t *testing.T) {
	repo := newFakeTaskRepository(newTask("task-1", "user-1", testNow))
	uc := usecase.NewGetTask(repo)

	out, err := uc.Execute(context.Background(), usecase.GetTaskInput{UserID: "user-1", TaskID: "task-1"})
	if err != nil {
		t.Fatal(err)
	}
	if out.ID != "task-1" {
		t.Errorf("want task-1, got %s", out.ID)
	}

	// 存在しないタスクも、他人のタスクも「見つからない」になる
	for _, in := range []usecase.GetTaskInput{
		{UserID: "user-1", TaskID: "no-such-task"},
		{UserID: "user-2", TaskID: "task-1"},
	} {
		if _, err := uc.Execute(context.Background(), in); !errors.Is(err, usecase.ErrTaskNotFound) {
			t.Errorf("%+v: want ErrTaskNotFound, got %v", in, err)
		}
	}
}

func TestListTasks(t *testing.T) {
	repo := newFakeTaskRepository(
		newTask("task-2", "user-1", testNow.Add(time.Hour)),
		newTask("task-1", "user-1", testNow),
		newTask("task-3", "user-2", testNow),
	)
	uc := usecase.NewListTasks(repo)

	out, err := uc.Execute(context.Background(), usecase.ListTasksInput{UserID: "user-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[0].ID != "task-1" || out[1].ID != "task-2" {
		t.Errorf("自分のタスクが作成順に返るはず: got %+v", out)
	}
}

func TestCompleteTask(t *testing.T) {
	repo := newFakeTaskRepository(newTask("task-1", "user-1", testNow))
	activities := &fakeActivityRepository{}
	completedAt := testNow.Add(2 * time.Hour)
	uc := usecase.NewCompleteTask(&fakeTransactor{}, repo, activities, fixedClock{now: completedAt})

	out, err := uc.Execute(context.Background(), usecase.CompleteTaskInput{UserID: "user-1", TaskID: "task-1"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "done" || out.CompletedAt == nil || !out.CompletedAt.Equal(completedAt) {
		t.Errorf("完了になっていない: %+v", out)
	}
	if repo.tasks["task-1"].Status() != domain.StatusDone {
		t.Error("完了した状態が保存されていない")
	}
	if len(activities.activities) != 1 || activities.activities[0].Action != domain.ActivityCompleted {
		t.Errorf("完了の履歴が記録されていない: %+v", activities.activities)
	}

	// 2回目は「完了済み」のエラー
	_, err = uc.Execute(context.Background(), usecase.CompleteTaskInput{UserID: "user-1", TaskID: "task-1"})
	if !errors.Is(err, domain.ErrTaskAlreadyCompleted) {
		t.Errorf("want ErrTaskAlreadyCompleted, got %v", err)
	}
}

func TestCompleteTask_OtherUsersTask(t *testing.T) {
	repo := newFakeTaskRepository(newTask("task-1", "user-1", testNow))
	uc := usecase.NewCompleteTask(&fakeTransactor{}, repo, &fakeActivityRepository{}, fixedClock{now: testNow})

	_, err := uc.Execute(context.Background(), usecase.CompleteTaskInput{UserID: "user-2", TaskID: "task-1"})
	if !errors.Is(err, usecase.ErrTaskNotFound) {
		t.Errorf("want ErrTaskNotFound, got %v", err)
	}
	if repo.tasks["task-1"].Status() != domain.StatusTodo {
		t.Error("他人のタスクが完了になってしまった")
	}
}
