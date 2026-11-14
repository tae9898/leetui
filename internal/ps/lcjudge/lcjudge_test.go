package lcjudge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"leetui/internal/ps/lcref"
)

func testQuestion() lcref.Question {
	return lcref.Question{
		Slug: "two-sum", QuestionID: "1",
		MetaData:    `{"name":"twoSum","params":[{"name":"nums","type":"integer[]"},{"name":"target","type":"integer"}],"return":{"type":"integer[]"}}`,
		CodeSnippet: "class Solution { public: vector<int> twoSum(vector<int>& a, int t) {} };",
	}
}

// fakeLC는 interpret_solution/check 엔드포인트를 흉내 낸다.
// expected는 data_input 그대로를 정답으로 되돌린다(동작 확인용).
type fakeLC struct {
	mu       sync.Mutex
	interps  int // interpret_solution 호출 수
	checks   int
	fail403  bool
	emptyAns string // 이 입력엔 정답 없음 처리
	pending  int    // SUCCESS 전 STARTED 응답 횟수
}

func (f *fakeLC) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/problems/two-sum/interpret_solution/":
			f.mu.Lock()
			defer f.mu.Unlock()
			f.interps++
			if f.fail403 {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			var req struct {
				Lang       string `json:"lang"`
				QuestionID string `json:"question_id"`
				TypedCode  string `json:"typed_code"`
				DataInput  string `json:"data_input"`
			}
			_ = json.NewDecoder(r.Body).Decode(&req)
			fmt.Fprintf(w, `{"interpret_id":"id-%d"}`, f.interps)
		case len(r.URL.Path) > len("/submissions/detail/") && r.URL.Path[len(r.URL.Path)-len("/check/"):] == "/check/":
			f.mu.Lock()
			defer f.mu.Unlock()
			f.checks++
			if f.checks <= f.pending {
				w.Write([]byte(`{"state":"STARTED"}`))
				return
			}
			ans, _ := json.Marshal([]string{"[0,1]"})
			fmt.Fprintf(w, `{"state":"SUCCESS","expected_code_answer":%s}`, ans)
		default:
			http.NotFound(w, r)
		}
	})
}

func fastJudge(t *testing.T, cachePath string) *Judge {
	t.Helper()
	return &Judge{Delay: time.Millisecond, Backoff: time.Millisecond, PollBudget: time.Second, CachePath: cachePath}
}

func TestExpectedPollsUntilSuccess(t *testing.T) {
	f := &fakeLC{pending: 2}
	srv := httptest.NewServer(f.handler())
	defer srv.Close()
	j := fastJudge(t, filepath.Join(t.TempDir(), "c.json"))
	j.Base, j.HTTP = srv.URL, srv.Client()

	got, err := j.Expected(context.Background(), Auth{Session: "s", CSRF: "c"}, testQuestion(),
		[][]string{{"[1,2,3]", "6"}})
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Answer != "[0,1]" || got[0].Err != nil {
		t.Fatalf("결과: %+v", got[0])
	}
	if f.checks < 3 {
		t.Fatalf("STARTED 2번 뒤 SUCCESS여야 — checks=%d", f.checks)
	}
}

func TestExpectedCacheHit(t *testing.T) {
	f := &fakeLC{}
	srv := httptest.NewServer(f.handler())
	defer srv.Close()
	cache := filepath.Join(t.TempDir(), "c.json")
	j := fastJudge(t, cache)
	j.Base, j.HTTP = srv.URL, srv.Client()
	auth := Auth{Session: "s", CSRF: "c"}
	inputs := [][]string{{"[1,2,3]", "6"}, {"[2,3,4]", "5"}}

	if _, err := j.Expected(context.Background(), auth, testQuestion(), inputs); err != nil {
		t.Fatal(err)
	}
	first := f.interps
	// 같은 입력 재요청 — 캐시로 네트워크 없음
	got, err := j.Expected(context.Background(), auth, testQuestion(), inputs)
	if err != nil {
		t.Fatal(err)
	}
	if f.interps != first {
		t.Fatalf("캐시가 무시됨: interpret %d회", f.interps)
	}
	if !got[0].Cached {
		t.Fatalf("Cached=false")
	}
	// 캐시 파일에서도 읽힌다(새 Judge)
	j2 := fastJudge(t, cache)
	got2, err := j2.Expected(context.Background(), auth, testQuestion(), inputs)
	if err != nil {
		t.Fatal(err)
	}
	if got2[0].Answer != "[0,1]" {
		t.Fatalf("캐시 파일 로드 실패: %+v", got2[0])
	}
}

func TestExpectedAuthError(t *testing.T) {
	f := &fakeLC{fail403: true}
	srv := httptest.NewServer(f.handler())
	defer srv.Close()
	j := fastJudge(t, "")
	j.Base, j.HTTP = srv.URL, srv.Client()

	_, err := j.Expected(context.Background(), Auth{Session: "s", CSRF: "c"}, testQuestion(),
		[][]string{{"[1]", "2"}})
	if err != ErrAuth {
		t.Fatalf("ErrAuth여야 함: %v", err)
	}
}

