package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// padW는 표시 폭 기준 오른쪽 패딩 (CJK 안전).
func padW(s string, w int) string {
	d := w - lipgloss.Width(s)
	if d <= 0 {
		return s
	}
	return s + strings.Repeat(" ", d)
}

// paneBox는 둥근 테두리 패널을 정확한 w×h로 그린다.
// 첫 줄에 제목, 둘째 줄에 구분선, 이후 content. focused가 참이면 테두리가 밝아진다.
func paneBox(title string, content string, w, h int, focused bool) string {
	if w < 8 || h < 4 {
		return strings.Repeat("·", w)
	}
	border := lipgloss.RoundedBorder()
	color := cBorderDim
	if focused {
		color = cBorderOn
	}
	inner := lipgloss.NewStyle().
		Width(w - 2).     // 테두리 2칸 제외
		MaxHeight(h - 2). // 잘림 방지
		PaddingLeft(1)    // 제목/행 왼쪽 숨
	return lipgloss.NewStyle().
		Border(border).
		BorderForeground(color).
		Width(w - 2).
		Height(h - 2).
		Render(inner.Render(titleLine(title) + "\n" + content))
}

func titleLine(title string) string {
	return stTitle.Render(title)
}

// selRow는 선택 행을 패널 내부 폭만큼 배경 하이라이트로 그린다.
func selRow(text string, innerW int, selected bool) string {
	if selected {
		return lipgloss.NewStyle().
			Width(innerW).Background(cSelBG).Foreground(cSelFG).Bold(true).
			Render(text)
	}
	return padW(text, innerW)
}

// keyBar는 하단 키 안내 바. 쌍 (키, 설명)을 균일 간격으로 나열한다.
func keyBar(pairs ...[2]string) string {
	var parts []string
	for _, p := range pairs {
		parts = append(parts, stKey.Render(p[0])+stDim.Render(" "+p[1]))
	}
	return strings.Join(parts, stDimmer.Render(" │ "))
}

// header는 상단 바: 좌측 앱 이름·화면, 우측 루트·날짜.
func header(left, right string, width int) string {
	return padW(" "+left, width-lipgloss.Width(right)-1) + right
}

// vjoin은 세로로 쌓되 빈 줄 없이 정확히 채운다.
func fitLines(s string, h int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < h {
		lines = append(lines, "")
	}
	if len(lines) > h && h > 0 {
		lines = lines[:h]
	}
	return strings.Join(lines, "\n")
}

// fitWidth는 각 줄을 표시 폭 w로 절단한다 (ANSI 시퀀스는 유지).
func fitWidth(s string, w int) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if lipgloss.Width(l) <= w {
			continue
		}
		var b strings.Builder
		cur := 0
		for _, r := range l {
			rw := lipgloss.Width(string(r))
			if cur+rw > w-1 {
				break
			}
			b.WriteRune(r)
			cur += rw
		}
		lines[i] = b.String() + "…"
	}
	return strings.Join(lines, "\n")
}
