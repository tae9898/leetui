package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"leetui/internal/ps/linediff"
	"leetui/internal/ps/problem"
	"leetui/internal/ps/runner"
)

type runModel struct {
	a    *App
	prob problem.Problem
	fast bool
	only string // 커스텀 입력처럼 이 파일만 실행 (""=전체)

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
}

func newRun(a *App) *runModel {
	return &runModel{a: a}
}

func (rm *runModel) start(p problem.Problem, fast bool, only string) {
	rm.prob = p
	rm.fast = fast
	rm.only = only
	rm.phase = 1
	rm.compileOut = ""
	rm.cases = nil
	rm.results = nil
	rm.cur = 0
	rm.cursor = 0
	rm.tab = 0
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
		if rm.only != "" {
			var filtered []runner.Case
			for _, c := range m.Cases {
				if filepath.Base(c.In) == rm.only {
					filtered = append(filtered, c)
				}
			}
			rm.cases = filtered
		}
		rm.results = make([]runner.CaseResult, len(rm.cases))
		if len(rm.cases) == 0 {
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
		case "tab", "l":
			rm.tab = (rm.tab + 1) % 3
			rm.refreshDiff()
		case "h":
			rm.tab = (rm.tab + 2) % 3
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
	w := rm.a.width*62/100 - 5
	h := rm.a.height - 2 - 1 - 3 // 키바 + 상태줄 + 박스 크롬
	if w < 20 || h < 3 {
		return
	}
	rm.vp.Width = w
	rm.vp.Height = h
	rm.vpOK = true
	rm.refreshDiff()
}

func (rm *runModel) allPass() bool {
	pass, judged := 0, 0
	for _, r := range rm.results {
		if r.Status == runner.NoExp {
			continue // 기대출력 없는 케이스(custom.in 등)는 판정에서 제외
		}
		judged++
		if r.Status == runner.Pass {
			pass++
		}
	}
	return pass == judged && judged > 0
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
	if rm.prob.Dir == "" || rm.only != "" {
		return // 커스텀 입력 실행은 요약에 남기지 않음
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
			b.WriteString(stWarn.Render("기대출력(.out) 없음 — 판정 제외, 출력만 표시"))
			b.WriteString("\n\n")
			for _, l := range splitLines(string(r.Actual)) {
				b.WriteString("  " + l + "\n")
			}
			return b.String()
		}
		if r.Status == runner.Fail && linediff.WhitespaceOnly(r.Expected, r.Actual) {
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

func (rm *runModel) statusLine() string {
	name := "sanitizer 빌드·테스트"
	if rm.fast {
		name = "fast 빌드"
	}
	if rm.only != "" {
		name = "커스텀 입력 · " + rm.only
	}
	var phase string
	switch rm.phase {
	case 1:
		phase = stTitle.Render("컴파일 중…")
	case 2:
		phase = stWarn.Render(fmt.Sprintf("케이스 %d/%d 실행 중", rm.cur+1, len(rm.cases)))
	case 4:
		phase = stFail.Render("취소됨")
	default:
		if rm.compileOut != "" && rm.results == nil {
			phase = stFail.Render("컴파일 실패 (Tab: 로그)")
		} else {
			phase = stOK.Render("완료")
		}
	}
	return " " + stDim.Render(name+" · "+rm.prob.Site+"/"+rm.prob.Name+" · ") + phase +
		stDim.Render(fmt.Sprintf(" · %d/%d PASS", rm.passCount(), len(rm.cases)))
}

func (rm *runModel) detailTitle() string {
	name := ""
	if rm.cursor < len(rm.cases) {
		name = " · " + basename(rm.cases[rm.cursor].In)
	}
	switch rm.tab {
	case 1:
		return "stderr" + name
	case 2:
		return "컴파일 로그"
	default:
		return "diff" + name
	}
}

func (rm *runModel) view() string {
	a := rm.a
	paneH := a.height - 2 - 1 // 키바 + 상태줄
	if paneH < 5 {
		paneH = 5
	}
	leftW := a.width * 38 / 100
	rightW := a.width - leftW - 1

	caseTitle := fmt.Sprintf("케이스 · %d", len(rm.cases))
	if len(rm.cases) == 0 && rm.phase == 3 {
		caseTitle = "케이스 없음 — 작업 화면 a 키로 생성"
	}
	left := paneBox(caseTitle, rm.caseRows(leftW-4), leftW, paneH, true)
	detail := rm.detailContent()
	if rm.vpOK {
		detail = rm.vp.View()
	}
	right := paneBox(rm.detailTitle(), detail, rightW, paneH, false)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, " ", right)
	return rm.statusLine() + "\n" + body
}

func (rm *runModel) caseRows(innerW int) string {
	if innerW < 18 {
		return ""
	}
	var b strings.Builder
	b.WriteString(stDimmer.Render(padW("케이스", innerW-12) + padW("판정", 6) + "시간"))
	b.WriteByte('\n')
	for i, c := range rm.cases {
		var status, ms string
		if i < len(rm.results) && rm.results[i].Status != "" {
			s := string(rm.results[i].Status)
			ms = fmt.Sprintf("%4dms", rm.results[i].Millis)
			switch rm.results[i].Status {
			case runner.Pass:
				status = stOK.Render(padW(s, 6))
			case runner.NoExp:
				status = stWarn.Render(padW(s, 6))
			default:
				status = stFail.Render(padW(s, 6))
			}
		} else if i == rm.cur && rm.phase == 2 {
			status = stTitle.Render(padW("…", 6))
		} else {
			status = padW("·", 6)
		}
		nameW := innerW - 12
		row := padW(truncateStr(basename(c.In), nameW-1), nameW) + status + ms
		b.WriteString(selRow(row, innerW, i == rm.cursor))
		if i < len(rm.cases)-1 {
			b.WriteByte('\n')
		}
	}
	if len(rm.cases) == 0 {
		b.WriteString(stDim.Render(" .in 파일이 없음"))
	}
	return b.String()
}

func (rm *runModel) helpText() string {
	esc := "뒤로"
	if rm.phase == 2 {
		esc = "취소"
	}
	return keyBar(
		[2]string{"j/k", "케이스"}, [2]string{"h/l", "diff/stderr/로그"},
		[2]string{"d", "결과 기록"}, [2]string{"esc", esc},
	)
}

func basename(p string) string {
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		return p[i+1:]
	}
	return p
}
