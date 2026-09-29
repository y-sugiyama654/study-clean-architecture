// internal/domain/errors.go
package domain

import (
	"errors"
	"fmt"
)

// エラーの分類。個々のエラーはこのどちらかをラップしているので、
// 呼び出し側は errors.Is(err, domain.ErrValidation) のように分類で判定できる
var (
	// ErrValidation は、値がルールを満たしていないことを表す
	ErrValidation = errors.New("入力値が不正です")
	// ErrConflict は、今の状態ではその操作ができないことを表す
	ErrConflict = errors.New("現在の状態ではその操作はできません")
)

var (
	ErrTaskIDRequired       = fmt.Errorf("%w: タスクIDは必須です", ErrValidation)
	ErrUserIDRequired       = fmt.Errorf("%w: ユーザーIDは必須です", ErrValidation)
	ErrTitleRequired        = fmt.Errorf("%w: タイトルは必須です", ErrValidation)
	ErrTitleTooLong         = fmt.Errorf("%w: タイトルは%d文字以内にしてください", ErrValidation, MaxTitleLength)
	ErrDescriptionTooLong   = fmt.Errorf("%w: 説明は%d文字以内にしてください", ErrValidation, MaxDescriptionLength)
	ErrInvalidStatus        = fmt.Errorf("%w: 不明なステータスです", ErrValidation)
	ErrTaskAlreadyCompleted = fmt.Errorf("%w: タスクは既に完了しています", ErrConflict)
)
