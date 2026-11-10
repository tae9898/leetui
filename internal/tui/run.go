package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"pstui/internal/ps/linediff"
	"pstui/internal/ps/problem"
	"pstui/internal/ps/runner"
)

type runModel struct {
	a    *App
	prob problem.Problem
	fast bool

	phase      int // 1 컴파일, 2 케이스 실행, 3 완료, 4 취소됨
	compileOut string
	cases      []runner.Case
	results    []runner.CaseResult
	cur        int
	cursor     int
	tab        int // 0 diff, 1 stderr, 2 컴파일

	ctx    context.Context
	cancel context.CancelFunc
	vp     viewport.Model
	vpOK   bool

	fastElapsed string
}

func newRun(a *App) *runModel {
	return &runModel{a: a}
}

func (rm *runModel) start(p problem.Problem, fast bool) {
	rm.prob = p
	rm.fast = fast
	rm.phase = 1
	rm.compileOut = ""
	rm.cases = nil
	rm.results = nil
	rm.cur = 0
	rm.cursor = 0
	rm.tab = 0
	rm.fastElapsed = ""
	rm.ctx, rm.cancel = context.WithCancel(context.Background())
	rm.vp = viewport.New(0, 0)
	rm.vpOK = false
}

func (rm *runModel) compileCmd() tea.Cmd {
	dir := rm.prob.Dir
	cxx := rm.a.cfg.CXX
	fast := rm.fast
	return func() tea.Msg {
		if fast {
			stderr, elapsed, err := runner.FastBuild(dir, cxx)
			return compileDoneMsg{Stderr: fmt.Sprintf("빌드 시간 %s\n%s", elapsed, stderr), Err: err}
		}
		stderr, err := runner.Compile(dir, cxx)
		return compileDoneMsg{Stderr: stderr, Err: err}
	}
}

func (rm *runModel) caseCmd(i int) tea.Cmd {
	if rm.ctx == nil {
		return nil
	}
	ctx := rm.ctx
	dir := rm.prob.Dir
	bin := "sol"
	c := rm.cases[i]
	return func() tea.Msg {
		res := runner.RunCaseWith(ctx, dir, bin, c)
		return caseDoneMsg{Index: i, Result: res}
	}
}

func (rm *runModel) update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		rm.resizeVP()

	case compileDoneMsg:
		if m.Err != nil {
			rm.compileOut = m.Stderr + "\n" + m.Err.Error()
			rm.phase = 3
			return nil
		}
		rm.compileOut = m.Stderr
		rm.phase = 2
		dir := rm.prob.Dir
		return func() tea.Msg {
			cases, err := runner.Cases(dir)
			return casesLoadedMsg{Cases: cases, Err: err}
		}

	case casesLoadedMsg:
		if m.Err != nil {
			rm.compileOut = m.Err.Error()
			rm.phase = 3
			return nil
		}
		rm.cases = m.Cases
		rm.results = make([]runner.CaseResult, len(m.Cases))
		if len(m.Cases) == 0 {
			rm.phase = 3
			return nil
		}
		rm.cur = 0
		return rm.caseCmd(0)

	case caseDoneMsg:
		rm.results[m.Index] = m.Result
		rm.cursor = m.Index
		rm.refreshDiff()
		if rm.phase == 4 {
			return nil
		}
		if m.Index+1 < len(rm.cases) {
			rm.cur = m.Index + 1
			return rm.caseCmd(m.Index + 1)
		}
		rm.phase = 3
		rm.finishSummary()

	case tea.KeyMsg:
		switch m.String() {
		case "esc":
			if rm.phase == 2 && rm.cancel != nil {
				rm.phase = 4
				rm.cancel()
				return nil
			}
			rm.finishSummary()
			p := rm.prob
			return func() tea.Msg { return openWorkbenchMsg{Prob: p} }
		case "tab":
			rm.tab = (rm.tab + 1) % 3
			rm.refreshDiff()
		case "d":
			if rm.phase == 3 {
				p := rm.prob
				def := "fail"
				if rm.allPass() {
					def = "ok"
				}
				return func() tea.Msg { return openDoneMsg{Prob: p, Default: def} }
			}
		case "up", "k":
			if rm.cursor > 0 {
				rm.cursor--
				rm.refreshDiff()
			}
		case "down", "j":
			if rm.cursor < len(rm.cases)-1 {
				rm.cursor++
				rm.refreshDiff()
			}
		}
	}
	return nil
}

func (rm *runModel) resizeVP() {
	w := rm.a.width*55/100 - 4
	h := rm.a.height - 8
	if w < 20 || h < 3 {
		return
	}
	rm.vp.Width = w
	rm.vp.Height = h
	rm.vpOK = true
	rm.refreshDiff()
}

func (rm *runModel) allPass() bool {
	pass := 0
	for _, r := range rm.results {
		if r.Status == runner.Pass {
			pass++
		}
	}
	return pass == len(rm.cases) && len(rm.cases) > 0
}

