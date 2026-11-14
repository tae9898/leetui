package lcref

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in             string
		site, id, name string
		ok             bool
	}{
		{"https://leetcode.com/problems/merge-intervals/", "lc", "", "merge-intervals", true},
		{"https://leetcode.cn/problems/two-sum/description", "lc", "", "two-sum", true},
		{"https://www.hackerrank.com/challenges/repeated-string/problem", "hr", "", "repeated-string", true},
		{"56. Merge Intervals", "lc", "0056", "merge-intervals", true},
		{"56 merge intervals", "lc", "0056", "merge-intervals", true},
		{"1) Two Sum", "lc", "0001", "two-sum", true},
		{"merge-intervals", "", "", "merge-intervals", true},
		{"", "", "", "", false},
		{"not a ref !!", "", "", "", false},
	}
	for _, c := range cases {
		got, ok := Parse(c.in)
		if ok != c.ok || got.Site != c.site || got.ID != c.id || got.Name != c.name {
			t.Errorf("Parse(%q) = %+v ok=%v, want site=%q id=%q name=%q ok=%v",
				c.in, got, ok, c.site, c.id, c.name, c.ok)
		}
	}
}

func TestSlugify(t *testing.T) {
	cases := [][2]string{
		{"Merge Intervals", "merge-intervals"},
		{"String to Integer (atoi)", "string-to-integer-atoi"},
		{"  Best  Time   ", "best-time"},
		{"3Sum", "3sum"},
	}
	for _, c := range cases {
		if got := Slugify(c[0]); got != c[1] {
			t.Errorf("Slugify(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestLookupNumber(t *testing.T) {
	if testing.Short() {
		t.Skip("네트워크 조회는 -short에서 스킵")
	}
	n, err := LookupNumber(context.Background(), "merge-intervals")
	if err != nil {
		t.Skipf("네트워크 실패: %v", err)
	}
	if PadNumber(n) != "0056" {
		t.Fatalf("merge-intervals 번호 = %q, want 56", n)
	}
}

// fixtureServer는 녹화해 둔 GraphQL 응답(testdata/<slug>.json)을 제공한다.
func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Variables struct {
				Slug string `json:"s"`
			} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", 400)
			return
		}
		data, err := os.ReadFile(filepath.Join("testdata", req.Variables.Slug+".json"))
		if err != nil {
			// 녹화되지 않은 슬러그는 question:null (유료/존재 않음 응답과 같은 형태)
			w.Write([]byte(`{"data":{"question":null}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestFetchQuestion(t *testing.T) {
	srv := fixtureServer(t)
	q, err := fetchQuestion(context.Background(), srv.URL, srv.Client(), "two-sum")
	if err != nil {
		t.Fatal(err)
	}
	if q.Title != "Two Sum" || q.Difficulty != "Easy" {
		t.Fatalf("제목/난이도 불일치: %+v", q)
	}
	if q.QuestionID != "1" || q.FrontendID != "1" {
		t.Fatalf("ID 불일치: %+v", q)
	}
	if len(q.Examples) != 3 || q.Examples[0] != "[2,7,11,15]\n9" {
		t.Fatalf("예제 케이스 불일치: %q", q.Examples)
	}
	if !strings.Contains(q.MetaData, `"twoSum"`) {
		t.Fatalf("metaData 없음: %q", q.MetaData)
	}
	if !strings.HasPrefix(q.CodeSnippet, "class Solution") {
		t.Fatalf("cpp 스텁 없음: %q", q.CodeSnippet)
	}
	if !strings.Contains(q.Content, "Input:") {
		t.Fatalf("지문 본문 없음")
	}
}

func TestFetchQuestionNoContent(t *testing.T) {
	srv := fixtureServer(t)
	if _, err := fetchQuestion(context.Background(), srv.URL, srv.Client(), "not-recorded"); err == nil {
		t.Fatal("question:null인데 에러가 없음")
	}
}
