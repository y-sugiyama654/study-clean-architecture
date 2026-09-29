// internal/usecase/errors.go
package usecase

import (
	"errors"
	"fmt"
)

// ErrNotFound は、対象が存在しない（または操作する権限のあるユーザーから見えない）ことを表す分類
var ErrNotFound = errors.New("見つかりません")

// ErrTaskNotFound はタスクが見つからないことを表す
var ErrTaskNotFound = fmt.Errorf("%w: タスクが存在しません", ErrNotFound)
