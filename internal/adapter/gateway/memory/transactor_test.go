// internal/adapter/gateway/memory/transactor_test.go
package memory_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/memory"
	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/system"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func TestTransactor_RollbackOnError(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	tx := memory.NewTransactor(store)
	repo := memory.NewTaskRepository(store)
	activities := memory.NewActivityRepository(store)

	title, _ := domain.NewTitle("牛乳を買う")
	task, _ := domain.NewTask("task-1", "user-1", title, domain.Description{}, time.Now())
	errBoom := errors.New("途中で失敗")

	err := tx.WithinTx(ctx, func(ctx context.Context) error {
		repo.Save(ctx, task)
		activities.Add(ctx, domain.Activity{TaskID: "task-1", ActorID: "user-1", Action: domain.ActivityCreated})
		return errBoom // 保存した後で失敗する
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("want errBoom, got %v", err)
	}

	// どちらの変更も取り消されている
	if _, err := repo.FindByID(ctx, "task-1"); !errors.Is(err, usecase.ErrTaskNotFound) {
		t.Errorf("ロールバックされたタスクが残っている: %v", err)
	}
	if n := len(store.Activities()); n != 0 {
		t.Errorf("ロールバックされた履歴が残っている: %d件", n)
	}
}

func TestCompleteTask_Concurrent(t *testing.T) {
	ctx := context.Background()
	store := memory.NewStore()
	tx := memory.NewTransactor(store)
	repo := memory.NewTaskRepository(store)
	activities := memory.NewActivityRepository(store)
	create := usecase.NewCreateTask(tx, repo, activities, system.UUIDGenerator{}, system.Clock{})
	complete := usecase.NewCompleteTask(tx, repo, activities, system.Clock{})

	out, err := create.Execute(ctx, usecase.CreateTaskInput{UserID: "user-1", Title: "牛乳を買う"})
	if err != nil {
		t.Fatal(err)
	}

	// 同じタスクを10個のゴルーチンから同時に完了にしようとする
	var wg sync.WaitGroup
	errs := make(chan error, 10)
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := complete.Execute(ctx, usecase.CompleteTaskInput{UserID: "user-1", TaskID: out.ID})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)

	succeeded := 0
	for err := range errs {
		switch {
		case err == nil:
			succeeded++
		case !errors.Is(err, domain.ErrTaskAlreadyCompleted):
			t.Errorf("想定外のエラー: %v", err)
		}
	}
	if succeeded != 1 {
		t.Errorf("完了に成功するのは1回だけのはず: %d回", succeeded)
	}
	completed := 0
	for _, a := range store.Activities() {
		if a.Action == domain.ActivityCompleted {
			completed++
		}
	}
	if completed != 1 {
		t.Errorf("完了の履歴は1件だけのはず: %d件", completed)
	}
}
