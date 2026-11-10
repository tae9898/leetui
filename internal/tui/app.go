// Package tui는 bubbletea 화면 계층이다. 도메인(internal/ps) 호출은 전부 tea.Cmd 안에서 실행한다.
package tui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"pstui/internal/ps/config"
	"pstui/internal/ps/problem"
	"pstui/internal/ps/pslog"
)

type screen int

const (
	scrHome screen = iota
	scrProblems
	scrWorkbench
	scrRun
	scrDone
	scrNew
	scrRetry
	scrLog
	scrStats
)

type App struct {
	cfg          config.Config
	width, height int

	entries  []pslog.Entry
	skipped  int
	loadErr  error
	problems []problem.Problem

	cur  screen
	prev screen

	home   *homeModel
	probs  *problemsModel
	bench  *workbenchModel
	runM   *runModel
	doneM  *doneModel
	newM   *newModel
	retryM *retryModel
	logM   *logModel
	statsM *statsModel

	showHelp bool
}

func Run(cfg config.Config) error {
	app := newApp(cfg)
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func newApp(cfg config.Config) *App {
	a := &App{cfg: cfg, width: 80, height: 24}
	a.home = newHome(a)
	a.probs = newProblems(a)
	a.bench = newWorkbench(a)
	a.runM = newRun(a)
	a.doneM = newDone(a)
	a.newM = newNew(a)
	a.retryM = newRetry(a)
	a.logM = newLog(a)
	a.statsM = newStats(a)
	return a
}

func (a *App) Init() tea.Cmd {
	return tea.Batch(loadLogCmd(a.cfg.Root), scanProblemsCmd(a.cfg.Root))
}

func loadLogCmd(root string) tea.Cmd {
	return func() tea.Msg {
		entries, skipped, err := pslog.Parse(filepath.Join(root, "log.tsv"))
		return logLoadedMsg{Entries: entries, Skipped: skipped, Err: err}
	}
}

func scanProblemsCmd(root string) tea.Cmd {
	return func() tea.Msg {
		probs, err := problem.Scan(root)
		return problemsLoadedMsg{Problems: probs, Err: err}
	}
}

func (a *App) editorCmd(prob problem.Problem, file string) tea.Cmd {
	cmd := exec.Command(a.cfg.Editor, file)
	cmd.Dir = prob.Dir
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return editorDoneMsg{Prob: prob, Err: err}
	})
}

func (a *App) open(s screen) tea.Cmd {
	a.prev = a.cur
	a.cur = s
	var cmd tea.Cmd
	switch s {
	case scrHome:
		cmd = tea.Batch(loadLogCmd(a.cfg.Root), scanProblemsCmd(a.cfg.Root))
		a.home.cursor = 0
	case scrProblems:
		cmd = scanProblemsCmd(a.cfg.Root)
	case scrLog:
		a.logM.reset()
	case scrStats:
		cmd = scanProblemsCmd(a.cfg.Root)
	}
	return cmd
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = m.Width, m.Height
		// 크기 정보는 화면(뷰포트 등)에도 전달되어야 하므로 fall-through

	case tea.KeyMsg:
		switch m.String() {
		case "ctrl+c":
			return a, tea.Quit
		case "?":
			a.showHelp = !a.showHelp
			return a, nil
		case "esc":
			if a.showHelp {
				a.showHelp = false
				return a, nil
			}
		}

	case logLoadedMsg:
		a.entries, a.skipped, a.loadErr = m.Entries, m.Skipped, m.Err
	case problemsLoadedMsg:
		if m.Err == nil {
			a.problems = m.Problems
		}
	case editorDoneMsg:
		return a, tea.Batch(loadLogCmd(a.cfg.Root), a.bench.reload(m.Prob))

	case openHomeMsg:
		return a, a.open(scrHome)
	case openProblemsMsg:
		return a, a.open(scrProblems)
	case openWorkbenchMsg:
		a.bench.setProblem(m.Prob)
		return a, tea.Batch(a.open(scrWorkbench), a.bench.reload(m.Prob))
	case openRunMsg:
		a.runM.start(m.Prob, m.Fast)
		return a, a.open(scrRun)
	case openDoneMsg:
		a.doneM.start(m.Prob, m.Default)
		return a, a.open(scrDone)
	case openNewMsg:
		a.newM.reset()
		return a, a.open(scrNew)
	case newCreatedMsg:
		if m.Err != nil {
			a.newM.err = m.Err.Error()
			return a, nil
		}
		a.bench.setProblem(m.Prob)
		return a, tea.Batch(a.open(scrWorkbench), a.bench.reload(m.Prob), a.editorCmd(m.Prob, "main.cpp"))
	case openRetryMsg:
		a.retryM.start(m.Prob)
		return a, a.open(scrRetry)
	case openLogMsg:
		return a, a.open(scrLog)
	case openStatsMsg:
		return a, a.open(scrStats)
	}

	var cmd tea.Cmd
	switch a.cur {
	case scrHome:
		cmd = a.home.update(msg)
	case scrProblems:
		cmd = a.probs.update(msg)
	case scrWorkbench:
		cmd = a.bench.update(msg)
	case scrRun:
		cmd = a.runM.update(msg)
	case scrDone:
		cmd = a.doneM.update(msg)
	case scrNew:
		cmd = a.newM.update(msg)
	case scrRetry:
		cmd = a.retryM.update(msg)
	case scrLog:
		cmd = a.logM.update(msg)
	case scrStats:
		cmd = a.statsM.update(msg)
	}
	return a, cmd
}

