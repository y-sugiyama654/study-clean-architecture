// internal/usecase/complete_task.go
package usecase

import (
	"context"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// CompleteTaskInput はタスク完了の入力
type CompleteTaskInput struct {
	UserID string
	TaskID string
}

// CompleteTask は自分のタスクを完了にするユースケース
type CompleteTask struct {
	tx         Transactor
	repo       TaskRepository
	activities ActivityRepository
	notifier   Notifier
	clock      Clock
}

func NewCompleteTask(tx Transactor, repo TaskRepository, activities ActivityRepository, notifier Notifier, clock Clock) *CompleteTask {
	return &CompleteTask{tx: tx, repo: repo, activities: activities, notifier: notifier, clock: clock}
}

func (uc *CompleteTask) Execute(ctx context.Context, in CompleteTaskInput) (TaskOutput, error) {
	var task *domain.Task
	err := uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		// 読んでから保存するまでの間に、他の処理が同じタスクを完了にしないよう、ロックを取って読む
		task, err = findOwnTask(ctx, uc.repo.FindByIDForUpdate, in.UserID, in.TaskID)
		if err != nil {
			return err
		}
		now := uc.clock.Now()
		// 「完了済みなら完了にできない」というルールはエンティティが知っている
		if err := task.Complete(now); err != nil {
			return err
		}
		if err := uc.repo.Save(ctx, task); err != nil {
			return err
		}
		return uc.activities.Add(ctx, domain.Activity{
			TaskID: task.ID(), ActorID: task.OwnerID(), Action: domain.ActivityCompleted, OccurredAt: now,
		})
	})
	if err != nil {
		return TaskOutput{}, err
	}

	// 通知はトランザクションがコミットされた後に送る
	// （トランザクションの中で送ると、送った後でロールバックされたときに、起きていないことを通知してしまう）
	out := newTaskOutput(task)
	uc.notifier.TaskCompleted(ctx, out)
	return out, nil
}
