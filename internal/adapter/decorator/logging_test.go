// internal/adapter/decorator/logging_test.go
package decorator_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/decorator"
)

// echo は入力をそのまま返し、"fail" のときだけ失敗するユースケース
type echo struct{}

func (echo) Execute(_ context.Context, in string) (string, error) {
	if in == "fail" {
		return "", errors.New("失敗しました")
	}
	return in, nil
}

func TestWithLogging(t *testing.T) {
	var buf bytes.Buffer
	lg := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	uc := decorator.WithLogging[string, string]("Echo", echo{}, lg)

	out, err := uc.Execute(context.Background(), "hello")
	if err != nil || out != "hello" {
		t.Fatalf("元のユースケースの結果がそのまま返るはず: %q, %v", out, err)
	}
	if _, err := uc.Execute(context.Background(), "fail"); err == nil {
		t.Fatal("エラーもそのまま返るはず")
	}

	logs := buf.String()
	if !strings.Contains(logs, `"usecase":"Echo"`) || !strings.Contains(logs, "ユースケースが失敗しました") {
		t.Errorf("ログが出ていない: %s", logs)
	}
}
