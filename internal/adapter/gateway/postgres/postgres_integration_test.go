//go:build integration

// internal/adapter/gateway/postgres/postgres_integration_test.go
//
// 実際のPostgreSQL（Dockerのコンテナ）を使うテスト。次のコマンドで実行する:
//
//	go test -tags=integration ./...
package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/notifier"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/postgres"
	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/system"
	"github.com/y-sugiyama654/study-clean-architecture/internal/testutil/postgrestest"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func TestPostgres(t *testing.T) {
	db, _ := postgrestest.Start(t)
	ctx := context.Background()
	tx := postgres.NewTransactor(db)
	repo := postgres.NewTaskRepository(db)
	activities := postgres.NewActivityRepository(db)
	now := time.Date(2025, 8, 1, 9, 0, 0, 0, time.UTC)

	newTask := func(id domain.TaskID, owner domain.UserID, createdAt time.Time) *domain.Task {
		title, _ := domain.NewTitle("タスク " + id.String())
		desc, _ := domain.NewDescription("説明")
		task, err := domain.NewTask(id, owner, title, desc, createdAt)
		if err != nil {
			t.Fatal(err)
		}
		return task
	}

	t.Run("保存して読み出す", func(t *testing.T) {
		task := newTask("repo-1", "alice", now)
		if err := repo.Save(ctx, task); err != nil {
			t.Fatal(err)
		}
		got, err := repo.FindByID(ctx, "repo-1")
		if err != nil {
			t.Fatal(err)
		}
		if got.Title() != task.Title() || got.Description() != task.Description() ||
			got.Status() != domain.StatusTodo || !got.CreatedAt().Equal(now) || got.CompletedAt() != nil {
			t.Errorf("読み出した内容が保存した内容と違う: %+v", got)
		}

		// 完了にして保存し直すと、更新される
		got.Complete(now.Add(time.Hour))
		if err := repo.Save(ctx, got); err != nil {
			t.Fatal(err)
		}
		again, _ := repo.FindByID(ctx, "repo-1")
		if again.Status() != domain.StatusDone || !again.CompletedAt().Equal(now.Add(time.Hour)) {
			t.Errorf("更新が反映されていない: %+v", again)
		}
	})

	t.Run("存在しないタスク", func(t *testing.T) {
		if _, err := repo.FindByID(ctx, "no-such-task"); !errors.Is(err, usecase.ErrTaskNotFound) {
			t.Errorf("want ErrTaskNotFound, got %v", err)
		}
	})

	t.Run("所有者のタスクを作成順に返す", func(t *testing.T) {
		repo.Save(ctx, newTask("list-2", "bob", now.Add(time.Minute)))
		repo.Save(ctx, newTask("list-1", "bob", now))
		repo.Save(ctx, newTask("list-3", "carol", now))
		tasks, err := repo.ListByOwner(ctx, "bob")
		if err != nil {
			t.Fatal(err)
		}
		if len(tasks) != 2 || tasks[0].ID() != "list-1" || tasks[1].ID() != "list-2" {
			t.Errorf("一覧が不正: %d件", len(tasks))
		}
	})

	t.Run("エラーならロールバックする", func(t *testing.T) {
		errBoom := errors.New("途中で失敗")
		err := tx.WithinTx(ctx, func(ctx context.Context) error {
			if err := repo.Save(ctx, newTask("tx-1", "alice", now)); err != nil {
				return err
			}
			return errBoom
		})
		if !errors.Is(err, errBoom) {
			t.Fatalf("want errBoom, got %v", err)
		}
		if _, err := repo.FindByID(ctx, "tx-1"); !errors.Is(err, usecase.ErrTaskNotFound) {
			t.Errorf("ロールバックしたタスクが残っている: %v", err)
		}
	})

	t.Run("更新するつもりで読むと、他のトランザクションはコミットまで待たされる", func(t *testing.T) {
		if err := repo.Save(ctx, newTask("lock-1", "alice", now)); err != nil {
			t.Fatal(err)
		}

		// トランザクション1: ロックを取って読み、200ミリ秒後に完了にしてコミットする
		locked := make(chan struct{})
		tx1Done := make(chan error, 1)
		go func() {
			tx1Done <- tx.WithinTx(ctx, func(ctx context.Context) error {
				task, err := repo.FindByIDForUpdate(ctx, "lock-1")
				if err != nil {
					return err
				}
				close(locked)
				time.Sleep(200 * time.Millisecond)
				task.Complete(now)
				return repo.Save(ctx, task)
			})
		}()
		<-locked

		// トランザクション2: 同じタスクを更新するつもりで読もうとすると、トランザクション1のコミットまで待たされ、
		// 完了済みの状態を読む
		start := time.Now()
		var status domain.Status
		err := tx.WithinTx(ctx, func(ctx context.Context) error {
			task, err := repo.FindByIDForUpdate(ctx, "lock-1")
			if err != nil {
				return err
			}
			status = task.Status()
			return nil
		})
		waited := time.Since(start)
		if err != nil {
			t.Fatal(err)
		}
		if err := <-tx1Done; err != nil {
			t.Fatal(err)
		}
		if status != domain.StatusDone || waited < 150*time.Millisecond {
			t.Errorf("ロックで待たされていない: 読んだ状態 %s、待った時間 %v", status, waited)
		}
	})

	t.Run("同時に完了にしても成功するのは1回だけ", func(t *testing.T) {
		create := usecase.NewCreateTask(tx, repo, activities, system.UUIDGenerator{}, system.Clock{})
		complete := usecase.NewCompleteTask(tx, repo, activities, notifier.Nop{}, system.Clock{})
		out, err := create.Execute(ctx, usecase.CreateTaskInput{UserID: "dave", Title: "牛乳を買う"})
		if err != nil {
			t.Fatal(err)
		}

		var wg sync.WaitGroup
		errs := make(chan error, 10)
		for range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := complete.Execute(ctx, usecase.CompleteTaskInput{UserID: "dave", TaskID: out.ID})
				errs <- err
			}()
		}
		wg.Wait()
		close(errs)

		succeeded := 0
		for err := range errs {
			if err == nil {
				succeeded++
			} else if !errors.Is(err, domain.ErrTaskAlreadyCompleted) {
				t.Errorf("想定外のエラー: %v", err)
			}
		}
		var completedCount int
		db.QueryRowContext(ctx,
			`SELECT count(*) FROM task_activities WHERE task_id = $1 AND action = 'completed'`, out.ID).Scan(&completedCount)
		if succeeded != 1 || completedCount != 1 {
			t.Errorf("成功 %d回・完了の履歴 %d件（どちらも1のはず）", succeeded, completedCount)
		}
	})
}
