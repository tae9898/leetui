package stmt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"leetui/internal/ps/lcjudge"
	"leetui/internal/ps/problem"
)

// oracleRoot는 .lc-auth와 two-sum Question 캐시를 갖춘 루트를 만든다.
func oracleRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeTemplate(t, root)
	t.Setenv("LC_SESSION", "")
	t.Setenv("LC_CSRF", "")
	if err := os.WriteFile(filepath.Join(root, ".lc-auth"), []byte(`{"session":"s","csrf":"c"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	q := loadFixture(t, "two-sum")
	q.Slug = "two-sum"
	qd, _ := json.Marshal(q)
	if err := os.MkdirAll(filepath.Join(root, ".stmt", "lc"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".stmt", "lc", "two-sum.json"), qd, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func oracleJudge(t *testing.T, root string, srv *httptest.Server) *lcjudge.Judge {
	t.Helper()
	return &lcjudge.Judge{
		Base: srv.URL, HTTP: srv.Client(),
		Delay: time.Millisecond, Backoff: time.Millisecond, PollBudget: time.Second,
		CachePath: filepath.Join(root, ".stmt", "lc", "two-sum.oracle.json"),
	}
}

// fakeOracle는 모든 케이스에 [0,1]을 정답으로 돌려준다.
func fakeOracle() (*atomic.Int32, *httptest.Server) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/problems/two-sum/interpret_solution/":
			hits.Add(1)
			fmt.Fprint(w, `{"interpret_id":"id-1"}`)
		case strings.HasSuffix(r.URL.Path, "/check/"):
			fmt.Fprint(w, `{"state":"SUCCESS","expected_code_answer":["[0,1]"]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	return &hits, srv
}

func TestFillOracleFillsMissingOuts(t *testing.T) {
	root := oracleRoot(t)
	hits, srv := fakeOracle()
	defer srv.Close()

	p := setupProblem(t, root, "two-sum")
	// 1.in/1.out은 problem.New가 만듦 — 손으로 쓴 정답이 있으니 보존 대상
	if err := os.WriteFile(filepath.Join(p.Dir, "1.out"), []byte("9 9\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// 2.in: 유효 입력 (nums=[3,3], target=6)
	if err := os.WriteFile(filepath.Join(p.Dir, "2.in"), []byte("2\n3 3\n6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// 3.in: 형식 오류 — 네트워크 안 가고 실패 목록에
	if err := os.WriteFile(filepath.Join(p.Dir, "3.in"), []byte("garbage\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// custom.in: .out 없어도 스크래치라 제외
	if f, err := os.OpenFile(filepath.Join(p.Dir, "custom.in"), os.O_CREATE, 0o644); err == nil {
		f.Close()
	}

	res, err := fillOracle(context.Background(), root, "two-sum", p, oracleJudge(t, root, srv))
	if err != nil {
		t.Fatal(err)
	}
	if res.Targets != 2 { // 2.in, 3.in (1.in은 .out 있음, custom 제외)
		t.Fatalf("대상 %d, want 2", res.Targets)
	}
	if len(res.Filled) != 1 || res.Filled[0] != "2" {
		t.Fatalf("Filled: %v", res.Filled)
	}
	if len(res.Failed) != 1 || !strings.Contains(res.Failed[0], "3.in") {
		t.Fatalf("Failed: %v", res.Failed)
	}
	if got := readFile(t, filepath.Join(p.Dir, "2.out")); got != "0 1\n" {
		t.Fatalf("2.out: %q", got)
	}
	if got := readFile(t, filepath.Join(p.Dir, "1.out")); got != "9 9\n" {
		t.Fatalf("기존 .out이 덮어써짐: %q", got)
	}
	if _, err := os.Stat(filepath.Join(p.Dir, "3.out")); err == nil {
		t.Fatal("실패 케이스에 .out이 생김")
	}
	marked := OracleMarked(p.Dir)
	if !marked["2"] || marked["3"] {
		t.Fatalf(".oracle 표식: %v", marked)
	}
	if !strings.Contains(res.Summary(), "2개 중 1개 채움") {
		t.Fatalf("Summary: %s", res.Summary())
	}
	if hits.Load() != 1 { // 유효 케이스만 전송
		t.Fatalf("interpret 호출 %d회, want 1", hits.Load())
	}

	// 다시 실행 — 2.out이 이미 있어 대상에서 빠지고 3.in만 남는다
	res2, err := fillOracle(context.Background(), root, "two-sum", p, oracleJudge(t, root, srv))
	if err != nil {
		t.Fatal(err)
	}
	if res2.Targets != 1 {
		t.Fatalf("재실행 대상 %d, want 1 (3.in만)", res2.Targets)
	}
}

func TestFillOracleNoTargets(t *testing.T) {
	root := oracleRoot(t)
	p := setupProblem(t, root, "two-sum") // 1.in/1.out만 있음
	res, err := fillOracle(context.Background(), root, "two-sum", p, oracleJudge(t, root, httptest.NewServer(nil)))
	if err != nil {
		t.Fatal(err)
	}
	if res.Targets != 0 || !strings.Contains(res.Summary(), "채울 .out 없음") {
		t.Fatalf("대상 없음 처리: %+v %s", res, res.Summary())
	}
}

func TestFillOracleNoAuth(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root)
	t.Setenv("LC_SESSION", "")
	t.Setenv("LC_CSRF", "")
	p := setupProblem(t, root, "oracle-noauth")
	if err := os.WriteFile(filepath.Join(p.Dir, "2.in"), []byte("1\n5\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, srv := fakeOracle()
	defer srv.Close()
	j := &lcjudge.Judge{Base: srv.URL, HTTP: srv.Client()}

	_, err := fillOracle(context.Background(), root, "two-sum", p, j)
	if err == nil || !strings.Contains(err.Error(), "미설정") {
		t.Fatalf("인증 없음 안내가 나와야 함: %v", err)
	}
}

// 캐시가 있으면 같은 입력을 다시 보내지 않는다 (히든 케이스 근사 반복 사용).
func TestFillOracleUsesCache(t *testing.T) {
	root := oracleRoot(t)
	hits, srv := fakeOracle()
	defer srv.Close()

	p := setupProblem(t, root, "two-sum")
	if err := os.WriteFile(filepath.Join(p.Dir, "2.in"), []byte("2\n3 3\n6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := fillOracle(context.Background(), root, "two-sum", p, oracleJudge(t, root, srv)); err != nil {
		t.Fatal(err)
	}
	// .out을 지우고 다시 — 같은 입력이므로 캐시로 채운다
	if err := os.Remove(filepath.Join(p.Dir, "2.out")); err != nil {
		t.Fatal(err)
	}
	res, err := fillOracle(context.Background(), root, "two-sum", p, oracleJudge(t, root, srv))
	if err != nil {
		t.Fatal(err)
	}
	if res.Cached != 1 || hits.Load() != 1 {
		t.Fatalf("캐시 미사용: Cached=%d hits=%d", res.Cached, hits.Load())
	}
}

func TestFillOracleUnsupported(t *testing.T) {
	root := oracleRoot(t)
	q := loadFixture(t, "lru-cache")
	qd, _ := json.Marshal(q)
	if err := os.WriteFile(filepath.Join(root, ".stmt", "lc", "lru-cache.json"), qd, 0o644); err != nil {
		t.Fatal(err)
	}
	p := problem.Problem{Site: "lc", Name: "0146-lru-cache", Dir: filepath.Join(root, "lc", "0146-lru-cache")}
	if err := os.MkdirAll(p.Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p.Dir, "2.in"), []byte("2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, srv := fakeOracle()
	defer srv.Close()

	_, err := fillOracle(context.Background(), root, "lru-cache", p, &lcjudge.Judge{Base: srv.URL, HTTP: srv.Client()})
	if err == nil || !strings.Contains(err.Error(), "미지원") {
		t.Fatalf("미지원 안내: %v", err)
	}
}
