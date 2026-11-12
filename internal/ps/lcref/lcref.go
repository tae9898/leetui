// Package lcref는 문제 참조(URL, "56. Merge Intervals", 슬러그)를 파싱해
// psnew에 필요한 site/id/name으로 바꾼다. 번호는 LeetCode GraphQL로도 조회한다.
package lcref

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Ref struct {
	Site string // "lc" | "hr" | "" (미정)
	ID   string // "0056" 형태, 모르면 ""
	Name string // kebab-case 슬러그
}

var (
	reLCURL = regexp.MustCompile(`leetcode\.[a-z.]+/problems/([a-z0-9-]+)`)
	reHRURL = regexp.MustCompile(`hackerrank\.[a-z.]+/challenges/([a-z0-9-]+)`)
	reTitle = regexp.MustCompile(`^\s*(\d{1,4})[.\s)]+\s*(.+?)\s*$`)
)

// Parse는 자유 형식 입력을 Ref로 바꾼다. 인식 실패 시 ok=false.
// 인식 가능 형식:
//   - https://leetcode.com/problems/merge-intervals/
//   - https://www.hackerrank.com/challenges/repeated-string/problem
//   - 56. Merge Intervals / 56 Merge Intervals / 56) merge intervals
//   - merge-intervals (슬러그 단독)
func Parse(input string) (Ref, bool) {
	s := strings.TrimSpace(input)
	if s == "" {
		return Ref{}, false
	}
	if m := reLCURL.FindStringSubmatch(s); m != nil {
		return Ref{Site: "lc", Name: m[1]}, true
	}
	if m := reHRURL.FindStringSubmatch(s); m != nil {
		return Ref{Site: "hr", Name: m[1]}, true
	}
	if m := reTitle.FindStringSubmatch(s); m != nil {
		n, _ := strconv.Atoi(m[1])
		return Ref{Site: "lc", ID: pad4(n), Name: Slugify(m[2])}, true
	}
	if isSlug(s) {
		return Ref{Name: Slugify(s)}, true
	}
	return Ref{}, false
}

func isSlug(s string) bool {
	if strings.ContainsAny(s, " ./") {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-') {
			return false
		}
	}
	return len(s) >= 3
}

// Slugify는 "Merge Intervals" → "merge-intervals".
func Slugify(title string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(title)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	out := b.String()
	for strings.Contains(out, "--") {
		out = strings.ReplaceAll(out, "--", "-")
	}
	return strings.Trim(out, "-")
}

func pad4(n int) string {
	s := strconv.Itoa(n)
	for len(s) < 4 {
		s = "0" + s
	}
	return s
}

