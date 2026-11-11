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

// helpCard는 화면 중앙에 띄우는 도움말 카드.
func helpCard(text string, width, height int) string {
	lines := strings.Count(text, "\n") + 1
	w := width - 8
	if w > 64 {
		w = 64
	}
	h := lines + 4
	if h > height-4 {
		h = height - 4
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
		paneBox("도움말", text, w, h, true))
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
