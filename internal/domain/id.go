// internal/domain/id.go
package domain

// TaskID はタスクを識別するID
type TaskID string

// NewTaskID は文字列からTaskIDを作る。空文字は許さない
func NewTaskID(s string) (TaskID, error) {
	if s == "" {
		return "", ErrTaskIDRequired
	}
	return TaskID(s), nil
}

func (id TaskID) String() string { return string(id) }

// UserID はユーザーを識別するID
type UserID string

// NewUserID は文字列からUserIDを作る。空文字は許さない
func NewUserID(s string) (UserID, error) {
	if s == "" {
		return "", ErrUserIDRequired
	}
	return UserID(s), nil
}

func (id UserID) String() string { return string(id) }
