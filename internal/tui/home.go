package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
			return func() tea.Msg { return openNewMsg{} }
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
	var b strings.Builder
	for _, w := range h.a.warnings() {
		b.WriteString(stBanner.Render("⚠ " + w))
		b.WriteByte('\n')
	}
	b.WriteString(h.duePane())
	b.WriteByte('\n')
	b.WriteString(h.weekLine())
	return b.String()
}

func (h *homeModel) duePane() string {
	due := h.a.dueList()
	recent := pslog.Recent(h.a.entries, 10)

	var left strings.Builder
	left.WriteString(stTitle.Render(fmt.Sprintf(" 오늘 재도전 (%d)", len(due))))
	left.WriteByte('\n')
	if len(due) == 0 {
		left.WriteString(stDim.Render(" 없음 · psdue 규칙: 마지막 행 기준 due<=오늘"))
	} else {
		left.WriteString(stDim.Render(" 재도전일  문제                 결과"))
		left.WriteByte('\n')
		for i, d := range due {
			site, name, _ := splitKey(d.Key)
			marker := " "
			style := stNormal
			if i == h.cursor {
				marker = ">"
				style = stSel
			}
			left.WriteString(style.Render(fmt.Sprintf("%s%s %-9s %-18s %s", marker, " ", d.Due, site+"/"+name, d.Result)))
			left.WriteByte('\n')
		}
	}

	var right strings.Builder
	right.WriteString(stTitle.Render(" 최근 기록"))
	right.WriteByte('\n')
	if len(recent) == 0 {
		right.WriteString(stDim.Render(" 기록 없음"))
	} else {
		for _, e := range recent {
			note := truncateStr(e.Note, 20)
			rs := stOK
			if e.Result != "ok" {
				rs = stFail
			}
			right.WriteString(fmt.Sprintf("%s %-16s %s %3d분 %s\n",
				e.Date, e.Site+"/"+e.Problem, rs.Render(string(e.Result)), e.Minutes, stDim.Render(note)))
		}
	}

	lp := padBlock(left.String(), h.rowsFor(len(due)+3))
	rp := padBlock(right.String(), h.rowsFor(len(recent)+1))
	return joinHorizontalPanes(lp, rp)
}

func (h *homeModel) rowsFor(rows int) int {
	max := h.a.height - 7
	if rows > max {
		return max
	}
	return rows
}

func (h *homeModel) weekLine() string {
	w := pslog.Week(h.a.entries, time.Now())
	return stDim.Render(fmt.Sprintf(" 이번 주: %d회 시도 · ok %d / fail %d / again %d · 평균 %d분",
		w.Solved, w.Ok, w.Fail, w.Again, w.AvgMinutes))
}

func (h *homeModel) helpText() string {
	return "Home: j/k 이동 · Enter 작업 열기 · t 재도전 · n 새 문제 · p 문제 목록 · l 로그 · s 통계 · r 새로고침 · q 종료"
}
