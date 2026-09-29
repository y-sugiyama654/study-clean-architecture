// internal/infrastructure/web/router.go
package web

import (
	"log/slog"
	"net/http"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/controller"
	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/middleware"
)

// NewRouter はURLとHTTPメソッドを、コントローラーのメソッドに対応づける
// Go 1.22 以降の http.ServeMux は「メソッド パス」の形のパターンと、{id} のようなパスパラメータを扱える
func NewRouter(tasks *controller.TaskController, verifier middleware.TokenVerifier, lg *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", tasks.Create)
	mux.HandleFunc("GET /tasks", tasks.List)
	mux.HandleFunc("GET /tasks/{id}", tasks.Get)
	mux.HandleFunc("POST /tasks/{id}/complete", tasks.Complete)

	// ミドルウェアは外側から順に実行される:
	// Recover → RequestID → AccessLog → Authenticate → ルーティング → コントローラー
	var h http.Handler = mux
	h = middleware.Authenticate(verifier)(h)
	h = middleware.AccessLog(lg)(h)
	h = middleware.RequestID(h)
	h = middleware.Recover(lg)(h)
	return h
}
