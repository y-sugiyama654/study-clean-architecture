// internal/usecase/get_task.go
package usecase

import (
	"context"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// GetTaskInput はタスク取得の入力
type GetTaskInput struct {
	UserID string
	TaskID string
}

// GetTask は自分のタスクを1件取得するユースケース
type GetTask struct {
	repo TaskRepository
}

func NewGetTask(repo TaskRepository) *GetTask {
	return &GetTask{repo: repo}
}

func (uc *GetTask) Execute(ctx context.Context, in GetTaskInput) (TaskOutput, error) {
	task, err := findOwnTask(ctx, uc.repo.FindByID, in.UserID, in.TaskID)
	if err != nil {
		return TaskOutput{}, err
	}
	return newTaskOutput(task), nil
}

// findOwnTask は、findで探したタスクのうち、ユーザーが所有するものだけを返す
// 他人のタスクは「存在しない」ものとして扱う（存在すること自体を知られないようにする）
func findOwnTask(ctx context.Context, find func(context.Context, domain.TaskID) (*domain.Task, error),
	userID, taskID string) (*domain.Task, error) {
	uid, err := domain.NewUserID(userID)
	if err != nil {
		return nil, err
	}
	tid, err := domain.NewTaskID(taskID)
	if err != nil {
		return nil, err
	}
	task, err := find(ctx, tid)
	if err != nil {
		return nil, err
	}
	if !task.IsOwnedBy(uid) {
		return nil, ErrTaskNotFound
	}
	return task, nil
}
