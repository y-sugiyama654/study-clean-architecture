// internal/adapter/middleware/auth.go
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/presenter"
	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// ErrInvalidToken はトークンが正しくないことを表す
var ErrInvalidToken = errors.New("トークンが正しくありません")

// TokenVerifier はトークンを検証し、そのトークンの持ち主のユーザーIDを返す
// 検証の方法（固定のトークン表、JWT、外部の認証サービスなど）は外側の層が決める
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (domain.UserID, error)
}

type userIDKey struct{}

// WithUserID は認証済みのユーザーIDを context に入れる
func WithUserID(ctx context.Context, id domain.UserID) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

// UserID は認証済みのユーザーIDを context から取り出す（認証されていなければ空文字）
func UserID(ctx context.Context) domain.UserID {
	id, _ := ctx.Value(userIDKey{}).(domain.UserID)
	return id
}

// Authenticate は Authorization: Bearer <トークン> を検証し、ユーザーIDを context に入れて次に渡す
// トークンがない・正しくない場合は、ここで 401 Unauthorized を返して先へ進ませない
func Authenticate(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				unauthorized(w, "認証が必要です")
				return
			}
			userID, err := verifier.Verify(r.Context(), token)
			if err != nil {
				unauthorized(w, err.Error())
				return
			}
			next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), userID)))
		})
	}
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	presenter.JSON(w, http.StatusUnauthorized, presenter.ErrorResponse{
		Error: presenter.ErrorBody{Code: "unauthenticated", Message: msg},
	})
}
