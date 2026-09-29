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
	repo  TaskRepository
	ids   IDGenerator
	clock Clock
}

func NewCreateTask(repo TaskRepository, ids IDGenerator, clock Clock) *CreateTask {
	return &CreateTask{repo: repo, ids: ids, clock: clock}
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

	task, err := domain.NewTask(uc.ids.NewTaskID(), ownerID, title, desc, uc.clock.Now())
	if err != nil {
		return TaskOutput{}, err
	}
	if err := uc.repo.Save(ctx, task); err != nil {
		return TaskOutput{}, err
	}
	return newTaskOutput(task), nil
}
