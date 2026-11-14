package stmt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"leetui/internal/ps/lcfmt"
	"leetui/internal/ps/lcref"
	"leetui/internal/ps/problem"
)

// loadFixture는 녹화해 둔 GraphQL 응답(../lcref/testdata)을 Question으로 만든다.
func loadFixture(t *testing.T, slug string) lcref.Question {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "lcref", "testdata", slug+".json"))
	if err != nil {
		t.Skipf("픽스처 없음: %v", err)
	}
	var out struct {
		Data struct {
			Question struct {
				QuestionID          string   `json:"questionId"`
				FrontendID          string   `json:"questionFrontendId"`
				Title               string   `json:"title"`
				Difficulty          string   `json:"difficulty"`
				Content             string   `json:"content"`
				ExampleTestcaseList []string `json:"exampleTestcaseList"`
				MetaData            string   `json:"metaData"`
				CodeSnippets        []struct {
					LangSlug string `json:"langSlug"`
					Code     string `json:"code"`
				} `json:"codeSnippets"`
			} `json:"question"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	rq := out.Data.Question
	q := lcref.Question{
		Slug: slug, QuestionID: rq.QuestionID, FrontendID: rq.FrontendID,
		Title: rq.Title, Difficulty: rq.Difficulty, Content: rq.Content,
		Examples: rq.ExampleTestcaseList, MetaData: rq.MetaData,
	}
	for _, s := range rq.CodeSnippets {
		if s.LangSlug == "cpp" {
			q.CodeSnippet = s.Code
		}
	}
	return q
}

func setupProblem(t *testing.T, root, name string) problem.Problem {
	t.Helper()
	dir, _, err := problem.New(root, "lc", "0001", name, filepath.Join(root, "template.cpp"), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return problem.Problem{Site: "lc", Name: "0001-" + name, Dir: dir}
}

// 녹화 픽스처로 실제 문제 4종이 실행 케이스로 변환되는지 확인
// (완료 기준 문제 중 오라클 없이 검증 가능한 부분).
func TestFillFromQuestionRealFixtures(t *testing.T) {
	cases := []struct {
		slug     string
		wantIn   string // ex1.in
		wantOut  string // ex1.out
		wantN    int    // 예제 수
		fallback string // 폴백이면 사유 키워드
	}{
		{"two-sum", "4\n2 7 11 15\n9\n", "0 1\n", 3, ""},
		{"merge-intervals", "4\n2 1 3\n2 2 6\n2 8 10\n2 15 18\n", "1 6\n8 10\n15 18\n", 3, ""},
		{"number-of-islands", "4 5\n11110\n11010\n11000\n00000\n", "1\n", 2, ""},
		{"reverse-linked-list", "5\n1 2 3 4 5\n", "5 4 3 2 1\n", 3, ""},
		{"lru-cache", "", "", 0, "design"}, // design 지문은 Input: 형식이 아니라 예제 0개 → 폴백
	}
	for _, c := range cases {
		t.Run(c.slug, func(t *testing.T) {
			root := t.TempDir()
			writeTemplate(t, root)
			p := setupProblem(t, root, c.slug)
			q := loadFixture(t, c.slug)
			res, err := fillFromQuestion(root, "lc", c.slug, p, q)
			if err != nil {
				t.Fatal(err)
			}
			if res.Examples != c.wantN {
				t.Fatalf("예제 수 %d, want %d", res.Examples, c.wantN)
			}
			if c.fallback != "" {
				if !res.Fallback {
					t.Fatalf("폴백이어야 함 (reason=%q)", res.Reason)
				}
				if !strings.Contains(res.Reason, c.fallback) {
					t.Fatalf("폴백 사유 %q에 %q 없음", res.Reason, c.fallback)
				}
				if _, err := os.Stat(filepath.Join(p.Dir, "ex1.in")); err == nil {
					t.Fatal("폴백인데 ex1.in이 생성됨")
				}
				if _, err := os.Stat(filepath.Join(p.Dir, ".ex1.in")); err == nil {
					t.Fatal("design 문제는 예제 파싱이 0개 — .ex1.in도 없어야 함")
				}
				return
			}
			if res.Fallback {
				t.Fatalf("지원 문제인데 폴백: %s", res.Reason)
			}
			if res.Cases != c.wantN*2 {
				t.Fatalf("만든 파일 수 %d, want %d", res.Cases, c.wantN*2)
			}
			gotIn := readFile(t, filepath.Join(p.Dir, "ex1.in"))
			if gotIn != c.wantIn {
				t.Fatalf("ex1.in:\n got %q\nwant %q", gotIn, c.wantIn)
			}
			gotOut := readFile(t, filepath.Join(p.Dir, "ex1.out"))
			if gotOut != c.wantOut {
				t.Fatalf("ex1.out:\n got %q\nwant %q", gotOut, c.wantOut)
			}
			// 지문 md에 형식 섹션
			md := readFile(t, MdPath(root, "lc", c.slug))
			if !strings.Contains(md, "## 입력 형식 (leetui)") {
				t.Fatalf("지문에 형식 섹션 없음")
			}
		})
	}
}

// 이미 있는 exN.*은 보존하고, 재실행은 나머지만 채운다.
func TestFillPreservesExisting(t *testing.T) {
	root := t.TempDir()
	writeTemplate(t, root)
	p := setupProblem(t, root, "two-sum")
	q := loadFixture(t, "two-sum")
	if err := os.WriteFile(filepath.Join(p.Dir, "ex1.in"), []byte("내가 고친 입력\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := fillFromQuestion(root, "lc", "two-sum", p, q)
	if err != nil {
		t.Fatal(err)
	}
	if res.Skipped != 1 || res.Cases != 5 {
		t.Fatalf("보존/신규 = %d/%d, want 1/5", res.Skipped, res.Cases)
	}
	if got := readFile(t, filepath.Join(p.Dir, "ex1.in")); got != "내가 고친 입력\n" {
		t.Fatalf("기존 파일이 덮어써짐: %q", got)
	}
	// 섹션은 두 번 넣지 않는다
	res2, err := fillFromQuestion(root, "lc", "two-sum", p, q)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Cases != 0 || res2.Skipped != 6 {
		t.Fatalf("재실행 보존 = %d/%d, want 0/6", res2.Cases, res2.Skipped)
	}
	md := readFile(t, MdPath(root, "lc", "two-sum"))
	if n := strings.Count(md, "## 입력 형식 (leetui)"); n != 1 {
		t.Fatalf("섹션이 %d번 삽입됨", n)
	}
}

// 기존 지문 md(형식 섹션 없음)에도 삽입한다.
func TestEnsureSpecInsertsIntoExistingMd(t *testing.T) {
	root := t.TempDir()
	slug := "two-sum"
	q := loadFixture(t, slug)
	md := MdPath(root, "lc", slug)
	if err := os.MkdirAll(filepath.Dir(md), 0o755); err != nil {
		t.Fatal(err)
	}
	old := "# Two Sum · Easy\n\nhttps://leetcode.com/problems/two-sum/\n\n본문 시작.\n"
	if err := os.WriteFile(md, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	meta, err := lcfmt.ParseMeta(q.MetaData)
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureSpec(root, "lc", slug, q, meta); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, md)
	if !strings.Contains(got, "## 입력 형식 (leetui)") || !strings.Contains(got, "본문 시작.") {
		t.Fatalf("삽입 결과 이상:\n%s", got)
	}
	if !strings.HasPrefix(got, "# ") {
		t.Fatalf("제목이 맨 앞이 아님:\n%.40s", got)
	}
}

func writeTemplate(t *testing.T, root string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "template.cpp"), []byte("int main(){}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
