// internal/usecase/fakes_test.go
package usecase_test

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// fakeTaskRepository はテスト用に、タスクをメモリ上のマップに保存するリポジトリ
type fakeTaskRepository struct {
	tasks   map[domain.TaskID]*domain.Task
	saveErr error // Saveでわざと失敗させたいときに設定する
}

func newFakeTaskRepository(tasks ...*domain.Task) *fakeTaskRepository {
	r := &fakeTaskRepository{tasks: map[domain.TaskID]*domain.Task{}}
	for _, t := range tasks {
		r.tasks[t.ID()] = t
	}
	return r
}

func (r *fakeTaskRepository) Save(_ context.Context, task *domain.Task) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	r.tasks[task.ID()] = task
	return nil
}

func (r *fakeTaskRepository) FindByID(_ context.Context, id domain.TaskID) (*domain.Task, error) {
	t, ok := r.tasks[id]
	if !ok {
		return nil, usecase.ErrTaskNotFound
	}
	return t, nil
}

func (r *fakeTaskRepository) FindByIDForUpdate(ctx context.Context, id domain.TaskID) (*domain.Task, error) {
	return r.FindByID(ctx, id)
}

func (r *fakeTaskRepository) ListByOwner(_ context.Context, ownerID domain.UserID) ([]*domain.Task, error) {
	var result []*domain.Task
	for _, t := range r.tasks {
		if t.IsOwnedBy(ownerID) {
			result = append(result, t)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt().Before(result[j].CreatedAt()) })
	return result, nil
}

// fakeActivityRepository は記録された履歴をスライスに溜める
type fakeActivityRepository struct {
	activities []domain.Activity
	addErr     error
}

func (r *fakeActivityRepository) Add(_ context.Context, a domain.Activity) error {
	if r.addErr != nil {
		return r.addErr
	}
	r.activities = append(r.activities, a)
	return nil
}

// fakeTransactor はトランザクションを張らずに fn を呼ぶだけ
// （ロールバックの動きは、Transactorの実装側のテストで確かめる）
type fakeTransactor struct{ calls int }

func (t *fakeTransactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	t.calls++
	return fn(ctx)
}

// fixedClock は常に同じ時刻を返す
type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

// fixedIDGenerator は常に同じIDを返す
type fixedIDGenerator struct{ id domain.TaskID }

func (g fixedIDGenerator) NewTaskID() domain.TaskID { return g.id }

var (
	testNow = time.Date(2025, 8, 1, 9, 0, 0, 0, time.UTC)
	errDB   = errors.New("データベースに接続できません")
)

// newTask はテスト用のタスクを作る
func newTask(id domain.TaskID, owner domain.UserID, createdAt time.Time) *domain.Task {
	title, _ := domain.NewTitle("タスク " + id.String())
	task, err := domain.NewTask(id, owner, title, domain.Description{}, createdAt)
	if err != nil {
		panic(err)
	}
	return task
}