func (rm *runModel) passCount() int {
	n := 0
	for _, r := range rm.results {
		if r.Status == runner.Pass {
			n++
		}
	}
	return n
}

func (rm *runModel) finishSummary() {
	if rm.prob.Dir == "" {
		return
	}
	bin := "sol"
	if rm.fast {
		bin = "fast"
	}
	start, ok := problem.ReadStart(rm.prob.Dir)
	mins := 0
	if ok {
		mins = problem.ElapsedMinutes(start, time.Now())
	}
	rm.a.bench.setSummary(bin, len(rm.results), rm.passCount(), mins)
}

func (rm *runModel) refreshDiff() {
	if !rm.vpOK || rm.cursor >= len(rm.results) {
		return
	}
	rm.vp.SetContent(rm.detailContent())
	rm.vp.GotoTop()
}

func (rm *runModel) detailContent() string {
	if rm.cursor >= len(rm.results) {
		return ""
	}
	r := rm.results[rm.cursor]
	switch rm.tab {
	case 1:
		return r.Stderr
	case 2:
		return rm.compileOut
	default:
		var b strings.Builder
		if r.Truncated {
			b.WriteString(stWarn.Render("(출력이 1MiB에서 잘림)"))
			b.WriteByte('\n')
		}
		if r.Status == runner.NoExp {
			return b.String() + stWarn.Render("기대출력(.out) 없음 — 판정 제외")
		}
		if linediff.WhitespaceOnly(r.Expected, r.Actual) {
			b.WriteString(stWarn.Render("공백만 다름 (줄 끝 공백/CR)"))
			b.WriteByte('\n')
		}
		for _, l := range linediff.Diff(splitLines(string(r.Expected)), splitLines(string(r.Actual))) {
			switch l.Kind {
			case linediff.Keep:
				b.WriteString("  " + l.Text + "\n")
			case linediff.Del:
				b.WriteString(stFail.Render("- "+l.Text) + "\n")
			case linediff.Add:
				b.WriteString(stOK.Render("+ "+l.Text) + "\n")
			}
		}
		return b.String()
	}
}

func splitLines(s string) []string {
	s = strings.TrimSuffix(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func (rm *runModel) view() string {
	var b strings.Builder
	name := "빌드와 테스트 (sanitizer)"
	if rm.fast {
		name = "fast 빌드와 시간 측정"
	}
	b.WriteString(stTitle.Render(" " + name + " · " + rm.prob.Site + "/" + rm.prob.Name))
	b.WriteString("  " + stDim.Render(rm.phaseText()))
	b.WriteByte('\n')
	if len(rm.cases) == 0 && rm.phase == 3 {
		b.WriteString(stWarn.Render(" 테스트 케이스 없음 — Workbench에서 a 키로 생성"))
		b.WriteByte('\n')
		return b.String()
	}
	left := rm.caseList()
	right := rm.vp.View()
	if !rm.vpOK {
		right = rm.detailContent()
	}
	b.WriteString(joinHorizontalPanes(padBlock(left, rm.a.height-7), padBlock(right, rm.a.height-7)))
	b.WriteByte('\n')
	b.WriteString(stDim.Render(fmt.Sprintf(" %d/%d PASS · Tab: diff/stderr/컴파일 · d: 결과 기록 · esc: %s",
		rm.passCount(), len(rm.cases), escLabel(rm.phase))))
	return b.String()
}

func escLabel(phase int) string {
	if phase == 2 {
		return "취소"
	}
	return "뒤로"
}

func (rm *runModel) phaseText() string {
	switch rm.phase {
	case 1:
		return "컴파일 중…"
	case 2:
		return fmt.Sprintf("케이스 %d/%d 실행 중", rm.cur+1, len(rm.cases))
	case 4:
		return "취소됨"
	default:
		if rm.compileOut != "" && rm.results == nil {
			return "컴파일 실패 (Tab: 컴파일 로그)"
		}
		return "완료"
	}
}

func (rm *runModel) caseList() string {
	var b strings.Builder
	b.WriteString(stTitle.Render(" 케이스"))
	b.WriteByte('\n')
	for i, c := range rm.cases {
		marker, style := " ", stNormal
		if i == rm.cursor {
			marker, style = ">", stSel
		}
		status := "…"
		var ms int64
		if rm.phase == 1 && i == 0 {
			status = ""
		}
		if i < len(rm.results) && rm.results[i].Status != "" {
			s := rm.results[i].Status
			ms = rm.results[i].Millis
			switch s {
			case runner.Pass:
				status = stOK.Render(string(s))
			case runner.NoExp:
				status = stWarn.Render(string(s))
			default:
				status = stFail.Render(string(s))
			}
		}
		b.WriteString(style.Render(fmt.Sprintf("%s %-8s %-6s %4dms", marker, basename(c.In), status, ms)))
		b.WriteByte('\n')
	}
	return b.String()
}

func basename(p string) string {
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}
