// internal/infrastructure/web/server.go
package web

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// ShutdownTimeout は、シャットダウン時に処理中のリクエストの完了を待つ最大時間
const ShutdownTimeout = 10 * time.Second

// Serve はlnでHTTPサーバーを動かし、ctxがキャンセルされたら安全に停止する（グレースフルシャットダウン）
// 新しいリクエストの受け付けをやめたうえで、処理中のリクエストが終わるのを待ってから戻る
func Serve(ctx context.Context, srv *http.Server, ln net.Listener, lg *slog.Logger) error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err // シャットダウンを指示する前にサーバーが止まった
	case <-ctx.Done():
	}

	lg.Info("シャットダウンを開始します")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	lg.Info("シャットダウンが完了しました")
	return nil
}
