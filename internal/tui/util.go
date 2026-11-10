package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// splitKey는 psdue 키 "site/problem"을 나눈다.
func splitKey(key string) (site, name string, ok bool) {
	i := strings.IndexByte(key, '/')
	if i <= 0 || i == len(key)-1 {
		return "", "", false
	}
	return key[:i], key[i+1:], true
}

// padBlock은 문자열을 지정 줄 수에 맞게 빈 줄로 채운다.
func padBlock(s string, rows int) string {
	lines := strings.Split(strings.TrimSuffix(s, "\n"), "\n")
	for len(lines) < rows {
		lines = append(lines, "")
	}
	if len(lines) > rows && rows > 0 {
		lines = lines[:rows]
	}
	return strings.Join(lines, "\n")
}

func joinHorizontalPanes(left, right string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

// overlayHelp는 본문 위에 도움말 상자를 겹쳐 그린다.
func overlayHelp(body, text string) string {
	lines := strings.Split(body, "\n")
	box := stBorder.Render(stKeyHelp.Render(" " + text + " "))
	boxLines := strings.Split(box, "\n")
	y := len(lines)/2 - len(boxLines)/2
	if y < 0 {
		y = 0
	}
	for i, bl := range boxLines {
		if y+i >= len(lines) {
			break
		}
		lines[y+i] = overlayLine(lines[y+i], bl)
	}
	return strings.Join(lines, "\n")
}

// overlayLine은 한 줄 위에 다른 줄을 왼쪽 정렬로 덮어쓴다 (터미널 폭 초과 방지).
func overlayLine(base, top string) string {
	w := lipgloss.Width(base)
	tw := lipgloss.Width(top)
	if tw >= w {
		return top
	}
	return top + base[tw:]
}

func truncateStr(s string, n int) string {
	if lipgloss.Width(s) <= n {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r))+1 > n {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}

// isoWeekKey는 "2006-01-02"를 ISO 주(월요일 시작) 키로 바꾼다.
func isoWeekKey(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "????-??"
	}
	y, w := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", y, w)
}
