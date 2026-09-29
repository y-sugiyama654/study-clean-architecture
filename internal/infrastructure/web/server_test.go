// internal/infrastructure/web/server_test.go
package web_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/web"
)

func TestServe_GracefulShutdown(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	// 処理に300msかかるハンドラ
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		io.WriteString(w, "done")
	})}

	ctx, cancel := context.WithCancel(context.Background())
	served := make(chan error, 1)
	go func() { served <- web.Serve(ctx, srv, ln, slog.New(slog.DiscardHandler)) }()

	// リクエストを送り、処理中にシャットダウンを指示する
	resCh := make(chan *http.Response, 1)
	go func() {
		res, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			t.Error(err)
		}
		resCh <- res
	}()
	time.Sleep(100 * time.Millisecond)
	cancel()

	// 処理中だったリクエストは、最後まで処理されてから返る
	res := <-resCh
	if res == nil || res.StatusCode != http.StatusOK {
		t.Fatalf("処理中のリクエストが完了しなかった: %v", res)
	}
	body, _ := io.ReadAll(res.Body)
	res.Body.Close()
	if string(body) != "done" {
		t.Errorf("レスポンスが途中で切れた: %q", body)
	}
	if err := <-served; err != nil {
		t.Errorf("Serveがエラーを返した: %v", err)
	}

	// シャットダウン後は新しい接続を受け付けない
	if _, err := http.Get("http://" + ln.Addr().String()); err == nil {
		t.Error("シャットダウン後もリクエストを受け付けている")
	}
}
