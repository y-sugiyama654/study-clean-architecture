// internal/infrastructure/auth/static_token_verifier.go
package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/middleware"
	"github.com/y-sugiyama654/study-clean-architecture/internal/domain"
)

// StaticTokenVerifier は、あらかじめ設定したトークンとユーザーIDの対応表でトークンを検証する
// 学習用の単純な実装。実際のサービスでは、OpenID Connect のIDトークンやJWTの検証などに置き換える
type StaticTokenVerifier struct {
	users map[[sha256.Size]byte]domain.UserID // トークンのハッシュ → ユーザーID
}

var _ middleware.TokenVerifier = (*StaticTokenVerifier)(nil)

// NewStaticTokenVerifier は「トークン → ユーザーID」の対応表から検証器を作る
// トークンそのものはメモリに残さず、SHA-256のハッシュだけを持つ
func NewStaticTokenVerifier(tokens map[string]string) *StaticTokenVerifier {
	users := make(map[[sha256.Size]byte]domain.UserID, len(tokens))
	for token, user := range tokens {
		users[sha256.Sum256([]byte(token))] = domain.UserID(user)
	}
	return &StaticTokenVerifier{users: users}
}

func (v *StaticTokenVerifier) Verify(_ context.Context, token string) (domain.UserID, error) {
	got := sha256.Sum256([]byte(token))
	// 比較にかかる時間からトークンを推測されないよう、一致するかどうかに関係なく全件を定数時間で比べる
	var found domain.UserID
	for want, user := range v.users {
		if subtle.ConstantTimeCompare(got[:], want[:]) == 1 {
			found = user
		}
	}
	if found == "" {
		return "", middleware.ErrInvalidToken
	}
	return found, nil
}
