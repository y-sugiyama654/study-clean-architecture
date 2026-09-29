// internal/domain/title_test.go
package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

func TestNewTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr error
	}{
		{name: "通常のタイトル", input: "牛乳を買う", want: "牛乳を買う"},
		{name: "前後の空白は取り除く", input: "  牛乳を買う  ", want: "牛乳を買う"},
		{name: "ちょうど100文字", input: strings.Repeat("あ", 100), want: strings.Repeat("あ", 100)},
		{name: "空文字", input: "", wantErr: domain.ErrTitleRequired},
		{name: "空白だけ", input: "   ", wantErr: domain.ErrTitleRequired},
		{name: "101文字", input: strings.Repeat("あ", 101), wantErr: domain.ErrTitleTooLong},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewTitle(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: want %v, got %v", tt.wantErr, err)
			}
			if got.String() != tt.want {
				t.Errorf("want %q, got %q", tt.want, got.String())
			}
		})
	}
}

func TestNewDescription(t *testing.T) {
	if _, err := domain.NewDescription(""); err != nil {
		t.Errorf("説明は空でもよい: %v", err)
	}
	if _, err := domain.NewDescription(strings.Repeat("あ", 1001)); !errors.Is(err, domain.ErrDescriptionTooLong) {
		t.Errorf("1001文字はエラーになるはず: got %v", err)
	}
}

func TestErrorCategories(t *testing.T) {
	// 個々のエラーは分類（ErrValidation / ErrConflict）で判定できる
	if !errors.Is(domain.ErrTitleTooLong, domain.ErrValidation) {
		t.Error("ErrTitleTooLong は ErrValidation に分類されるはず")
	}
	if !errors.Is(domain.ErrTaskAlreadyCompleted, domain.ErrConflict) {
		t.Error("ErrTaskAlreadyCompleted は ErrConflict に分類されるはず")
	}
	t.Log(domain.ErrTitleTooLong)
}
