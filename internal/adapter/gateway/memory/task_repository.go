// internal/adapter/gateway/memory/task_repository.go
package memory

import (
	"context"
	"sort"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// TaskRepository はタスクをメモリ上に保存する usecase.TaskRepository の実装
// プロセスを終了すると内容は消える。開発時の動作確認やデモに使う
type TaskRepository struct {
	s *Store
}

var _ usecase.TaskRepository = (*TaskRepository)(nil)

func NewTaskRepository(s *Store) *TaskRepository {
	return &TaskRepository{s: s}
}

func (r *TaskRepository) Save(_ context.Context, task *domain.Task) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.tasks[task.ID()] = clone(task)
	return nil
}

func (r *TaskRepository) FindByID(_ context.Context, id domain.TaskID) (*domain.Task, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	t, ok := r.s.tasks[id]
	if !ok {
		return nil, usecase.ErrTaskNotFound
	}
	return clone(t), nil
}

// FindByIDForUpdate は FindByID と同じ。メモリ実装では Transactor がトランザクションを
// 1つずつ実行するので、別途ロックを取る必要がない
func (r *TaskRepository) FindByIDForUpdate(ctx context.Context, id domain.TaskID) (*domain.Task, error) {
	return r.FindByID(ctx, id)
}

func (r *TaskRepository) ListByOwner(_ context.Context, ownerID domain.UserID) ([]*domain.Task, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	result := []*domain.Task{}
	for _, t := range r.s.tasks {
		if t.IsOwnedBy(ownerID) {
			result = append(result, clone(t))
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt().Before(result[j].CreatedAt()) })
	return result, nil
}

// clone はタスクのコピーを作る
// 保存したものと同じポインタを返すと、Saveを呼ばずに書き換えた内容まで「保存」されてしまうので、
// データベースと同じように、保存と取得のたびに別のオブジェクトにする
func clone(t *domain.Task) *domain.Task {
	var completedAt = t.CompletedAt()
	if completedAt != nil {
		c := *completedAt
		completedAt = &c
	}
	return domain.ReconstructTask(t.ID(), t.OwnerID(), t.Title(), t.Description(),
		t.Status(), t.CreatedAt(), completedAt)
}
