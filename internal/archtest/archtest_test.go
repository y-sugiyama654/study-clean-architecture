// internal/archtest/archtest_test.go
//
// 依存性のルール（内側の層は外側の層を知らない）が守られているかを、import文から機械的に検査するテスト
package archtest_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
)

const module = "github.com/y-sugiyama654/study-clean-architecture"

// layers は内側から外側への層の並び。パッケージのパスの前方一致で、どの層に属するかを決める
var layers = []struct {
	name     string
	prefixes []string
}{
	{"Entities", []string{module + "/internal/domain"}},
	{"Use Cases", []string{module + "/internal/usecase"}},
	{"Interface Adapters", []string{module + "/internal/adapter"}},
	{"Frameworks & Drivers", []string{module + "/internal/infrastructure", module + "/internal/testutil", module + "/cmd"}},
}

// forbiddenStdlib は、内側の2つの層で使ってはいけない標準ライブラリ（HTTPやDBの詳細を持ち込まないため）
var forbiddenStdlib = []string{"net/http", "database/sql", "encoding/json"}

type pkg struct {
	ImportPath string
	Imports    []string
}

func TestDependencyRule(t *testing.T) {
	for _, p := range listPackages(t) {
		from := layerOf(p.ImportPath)
		if from < 0 {
			continue
		}
		for _, imp := range p.Imports {
			if to := layerOf(imp); to > from {
				t.Errorf("%s（%s）が外側の %s（%s）をimportしている",
					short(p.ImportPath), layers[from].name, short(imp), layers[to].name)
			}
			if from <= 1 && contains(forbiddenStdlib, imp) {
				t.Errorf("%s（%s）が %s をimportしている", short(p.ImportPath), layers[from].name, imp)
			}
			if strings.HasPrefix(imp, module+"/internal/testutil") {
				t.Errorf("%s がテスト用のパッケージ %s をimportしている", short(p.ImportPath), short(imp))
			}
		}
	}
}

// listPackages はモジュール内の全パッケージと、そのimportの一覧を go list で取得する
// （テストファイルのimportは含まない）
func listPackages(t *testing.T) []pkg {
	t.Helper()
	cmd := exec.Command("go", "list", "-json", module+"/...")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var pkgs []pkg
	dec := json.NewDecoder(out)
	for {
		var p pkg
		if err := dec.Decode(&p); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		pkgs = append(pkgs, p)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("go list に失敗しました（循環importなど、コンパイルできない状態かもしれません）: %v\n%s", err, stderr.String())
	}
	if len(pkgs) == 0 {
		t.Fatal("パッケージが見つかりません")
	}
	return pkgs
}

func layerOf(path string) int {
	for i, l := range layers {
		for _, prefix := range l.prefixes {
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				return i
			}
		}
	}
	return -1
}

func short(path string) string { return strings.TrimPrefix(path, module+"/") }

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
