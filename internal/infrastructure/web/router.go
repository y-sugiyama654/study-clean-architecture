// internal/infrastructure/web/router.go
package web

import (
	"net/http"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/controller"
)

// NewRouter はURLとHTTPメソッドを、コントローラーのメソッドに対応づける
// Go 1.22 以降の http.ServeMux は「メソッド パス」の形のパターンと、{id} のようなパスパラメータを扱える
func NewRouter(tasks *controller.TaskController) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /tasks", tasks.Create)
	mux.HandleFunc("GET /tasks", tasks.List)
	mux.HandleFunc("GET /tasks/{id}", tasks.Get)
	mux.HandleFunc("POST /tasks/{id}/complete", tasks.Complete)
	return mux
}
