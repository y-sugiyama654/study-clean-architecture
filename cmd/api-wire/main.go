// cmd/api-wire/main.go
//
// cmd/api と同じアプリケーションを、依存の組み立てにWireを使って起動する（比較用）
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/config"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/database"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/logger"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/web"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	lg := logger.New(os.Stdout, cfg.LogLevel)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.OpenPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	ln, err := net.Listen("tcp", cfg.HTTPAddr)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           initializeHandler(db), // wire_gen.go に生成された関数
		ReadHeaderTimeout: 5 * time.Second,
	}
	lg.Info("HTTPサーバーを起動します", slog.String("addr", ln.Addr().String()))
	return web.Serve(ctx, srv, ln, lg)
}
