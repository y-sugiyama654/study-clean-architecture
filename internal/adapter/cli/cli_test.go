// internal/adapter/cli/cli_test.go
package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/cli"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/memory"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/gateway/notifier"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/system"
	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

func TestCLI(t *testing.T) {
	s := memory.NewStore()
	tx, tasks, activities := memory.NewTransactor(s), memory.NewTaskRepository(s), memory.NewActivityRepository(s)
	var out bytes.Buffer
	app := cli.New(
		usecase.NewCreateTask(tx, tasks, activities, system.UUIDGenerator{}, system.Clock{}),
		usecase.NewListTasks(tasks),
		usecase.NewCompleteTask(tx, tasks, activities, notifier.Nop{}, system.Clock{}),
		&out,
	)
	ctx := context.Background()

	if err := app.Run(ctx, "alice", []string{"add", "牛乳を買う"}); err != nil {
		t.Fatal(err)
	}
	// 「作成しました: <ID> 牛乳を買う」からIDを取り出す
	id := strings.Fields(out.String())[1]

	if err := app.Run(ctx, "alice", []string{"done", id}); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := app.Run(ctx, "alice", []string{"list"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), id) || !strings.Contains(out.String(), "done") {
		t.Errorf("一覧に完了したタスクが出ていない:\n%s", out.String())
	}
	t.Logf("\n%s", out.String())

	if err := app.Run(ctx, "alice", []string{"remove", id}); !errors.Is(err, cli.ErrUsage) {
		t.Errorf("未知のコマンドは使い方のエラーになるはず: %v", err)
	}
}
