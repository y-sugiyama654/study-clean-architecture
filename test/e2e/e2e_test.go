//go:build integration

// test/e2e/e2e_test.go
//
// アプリケーションを実際にビルドして起動し、HTTPで外側から操作するE2Eテスト。
//
//	go test -tags=integration ./test/e2e/
package e2e_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/y-sugiyama654/study-clean-architecture/internal/testutil/postgrestest"
)

func TestAPI(t *testing.T) {
	_, dbURL := postgrestest.Start(t)

	// cmd/api をビルドする
	bin := filepath.Join(t.TempDir(), "api")
	build := exec.Command("go", "build", "-o", bin, "../../cmd/api")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("ビルドに失敗しました: %v\n%s", err, out)
	}

	// 空いているポートで起動する
	addr := freeAddr(t)
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"HTTP_ADDR="+addr,
		"DATABASE_URL="+dbURL,
		"API_TOKENS=e2e-token-alice:alice,e2e-token-bob:bob", // テスト用のダミーのトークン
		"LOG_LEVEL=warn",
	)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	base := "http://" + addr
	waitReady(t, base)

	c := client{t: t, base: base}

	// 認証
	c.expect("GET", "/tasks", "", "", http.StatusUnauthorized)

	// 作成 → 取得 → 完了
	created := c.expect("POST", "/tasks", "e2e-token-alice", `{"title":"牛乳を買う"}`, http.StatusCreated)
	id := created["id"].(string)
	c.expect("GET", "/tasks/"+id, "e2e-token-alice", "", http.StatusOK)
	c.expect("GET", "/tasks/"+id, "e2e-token-bob", "", http.StatusNotFound) // 他人のタスクは見えない
	done := c.expect("POST", "/tasks/"+id+"/complete", "e2e-token-alice", "", http.StatusOK)
	if done["status"] != "done" {
		t.Errorf("完了になっていない: %v", done)
	}
	conflict := c.expect("POST", "/tasks/"+id+"/complete", "e2e-token-alice", "", http.StatusConflict)
	if conflict["error"].(map[string]any)["code"] != "conflict" {
		t.Errorf("エラーコードが不正: %v", conflict)
	}
	c.expect("POST", "/tasks", "e2e-token-alice", `{"title":""}`, http.StatusBadRequest)

	// SIGTERM を送ると、グレースフルシャットダウンして正常終了する
	cmd.Process.Signal(syscall.SIGTERM)
	exited := make(chan error, 1)
	go func() { exited <- cmd.Wait() }()
	select {
	case err := <-exited:
		if err != nil {
			t.Errorf("正常終了しなかった: %v", err)
		}
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
		t.Fatal("SIGTERMを送っても終了しない")
	}
}

type client struct {
	t    *testing.T
	base string
}

// expect はリクエストを送り、ステータスコードを確かめて、JSONのボディを返す
func (c client) expect(method, path, token, body string, want int) map[string]any {
	c.t.Helper()
	req, _ := http.NewRequest(method, c.base+path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		c.t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode != want {
		c.t.Fatalf("%s %s: want %d, got %d: %s", method, path, want, res.StatusCode, raw)
	}
	var m map[string]any
	json.Unmarshal(raw, &m)
	return m
}

func freeAddr(t *testing.T) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	return ln.Addr().String()
}

// waitReady はサーバーが応答するようになるまで待つ
func waitReady(t *testing.T, base string) {
	for range 50 {
		res, err := http.Get(base + "/tasks")
		if err == nil {
			res.Body.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal(fmt.Sprintf("サーバーが起動しません: %s", base))
}
