// internal/adapter/cli/cli.go
package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/usecase"
)

// CLIが使うユースケース（HTTPのコントローラーと同じユースケースを、別の入口から使う）
type (
	CreateTaskUsecase interface {
		Execute(ctx context.Context, in usecase.CreateTaskInput) (usecase.TaskOutput, error)
	}
	ListTasksUsecase interface {
		Execute(ctx context.Context, in usecase.ListTasksInput) ([]usecase.TaskOutput, error)
	}
	CompleteTaskUsecase interface {
		Execute(ctx context.Context, in usecase.CompleteTaskInput) (usecase.TaskOutput, error)
	}
)

// CLI はコマンドライン引数をユースケースの入力に変換し、結果をテキストで出力する
// HTTPのコントローラーとプレゼンターにあたる役割を、CLI向けに担う
type CLI struct {
	create   CreateTaskUsecase
	list     ListTasksUsecase
	complete CompleteTaskUsecase
	out      io.Writer
}

func New(create CreateTaskUsecase, list ListTasksUsecase, complete CompleteTaskUsecase, out io.Writer) *CLI {
	return &CLI{create: create, list: list, complete: complete, out: out}
}

// ErrUsage は使い方が正しくないことを表す
var ErrUsage = errors.New(`使い方:
  taskctl add <タイトル> [説明]   タスクを作成する
  taskctl list                   タスクの一覧を表示する
  taskctl done <タスクID>         タスクを完了にする`)

// Run はサブコマンドを実行する。userID は操作するユーザー
func (c *CLI) Run(ctx context.Context, userID string, args []string) error {
	if len(args) == 0 {
		return ErrUsage
	}
	switch cmd, rest := args[0], args[1:]; {
	case cmd == "add" && (len(rest) == 1 || len(rest) == 2):
		in := usecase.CreateTaskInput{UserID: userID, Title: rest[0]}
		if len(rest) == 2 {
			in.Description = rest[1]
		}
		out, err := c.create.Execute(ctx, in)
		if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "作成しました: %s %s\n", out.ID, out.Title)
		return nil
	case cmd == "list" && len(rest) == 0:
		outs, err := c.list.Execute(ctx, usecase.ListTasksInput{UserID: userID})
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(c.out, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tSTATUS\tTITLE\tCREATED")
		for _, t := range outs {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.ID, t.Status, t.Title, t.CreatedAt.Local().Format(time.DateTime))
		}
		return w.Flush()
	case cmd == "done" && len(rest) == 1:
		out, err := c.complete.Execute(ctx, usecase.CompleteTaskInput{UserID: userID, TaskID: rest[0]})
		if err != nil {
			return err
		}
		fmt.Fprintf(c.out, "完了にしました: %s %s\n", out.ID, out.Title)
		return nil
	default:
		return ErrUsage
	}
}