func (a *App) View() string {
	if a.width < 80 || a.height < 24 {
		return "터미널이 너무 작습니다. 80x24 이상으로 키우세요."
	}
	var body string
	switch a.cur {
	case scrHome:
		body = a.home.view()
	case scrProblems:
		body = a.probs.view()
	case scrWorkbench:
		body = a.bench.view()
	case scrRun:
		body = a.runM.view()
	case scrDone:
		body = a.doneM.view()
	case scrNew:
		body = a.newM.view()
	case scrRetry:
		body = a.retryM.view()
	case scrLog:
		body = a.logM.view()
	case scrStats:
		body = a.statsM.view()
	}
	if a.showHelp {
		body = overlayHelp(body, a.helpText())
	}
	return a.frame(body)
}

func (a *App) frame(body string) string {
	title := fmt.Sprintf(" pstui · %s · %s ", a.cfg.Root, time.Now().Format("2006-01-02"))
	footer := " ?: 도움말 · ctrl+c 종료"
	if a.cur != scrHome {
		footer = " esc: 뒤로 ·" + footer
	} else {
		footer = " q: 종료 ·" + footer
	}
	return stTitle.Render(title) + "\n" + body + "\n" + stDim.Render(footer)
}

func (a *App) dueList() []pslog.DueItem {
	return pslog.Due(a.entries, time.Now().Format("2006-01-02"))
}

// lastEntry는 문제별 마지막 로그 행을 반환한다 (없으면 nil).
func (a *App) lastEntry(site, name string) *pslog.Entry {
	var found *pslog.Entry
	for i := range a.entries {
		e := &a.entries[i]
		if e.Site == site && e.Problem == name {
			found = e
		}
	}
	return found
}

func (a *App) problemEntries(site, name string) []pslog.Entry {
	var out []pslog.Entry
	for _, e := range a.entries {
		if e.Site == site && e.Problem == name {
			out = append(out, e)
		}
	}
	return out
}

func (a *App) findProblem(site, name string) *problem.Problem {
	for i := range a.problems {
		if a.problems[i].Site == site && a.problems[i].Name == name {
			return &a.problems[i]
		}
	}
	return nil
}

func (a *App) warnings() []string {
	var ws []string
	if !a.cfg.TemplateOK() {
		ws = append(ws, "template.cpp 없음 — 새 문제/재도전 비활성")
	}
	if a.loadErr != nil {
		ws = append(ws, "log.tsv 읽기 실패: "+a.loadErr.Error())
	}
	return ws
}

func (a *App) helpText() string {
	switch a.cur {
	case scrHome:
		return a.home.helpText()
	case scrProblems:
		return a.probs.helpText()
	case scrWorkbench:
		return a.bench.helpText()
	case scrRun:
		return "Run: j/k 케이스 · Tab diff/stderr/컴파일 · d 결과 기록 · esc 취소/뒤로"
	case scrDone:
		return a.doneM.helpText()
	case scrNew:
		return a.newM.helpText()
	case scrRetry:
		return a.retryM.helpText()
	case scrLog:
		return a.logM.helpText()
	case scrStats:
		return a.statsM.helpText()
	}
	return ""
}