// 429는 남은 케이스를 계속 치지 않고 전체 중단.
func TestExpectedRateLimitAborts(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	j := fastJudge(t, "")
	j.Base, j.HTTP = srv.URL, srv.Client()

	_, err := j.Expected(context.Background(), Auth{Session: "s", CSRF: "c"}, testQuestion(),
		[][]string{{"[1]", "2"}, {"[2]", "3"}, {"[3]", "4"}})
	if !errors.Is(err, ErrRate) {
		t.Fatalf("ErrRate여야 함: %v", err)
	}
	if hits.Load() != 1 {
		t.Fatalf("429 뒤 계속 요청함: %d회", hits.Load())
	}
}

func TestExpectedNoAnswerPerCase(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/problems/two-sum/interpret_solution/" {
			w.Write([]byte(`{"interpret_id":"id-1"}`))
			return
		}
		w.Write([]byte(`{"state":"SUCCESS","expected_code_answer":[]}`))
	}))
	defer srv.Close()
	j := fastJudge(t, "")
	j.Base, j.HTTP = srv.URL, srv.Client()

	got, err := j.Expected(context.Background(), Auth{Session: "s", CSRF: "c"}, testQuestion(),
		[][]string{{"[1]", "2"}})
	if err != nil {
		t.Fatalf("케이스 단위 실패는 치명 오류가 아님: %v", err)
	}
	if got[0].Err == nil {
		t.Fatal("정답 없음이 케이스 에러로 안 잡힘")
	}
}

// 요청 헤더/바디 — 쿠키, csrf, data_input 조립.
func TestExpectedRequestShape(t *testing.T) {
	var gotCookie, gotCSRF, gotReferer string
	var gotBody struct {
		Lang       string `json:"lang"`
		QuestionID string `json:"question_id"`
		TypedCode  string `json:"typed_code"`
		DataInput  string `json:"data_input"`
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/problems/two-sum/interpret_solution/":
			gotCookie = r.Header.Get("Cookie")
			gotCSRF = r.Header.Get("x-csrftoken")
			gotReferer = r.Header.Get("Referer")
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			w.Write([]byte(`{"interpret_id":"id-1"}`))
		default:
			w.Write([]byte(`{"state":"SUCCESS","expected_code_answer":["[0,1]"]}`))
		}
	}))
	defer srv.Close()
	j := fastJudge(t, "")
	j.Base, j.HTTP = srv.URL, srv.Client()

	_, err := j.Expected(context.Background(), Auth{Session: "SESS", CSRF: "TK"}, testQuestion(),
		[][]string{{"[1,2]", "3"}})
	if err != nil {
		t.Fatal(err)
	}
	if gotCookie != "csrftoken=TK; LEETCODE_SESSION=SESS" {
		t.Fatalf("Cookie: %q", gotCookie)
	}
	if gotCSRF != "TK" {
		t.Fatalf("x-csrftoken: %q", gotCSRF)
	}
	if want := srv.URL + "/problems/two-sum/"; gotReferer != want {
		t.Fatalf("Referer: %q", gotReferer)
	}
	if gotBody.Lang != "cpp" || gotBody.QuestionID != "1" {
		t.Fatalf("lang/question_id: %+v", gotBody)
	}
	if gotBody.DataInput != "[1,2]\n3" {
		t.Fatalf("data_input: %q", gotBody.DataInput)
	}
	if gotBody.TypedCode == "" {
		t.Fatal("typed_code(stub) 비어 있음")
	}
}

func TestLoadAuth(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".lc-auth")

	if _, err := LoadAuth(root); err == nil {
		t.Fatal("설정 없는데 성공")
	}

	t.Setenv("LC_SESSION", "")
	t.Setenv("LC_CSRF", "")
	if err := os.WriteFile(path, []byte(`{"session":"s","csrf":"c"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	a, err := LoadAuth(root)
	if err != nil || a.Session != "s" || a.CSRF != "c" {
		t.Fatalf("0600 파일 로드: %+v %v", a, err)
	}

	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAuth(root); err == nil {
		t.Fatal("느슨한 권한인데 성공")
	}

	t.Setenv("LC_SESSION", "envs")
	t.Setenv("LC_CSRF", "envc")
	a, err = LoadAuth(root)
	if err != nil || a.Session != "envs" {
		t.Fatalf("env 우선: %+v %v", a, err)
	}
}

// 스텁이 없으면 네트워크 나가기 전에 실패.
func TestExpectedNoStub(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
	}))
	defer srv.Close()
	j := fastJudge(t, "")
	j.Base, j.HTTP = srv.URL, srv.Client()
	q := testQuestion()
	q.CodeSnippet = ""

	_, err := j.Expected(context.Background(), Auth{}, q, [][]string{{"1"}})
	if err == nil || hits.Load() != 0 {
		t.Fatalf("스텁 없음 에러여야 하고 요청도 없어야 함: %v %d", err, hits.Load())
	}
}
