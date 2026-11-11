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
