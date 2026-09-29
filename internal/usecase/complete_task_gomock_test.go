// internal/usecase/complete_task_gomock_test.go
package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase/mock"
)

// gomock で生成したモックを使うと、「どのメソッドが、どんな引数で、何回呼ばれるか」を検証できる

// runInTx は、WithinTx が呼ばれたら fn をそのまま実行するようにモックを設定する
func runInTx(tx *mock.MockTransactor) {
	tx.EXPECT().WithinTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })
}

func TestCompleteTask_WithGomock(t *testing.T) {
	ctrl := gomock.NewController(t) // テストの終わりに、期待した呼び出しがすべて行われたかを確かめる
	tx := mock.NewMockTransactor(ctrl)
	repo := mock.NewMockTaskRepository(ctrl)
	activities := mock.NewMockActivityRepository(ctrl)
	clock := mock.NewMockClock(ctrl)

	completedAt := testNow.Add(time.Hour)
	runInTx(tx)
	clock.EXPECT().Now().Return(completedAt)
	repo.EXPECT().FindByIDForUpdate(gomock.Any(), domain.TaskID("task-1")).
		Return(newTask("task-1", "user-1", testNow), nil)
	// 完了状態になったタスクが保存されること
	repo.EXPECT().Save(gomock.Any(), gomock.Cond(func(t *domain.Task) bool {
		return t.Status() == domain.StatusDone
	})).Return(nil)
	// 完了の履歴が、完了した時刻で記録されること
	activities.EXPECT().Add(gomock.Any(), domain.Activity{
		TaskID: "task-1", ActorID: "user-1", Action: domain.ActivityCompleted, OccurredAt: completedAt,
	}).Return(nil)

	uc := usecase.NewCompleteTask(tx, repo, activities, clock)
	if _, err := uc.Execute(context.Background(), usecase.CompleteTaskInput{UserID: "user-1", TaskID: "task-1"}); err != nil {
		t.Fatal(err)
	}
}

func TestCompleteTask_WithGomock_AlreadyCompleted(t *testing.T) {
	ctrl := gomock.NewController(t)
	tx := mock.NewMockTransactor(ctrl)
	repo := mock.NewMockTaskRepository(ctrl)
	activities := mock.NewMockActivityRepository(ctrl)
	clock := mock.NewMockClock(ctrl)

	done := newTask("task-1", "user-1", testNow)
	done.Complete(testNow)

	runInTx(tx)
	clock.EXPECT().Now().Return(testNow.Add(time.Hour))
	repo.EXPECT().FindByIDForUpdate(gomock.Any(), domain.TaskID("task-1")).Return(done, nil)
	// Save と Add は EXPECT していないので、呼ばれたらテストが失敗する

	uc := usecase.NewCompleteTask(tx, repo, activities, clock)
	_, err := uc.Execute(context.Background(), usecase.CompleteTaskInput{UserID: "user-1", TaskID: "task-1"})
	if !errors.Is(err, domain.ErrTaskAlreadyCompleted) {
		t.Errorf("want ErrTaskAlreadyCompleted, got %v", err)
	}
}
