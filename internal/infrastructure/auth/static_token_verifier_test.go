// internal/infrastructure/auth/static_token_verifier_test.go
package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/y-sugiyama654/study-clean-architecture/internal/adapter/middleware"
	"github.com/y-sugiyama654/study-clean-architecture/internal/infrastructure/auth"
)

func TestStaticTokenVerifier(t *testing.T) {
	v := auth.NewStaticTokenVerifier(map[string]string{"test-token-a": "alice", "test-token-b": "bob"})

	user, err := v.Verify(context.Background(), "test-token-b")
	if err != nil || user != "bob" {
		t.Errorf("want bob, got %q (%v)", user, err)
	}
	if _, err := v.Verify(context.Background(), "unknown"); !errors.Is(err, middleware.ErrInvalidToken) {
		t.Errorf("want ErrInvalidToken, got %v", err)
	}
}
