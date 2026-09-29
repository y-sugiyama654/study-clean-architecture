// internal/domain/activity.go
package domain

import "time"

// ActivityAction はタスクに対して行われた操作の種類
type ActivityAction string

const (
	ActivityCreated   ActivityAction = "created"   // 作成
	ActivityCompleted ActivityAction = "completed" // 完了
)

// Activity はタスクに対する操作の履歴（誰が・いつ・何をしたか）
// 一度記録したら変わらない記録なので、フィールドを公開した単純な値として扱う
type Activity struct {
	TaskID     TaskID
	ActorID    UserID
	Action     ActivityAction
	OccurredAt time.Time
}
