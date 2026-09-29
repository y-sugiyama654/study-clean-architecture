// internal/adapter/decorator/logging.go
package decorator

import (
	"context"
	"log/slog"
	"time"
)

// Usecase は Execute を1つ持つユースケースを表す（コントローラーの入力ポートと同じ形）
type Usecase[In, Out any] interface {
	Execute(ctx context.Context, in In) (Out, error)
}

// logging はユースケースを包み、呼び出しの結果と処理時間をログに出すデコレータ
type logging[In, Out any] struct {
	name string
	next Usecase[In, Out]
	lg   *slog.Logger
}

// WithLogging はユースケースにログ出力を付け足す
// 元のユースケースと同じインターフェースを実装するので、コントローラーからは区別がつかない
func WithLogging[In, Out any](name string, next Usecase[In, Out], lg *slog.Logger) Usecase[In, Out] {
	return &logging[In, Out]{name: name, next: next, lg: lg}
}

func (d *logging[In, Out]) Execute(ctx context.Context, in In) (Out, error) {
	start := time.Now()
	out, err := d.next.Execute(ctx, in)
	attrs := []any{slog.String("usecase", d.name), slog.Duration("duration", time.Since(start))}
	if err != nil {
		d.lg.WarnContext(ctx, "ユースケースが失敗しました", append(attrs, slog.Any("error", err))...)
	} else {
		d.lg.DebugContext(ctx, "ユースケースを実行しました", attrs...)
	}
	return out, err
}
