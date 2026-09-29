// internal/domain/title.go
package domain

import (
	"strings"
	"unicode/utf8"
)

// MaxTitleLength はタイトルの最大文字数
const MaxTitleLength = 100

// Title はタスクのタイトルを表す値オブジェクト
// フィールドを非公開にしているので、NewTitle を通さずに不正な値を作ることはできない
type Title struct {
	value string
}

// NewTitle は前後の空白を取り除いたうえで、ルールを満たすタイトルを作る
func NewTitle(s string) (Title, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Title{}, ErrTitleRequired
	}
	if utf8.RuneCountInString(s) > MaxTitleLength {
		return Title{}, ErrTitleTooLong
	}
	return Title{value: s}, nil
}

func (t Title) String() string { return t.value }
