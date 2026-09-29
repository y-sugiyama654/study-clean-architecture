// internal/domain/description.go
package domain

import "unicode/utf8"

// MaxDescriptionLength は説明の最大文字数
const MaxDescriptionLength = 1000

// Description はタスクの説明を表す値オブジェクト（空でもよい）
type Description struct {
	value string
}

// NewDescription はルールを満たす説明を作る
func NewDescription(s string) (Description, error) {
	if utf8.RuneCountInString(s) > MaxDescriptionLength {
		return Description{}, ErrDescriptionTooLong
	}
	return Description{value: s}, nil
}

func (d Description) String() string { return d.value }
