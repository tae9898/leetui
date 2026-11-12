// Package stmt은 지문 캐시와 예제 채움을 제공한다.
// 지문은 PS_ROOT/.stmt/<site>/<slug>.md로 캐시하고 편집기(nvim)로 열어 본다.
package stmt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"leetui/internal/ps/lcref"
	"leetui/internal/ps/problem"
)

func MdPath(root, site, slug string) string {
	return filepath.Join(root, ".stmt", site, slug+".md")
}

func htmlPath(root, site, slug string) string {
	return filepath.Join(root, ".stmt", site, slug+".html")
}

// Ensure는 지문 마크다운 파일 경로를 반환한다. 캐시가 없으면 만든다(lc는 fetch 필요).
func Ensure(ctx context.Context, root, site, slug string) (string, error) {
	md := MdPath(root, site, slug)
	if data, err := os.ReadFile(md); err == nil && len(data) > 0 {
		return md, nil
	}
	var text, html string
	switch site {
	case "lc":
		st, err := lcref.FetchStatement(ctx, slug)
		if err != nil {
			return "", err
		}
		html = st.BodyHTML
		text = render(st.Title, st.Difficulty, "https://leetcode.com/problems/"+slug+"/", lcref.HTMLToMarkdown(st.BodyHTML))
	case "hr":
		url := "https://www.hackerrank.com/challenges/" + slug + "/problem"
		text = "HackerRank 지문은 웹에서 보세요.\n\n" + url + "\n\n(슬러그는 표시명 기준 추정 — 열리지 않으면 사이트에서 검색)"
	default: // task 등 내장 지문
		title, body, _ := TaskStatement(slug)
		text = render(title, "", "", body)
	}
	if err := os.MkdirAll(filepath.Dir(md), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(md, []byte(text), 0644); err != nil {
		return "", err
	}
	_ = os.Remove(strings.TrimSuffix(md, ".md") + ".txt") // 예전 평문 캐시 정리
	if html != "" {
		_ = os.WriteFile(htmlPath(root, site, slug), []byte(html), 0644)
	}
	return md, nil
}

func render(title, diff, url, body string) string {
	var b strings.Builder
	if title != "" {
		b.WriteString("# " + title)
		if diff != "" {
			b.WriteString(" · " + diff)
		}
		b.WriteString("\n")
	}
	if url != "" {
		b.WriteString("\n" + url + "\n")
	}
	if title != "" || url != "" {
		b.WriteString("\n")
	}
	b.WriteString(body)
	if !strings.HasSuffix(body, "\n") {
		b.WriteString("\n")
	}
	return b.String()
}

// Fill은 예제를 숨김 스펙 파일(.exN.in/.out)로 저장한다.
// HackerRank 방식 연습 규칙(모든 문제 stdin/stdout)에 따라 실행 케이스(1.in…)는
// 사용자가 자기 포맷으로 직접 쓰고, LC 예제는 참고용 스펙으로만 남긴다.
// 저장한 예제 개수를 반환한다.
func Fill(ctx context.Context, root, slug string, p problem.Problem, template string, now time.Time) (int, error) {
	if p.Site != "lc" {
		return 0, fmt.Errorf("예제 자동 채움은 LeetCode만 지원")
	}
	if _, err := os.Stat(p.Dir); err != nil {
		if template == "" {
			return 0, fmt.Errorf("template.cpp 없음 — 먼저 ⏎ 로 문제를 시작하세요")
		}
		id, name := SplitName(p.Name, slug)
		if _, _, err := problem.New(root, p.Site, id, name, template, now); err != nil {
			return 0, err
		}
	}
	html, err := ensureHTML(ctx, root, p.Site, slug)
	if err != nil {
		return 0, err
	}
	ex := lcref.ParseExamples(html)
	if len(ex) == 0 {
		return 0, fmt.Errorf("지문에서 예제를 찾지 못함 — rm %s 후 재시도", htmlPath(root, p.Site, slug))
	}
	for i, e := range ex {
		_ = os.WriteFile(filepath.Join(p.Dir, fmt.Sprintf(".ex%d.in", i+1)), []byte(e.In+"\n"), 0644)
		_ = os.WriteFile(filepath.Join(p.Dir, fmt.Sprintf(".ex%d.out", i+1)), []byte(e.Out+"\n"), 0644)
	}
	return len(ex), nil
}

func ensureHTML(ctx context.Context, root, site, slug string) (string, error) {
	hp := htmlPath(root, site, slug)
	if data, err := os.ReadFile(hp); err == nil && len(data) > 0 {
		return string(data), nil
	}
	st, err := lcref.FetchStatement(ctx, slug)
	if err != nil {
		return "", err
	}
	_ = os.MkdirAll(filepath.Dir(hp), 0755)
	_ = os.WriteFile(hp, []byte(st.BodyHTML), 0644)
	return st.BodyHTML, nil
}

// SplitName은 "0001-two-sum" → ("0001", "two-sum").
func SplitName(full, fallback string) (id, name string) {
	if i := strings.IndexByte(full, '-'); i > 0 {
		return full[:i], full[i+1:]
	}
	return "0000", fallback
}

// TaskStatement는 plan.md 7절의 과제형 요구사항을 마크다운으로 반환한다.
func TaskStatement(slug string) (string, string, bool) {
	common := "\n\n## 규칙 (plan.md 7절)\n\n- 40분 제한, assert 기반 테스트를 같은 파일에 작성\n- sanitizer를 켠 상태로 통과\n- C와 C++ 두 버전으로 한 번씩"
	switch slug {
	case "ring-buffer":
		return "링버퍼", "## 요구사항\n\n- 고정 크기, push/pop\n- 가득 찼을 때 정책 선택: 거부 또는 덮어쓰기\n\n## 확인 포인트\n\n- full과 empty 구분\n- 인덱스 wrap\n- 단일 생산자-단일 소비자에서의 스레드 안전성 설명" + common, true
	case "memory-pool":
		return "고정 블록 메모리 풀", "## 요구사항\n\n- 정적 배열 위에서 alloc/free\n- 동적 할당 없음\n\n## 확인 포인트\n\n- free list 구성\n- 이중 해제와 범위 밖 포인터 검출\n- 정렬(alignment)" + common, true
	case "protocol-parser":
		return "프로토콜 파서", "## 요구사항\n\n- 헤더 · 길이 · 페이로드 · 체크섬 구조의 바이트 스트림 파싱\n\n## 확인 포인트\n\n- 상태머신 구성\n- 조각난 입력 처리\n- 길이 필드 검증\n- 엔디안" + common, true
	}
	return "과제", "plan.md 7절 참고", true
}
