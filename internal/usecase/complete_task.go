// internal/usecase/complete_task.go
package usecase

import "context"

// CompleteTaskInput はタスク完了の入力
type CompleteTaskInput struct {
	UserID string
	TaskID string
}

// CompleteTask は自分のタスクを完了にするユースケース
type CompleteTask struct {
	repo  TaskRepository
	clock Clock
}

func NewCompleteTask(repo TaskRepository, clock Clock) *CompleteTask {
	return &CompleteTask{repo: repo, clock: clock}
}

func (uc *CompleteTask) Execute(ctx context.Context, in CompleteTaskInput) (TaskOutput, error) {
	task, err := findOwnTask(ctx, uc.repo, in.UserID, in.TaskID)
	if err != nil {
		return TaskOutput{}, err
	}
	// 「完了済みなら完了にできない」というルールはエンティティが知っている
	if err := task.Complete(uc.clock.Now()); err != nil {
		return TaskOutput{}, err
	}
	if err := uc.repo.Save(ctx, task); err != nil {
		return TaskOutput{}, err
	}
	return newTaskOutput(task), nil
}