// LookupNumber는 슬러그로 LeetCode 문제 번호를 조회한다 ("56" 형태 반환).
func LookupNumber(ctx context.Context, slug string) (string, error) {
	body := fmt.Sprintf(`{"query":"query($s:String!){question(titleSlug:$s){questionId}}","variables":{"s":%q}}`, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://leetcode.com/graphql", strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://leetcode.com/")

	cctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	req = req.WithContext(cctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		Data struct {
			Question struct {
				QuestionID string `json:"questionId"`
			} `json:"question"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Data.Question.QuestionID == "" {
		return "", fmt.Errorf("leetcode에서 %s 조회 실패", slug)
	}
	return out.Data.Question.QuestionID, nil
}

// PadNumber은 조회된 번호를 psnew 규약(4자리)으로 맞춘다.
func PadNumber(n string) string {
	if v, err := strconv.Atoi(n); err == nil {
		return pad4(v)
	}
	return n
}

type Statement struct {
	Title      string
	Difficulty string
	BodyHTML   string
}

// FetchStatement는 지문(제목·난이도·본문 HTML)을 가져온다.
func FetchStatement(ctx context.Context, slug string) (Statement, error) {
	body := fmt.Sprintf(`{"query":"query($s:String!){question(titleSlug:$s){title difficulty content}}","variables":{"s":%q}}`, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://leetcode.com/graphql", strings.NewReader(body))
	if err != nil {
		return Statement{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	req = req.WithContext(cctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Statement{}, err
	}
	defer resp.Body.Close()
	var out struct {
		Data struct {
			Question struct {
				Title      string `json:"title"`
				Difficulty string `json:"difficulty"`
				Content    string `json:"content"`
			} `json:"question"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Statement{}, err
	}
	q := out.Data.Question
	if q.Content == "" {
		return Statement{}, fmt.Errorf("지문 없음 (유료 문제일 수 있음): %s", slug)
	}
	return Statement{Title: q.Title, Difficulty: q.Difficulty, BodyHTML: q.Content}, nil
}

// Example은 지문 예제의 입출력 텍스트.
type Example struct {
	In  string
	Out string
}

// ParseExamples는 지문 본문(HTML)에서 "Input: ... / Output: ..." 쌍을 뽑는다.
// <pre> 블록 형식과 <p> 문단 형식(LC가 섞어 씀) 모두 처리한다.
func ParseExamples(bodyHTML string) []Example {
	var out []Example
	var in, outBuf strings.Builder
	mode := 0 // 0=대기, 1=in 수집, 2=out 수집
	flush := func() {
		if mode == 2 && outBuf.Len() > 0 {
			out = append(out, Example{In: strings.TrimSpace(in.String()), Out: strings.TrimSpace(outBuf.String())})
		}
		in.Reset()
		outBuf.Reset()
		mode = 0
	}
	for _, line := range strings.Split(HTMLToText(bodyHTML), "\n") {
		t := strings.TrimSpace(line)
		lt := strings.ToLower(t)
		switch {
		case strings.HasPrefix(lt, "input:"):
			flush()
			mode = 1
			in.WriteString(strings.TrimSpace(t[len("input:"):]))
			in.WriteByte('\n')
		case strings.HasPrefix(lt, "output:"):
			if mode == 0 {
				continue
			}
			mode = 2
			outBuf.WriteString(strings.TrimSpace(t[len("output:"):]))
			outBuf.WriteByte('\n')
		case strings.HasPrefix(lt, "explanation:"), strings.HasPrefix(lt, "constraint"),
			strings.HasPrefix(lt, "note:"), strings.HasPrefix(lt, "follow up"),
			strings.HasPrefix(lt, "example "), lt == "":
			// 문단 구분 — Output 뒤에서만 예제 확정
			if mode == 2 && lt == "" {
				flush()
			} else if mode == 2 && (strings.HasPrefix(lt, "explanation:") || strings.HasPrefix(lt, "constraint") ||
				strings.HasPrefix(lt, "note:") || strings.HasPrefix(lt, "follow up") || strings.HasPrefix(lt, "example ")) {
				flush()
			}
		default:
			if mode == 1 {
				in.WriteString(t)
				in.WriteByte('\n')
			} else if mode == 2 {
				outBuf.WriteString(t)
				outBuf.WriteByte('\n')
			}
		}
	}
	flush()
	return out
}

var (
	reBr     = regexp.MustCompile(`(?i)<br\s*/?>`)
	reBlock  = regexp.MustCompile(`(?i)</?(p|div|pre|ul|ol|li|h[1-6]|table|tr|blockquote)\b[^>]*>`)
	reAnyTag = regexp.MustCompile(`(?s)<[^>]+>`)
)

// HTMLToText는 LC 지문 HTML을 터미널 텍스트로 바꾼다 (예제 <pre>는 개행 보존).
func HTMLToText(html string) string {
	s := reBr.ReplaceAllString(html, "\n")
	s = reBlock.ReplaceAllStringFunc(s, func(tag string) string {
		l := strings.ToLower(tag)
		if strings.HasPrefix(l, "<li") {
			return "\n  • "
		}
		return "\n"
	})
	s = reAnyTag.ReplaceAllString(s, "")
	s = decodeEntities(s)
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	return strings.TrimSpace(s)
}

func decodeEntities(s string) string {
	return strings.NewReplacer(
		"&nbsp;", " ", "&lt;", "<", "&gt;", ">", "&quot;", `"`,
		"&#39;", "'", "&apos;", "'", "&amp;", "&",
	).Replace(s)
}

var (
	rePre     = regexp.MustCompile(`(?is)<pre[^>]*>(.*?)</pre>`)
	reLink    = regexp.MustCompile(`(?is)<a\s[^>]*href="([^"]*)"[^>]*>(.*?)</a>`)
	reImg     = regexp.MustCompile(`(?is)<img\s[^>]*src="([^"]*)"[^>]*>`)
	reStrong  = regexp.MustCompile(`(?is)</?(strong|b)\b[^>]*>`)
	reEm      = regexp.MustCompile(`(?is)</?(em|i)\b[^>]*>`)
	reCode    = regexp.MustCompile(`(?is)</?code\b[^>]*>`)
	reSub     = regexp.MustCompile(`(?is)</?sub\b[^>]*>`)
	reSupOpen = regexp.MustCompile(`(?is)<sup\b[^>]*>`)
	reSupClos = regexp.MustCompile(`(?is)</sup>`)
	reHeading = regexp.MustCompile(`(?is)<h([1-6])\b[^>]*>`)
	reUlOl    = regexp.MustCompile(`(?is)</?(ul|ol)\b[^>]*>`)
	reLiOpen  = regexp.MustCompile(`(?is)<li\b[^>]*>`)
	reLiClos  = regexp.MustCompile(`(?is)</li>`)
	rePDiv    = regexp.MustCompile(`(?is)</?(p|div|blockquote|table|tr)\b[^>]*>`)
)

// HTMLToMarkdown은 LC 지문 HTML을 마크다운으로 바꾼다.
// <pre>는 펜스 코드 블록, strong/em/code는 **/*/`, <li>는 - 목록, <sup>는 ^.
func HTMLToMarkdown(html string) string {
	// <pre>는 코드 블록으로 통째로 보존 — 인라인 변환이 내부를 건드리지 않게 먼저 뺀다
	var pres []string
	s := rePre.ReplaceAllStringFunc(html, func(m string) string {
		code := reAnyTag.ReplaceAllString(rePre.FindStringSubmatch(m)[1], "")
		code = strings.TrimSpace(decodeEntities(code))
		pres = append(pres, "```\n"+code+"\n```")
		return "\x00" + strconv.Itoa(len(pres)-1) + "\x00"
	})

	s = reLink.ReplaceAllString(s, "[$2]($1)")
	s = reImg.ReplaceAllString(s, "![img]($1)")
	s = reStrong.ReplaceAllString(s, "**")
	s = reEm.ReplaceAllString(s, "*")
	s = reCode.ReplaceAllString(s, "`")
	s = reSub.ReplaceAllString(s, "_")
	s = reSupOpen.ReplaceAllString(s, "^")
	s = reSupClos.ReplaceAllString(s, "")

	s = reBr.ReplaceAllString(s, "\n")
	s = reHeading.ReplaceAllStringFunc(s, func(m string) string {
		n := m[2] - '0'
		return "\n" + strings.Repeat("#", int(n)) + " "
	})
	s = rePDiv.ReplaceAllString(s, "\n\n")
	s = reUlOl.ReplaceAllString(s, "\n")
	s = reLiOpen.ReplaceAllString(s, "\n- ")
	s = reLiClos.ReplaceAllString(s, "")
	s = reAnyTag.ReplaceAllString(s, "")
	s = decodeEntities(s)
	s = strings.ReplaceAll(s, "\u00a0", " ") // non-breaking space
	s = strings.ReplaceAll(s, "\u200b", "")  // zero-width space (LC가 종종 심음)
	s = strings.ReplaceAll(s, "\ufeff", "")

	// 줄별 정리 + 빈 줄 축소
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	s = strings.Join(lines, "\n")
	for strings.Contains(s, "\n\n\n") {
		s = strings.ReplaceAll(s, "\n\n\n", "\n\n")
	}
	for strings.Contains(s, "\n\n- ") { // 목록 항목 사이 빈 줄 제거
		s = strings.ReplaceAll(s, "\n\n- ", "\n- ")
	}
	s = strings.TrimSpace(s)

	// 코드 블록 복원
	for i, p := range pres {
		s = strings.Replace(s, "\x00"+strconv.Itoa(i)+"\x00", p, 1)
	}
	return s
}
