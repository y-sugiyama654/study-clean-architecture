// internal/domain/status.go
package domain

// Status はタスクの状態
type Status string

const (
	StatusTodo Status = "todo" // 未完了
	StatusDone Status = "done" // 完了
)

// ParseStatus は文字列からStatusを作る。保存されていた値を読み込むときに使う
func ParseStatus(s string) (Status, error) {
	switch Status(s) {
	case StatusTodo, StatusDone:
		return Status(s), nil
	}
	return "", ErrInvalidStatus
}

func (s Status) String() string { return string(s) }
