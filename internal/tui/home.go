package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"pstui/internal/ps/problem"
	"pstui/internal/ps/pslog"
)

type homeModel struct {
	a      *App
	cursor int
}

func newHome(a *App) *homeModel { return &homeModel{a: a} }

func (h *homeModel) update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.KeyMsg:
		due := h.a.dueList()
		switch m.String() {
		case "q":
			return tea.Quit
		case "n":
			return func() tea.Msg { return openPlanMsg{} }
		case "p":
			return func() tea.Msg { return openProblemsMsg{} }
		case "l":
			return func() tea.Msg { return openLogMsg{} }
		case "s":
			return func() tea.Msg { return openStatsMsg{} }
		case "r":
			return tea.Batch(loadLogCmd(h.a.cfg.Root), scanProblemsCmd(h.a.cfg.Root))
		case "up", "k":
			if h.cursor > 0 {
				h.cursor--
			}
		case "down", "j":
			if h.cursor < len(due)-1 {
				h.cursor++
			}
		case "enter":
			if p := h.selectedProblem(due); p != nil {
				p2 := *p
				return func() tea.Msg { return openWorkbenchMsg{Prob: p2} }
			}
		case "t":
			if p := h.selectedProblem(due); p != nil {
				p2 := *p
				return func() tea.Msg { return openRetryMsg{Prob: p2} }
			}
		}
	}
	return nil
}

func (h *homeModel) selectedProblem(due []pslog.DueItem) *problem.Problem {
	if len(due) == 0 {
		return nil
	}
	if h.cursor >= len(due) {
		h.cursor = len(due) - 1
	}
	d := due[h.cursor]
	site, name, ok := splitKey(d.Key)
	if !ok {
		return nil
	}
	return h.a.findProblem(site, name)
}

func (h *homeModel) view() string {
	a := h.a
	bodyH := a.height - 2 // 헤더 + 키바
	var top []string
	for _, w := range a.warnings() {
		top = append(top, stBanner.Render("⚠ "+w))
	}
	paneH := bodyH - len(top) - 1
	if paneH < 5 {
		paneH = 5
	}
	gap := 1
	leftW := a.width*55/100 - gap
	rightW := a.width - leftW - gap

	due := a.dueList()
	left := paneBox(fmt.Sprintf("오늘 재도전 · %d", len(due)), h.dueRows(due, leftW-4), leftW, paneH, true)
	right := paneBox("최근 기록", h.recentRows(rightW-4), rightW, paneH, false)

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, strings.Repeat(" ", gap), right)
	out := strings.Join(top, "\n")
	if out != "" {
		out += "\n"
	}
	return out + body + "\n" + h.weekLine()
}

// dueRows는 재도전 표 내용 (제목 제외, innerW는 패널 내부 폭).
func (h *homeModel) dueRows(due []pslog.DueItem, innerW int) string {
	if innerW < 20 {
		return ""
	}
	probW := innerW - 11 - 7
	var b strings.Builder
	b.WriteString(stDimmer.Render(padW("재도전", 11) + padW("문제", probW) + "결과"))
	b.WriteByte('\n')
	if len(due) == 0 {
		b.WriteString(stDim.Render("오늘은 없음 — n 으로 새 문제"))
		return b.String()
	}
	for i, d := range due {
		site, name, _ := splitKey(d.Key)
		prob := truncateStr(site+"/"+name, probW-1)
		res := resultStyle(d.Result).Render(padW(string(d.Result), 6))
		row := padW(shortDate(d.Due), 10) + " " + padW(prob, probW) + res
		b.WriteString(selRow(row, innerW, i == h.cursor))
		if i < len(due)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (h *homeModel) recentRows(innerW int) string {
	recent := pslog.Recent(h.a.entries, 10)
	if innerW < 24 {
		return ""
	}
	probW := innerW - 5 - 7 - 5
	var b strings.Builder
	b.WriteString(stDimmer.Render(padW("날짜", 6) + padW("문제", probW) + padW("결과", 7) + padW("분", 4)))
	b.WriteByte('\n')
	if len(recent) == 0 {
		b.WriteString(stDim.Render("기록 없음"))
		return b.String()
	}
	for _, e := range recent {
		prob := truncateStr(e.Site+"/"+e.Problem, probW-1)
		res := resultStyle(e.Result).Render(padW(string(e.Result), 7))
		b.WriteString(padW(shortDate(e.Date), 6) + padW(prob, probW) + res + padW(fmt.Sprintf("%d", e.Minutes), 4))
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func (h *homeModel) weekLine() string {
	w := pslog.Week(h.a.entries, time.Now())
	return " " + stDim.Render(fmt.Sprintf("이번 주 %d회 · ", w.Solved)) +
		stOK.Render(fmt.Sprintf("ok %d", w.Ok)) + stDim.Render(" / ") +
		stFail.Render(fmt.Sprintf("fail %d", w.Fail)) + stDim.Render(" / ") +
		stWarn.Render(fmt.Sprintf("again %d", w.Again)) +
		stDim.Render(fmt.Sprintf(" · 평균 %d분", w.AvgMinutes))
}

func (h *homeModel) helpText() string {
	return keyBar(
		[2]string{"j/k", "이동"}, [2]string{"⏎", "열기"}, [2]string{"t", "재도전"},
		[2]string{"n", "플랜"}, [2]string{"p", "문제"}, [2]string{"l", "로그"},
		[2]string{"s", "통계"}, [2]string{"r", "새로고침"}, [2]string{"q", "종료"},
	)
}

// shortDate는 "2026-11-14" → "11-14".
func shortDate(d string) string {
	if len(d) == 10 {
		return d[5:]
	}
	return d
}

func resultStyle(r pslog.Result) lipgloss.Style {
	switch r {
	case "ok":
		return stOK
	case "again":
		return stWarn
	default:
		return stFail
	}
}
