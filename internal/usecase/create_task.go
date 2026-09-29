// internal/usecase/create_task.go
package usecase

import (
	"context"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// CreateTaskInput はタスク作成の入力
type CreateTaskInput struct {
	UserID      string // 操作しているユーザー（タスクの所有者になる）
	Title       string
	Description string
}

// CreateTask はタスクを作成するユースケース
type CreateTask struct {
	tx         Transactor
	repo       TaskRepository
	activities ActivityRepository
	ids        IDGenerator
	clock      Clock
}

func NewCreateTask(tx Transactor, repo TaskRepository, activities ActivityRepository, ids IDGenerator, clock Clock) *CreateTask {
	return &CreateTask{tx: tx, repo: repo, activities: activities, ids: ids, clock: clock}
}

func (uc *CreateTask) Execute(ctx context.Context, in CreateTaskInput) (TaskOutput, error) {
	ownerID, err := domain.NewUserID(in.UserID)
	if err != nil {
		return TaskOutput{}, err
	}
	title, err := domain.NewTitle(in.Title)
	if err != nil {
		return TaskOutput{}, err
	}
	desc, err := domain.NewDescription(in.Description)
	if err != nil {
		return TaskOutput{}, err
	}

	now := uc.clock.Now()
	task, err := domain.NewTask(uc.ids.NewTaskID(), ownerID, title, desc, now)
	if err != nil {
		return TaskOutput{}, err
	}

	// タスクの保存と履歴の記録は、両方成功するか、両方なかったことになるかのどちらか
	err = uc.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := uc.repo.Save(ctx, task); err != nil {
			return err
		}
		return uc.activities.Add(ctx, domain.Activity{
			TaskID: task.ID(), ActorID: ownerID, Action: domain.ActivityCreated, OccurredAt: now,
		})
	})
	if err != nil {
		return TaskOutput{}, err
	}
	return newTaskOutput(task), nil
}
