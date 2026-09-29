// internal/domain/task.go
package domain

import "time"

// Task はタスクを表すエンティティ
// 状態の変更は必ずメソッドを通して行い、不変条件（ルール）が崩れないようにする
type Task struct {
	id          TaskID
	ownerID     UserID
	title       Title
	description Description
	status      Status
	createdAt   time.Time
	completedAt *time.Time // 完了していなければnil
}

// NewTask は新しい未完了のタスクを作る
// 現在時刻は引数で受け取る。エンティティの中で time.Now() を呼ばないことで、テストで時刻を固定できる
func NewTask(id TaskID, ownerID UserID, title Title, description Description, now time.Time) (*Task, error) {
	if id == "" {
		return nil, ErrTaskIDRequired
	}
	if ownerID == "" {
		return nil, ErrUserIDRequired
	}
	if title.String() == "" {
		return nil, ErrTitleRequired // ゼロ値のTitleを渡された場合
	}
	return &Task{
		id:          id,
		ownerID:     ownerID,
		title:       title,
		description: description,
		status:      StatusTodo,
		createdAt:   now,
	}, nil
}

// ReconstructTask は保存されていた値からタスクを復元する（リポジトリから読み込むときに使う）
// 保存されていた値は一度ルールを通ったものなので、ここでは検証しない
func ReconstructTask(id TaskID, ownerID UserID, title Title, description Description,
	status Status, createdAt time.Time, completedAt *time.Time) *Task {
	return &Task{
		id:          id,
		ownerID:     ownerID,
		title:       title,
		description: description,
		status:      status,
		createdAt:   createdAt,
		completedAt: completedAt,
	}
}

func (t *Task) ID() TaskID                 { return t.id }
func (t *Task) OwnerID() UserID            { return t.ownerID }
func (t *Task) Title() Title               { return t.title }
func (t *Task) Description() Description   { return t.description }
func (t *Task) Status() Status             { return t.status }
func (t *Task) CreatedAt() time.Time       { return t.createdAt }
func (t *Task) CompletedAt() *time.Time    { return t.completedAt }
func (t *Task) IsOwnedBy(user UserID) bool { return t.ownerID == user }

// Complete はタスクを完了にする。完了済みのタスクはもう一度完了にはできない
func (t *Task) Complete(now time.Time) error {
	if t.status == StatusDone {
		return ErrTaskAlreadyCompleted
	}
	t.status = StatusDone
	t.completedAt = &now
	return nil
}
