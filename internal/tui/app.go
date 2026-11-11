// Package tui는 bubbletea 화면 계층이다. 도메인(internal/ps) 호출은 전부 tea.Cmd 안에서 실행한다.
package tui

import (
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
	scrPlan
	scrStmt
)

type App struct {
	cfg           config.Config
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
	pickM  *pickerModel
	stmtM  *stmtModel

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
	a.pickM = newPicker(a)
	a.stmtM = newStmt(a)
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
		return a, tea.Batch(a.open(scrRun), a.runM.compileCmd())
	case openDoneMsg:
		a.doneM.start(m.Prob, m.Default)
		return a, a.open(scrDone)
	case openPlanMsg:
		return a, a.open(scrPlan)
	case openStmtMsg:
		cmd := a.stmtM.open(m.Prob, m.Site, m.Slug)
		return a, tea.Batch(a.open(scrStmt), cmd)
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
	case scrPlan:
		cmd = a.pickM.update(msg)
	case scrStmt:
		cmd = a.stmtM.update(msg)
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
	case scrPlan:
		body = a.pickM.view()
	case scrStmt:
		body = a.stmtM.view()
	}
	if a.showHelp {
		body = helpCard(a.helpText(), a.width, a.height-2)
	}
	return a.frame(body)
}

var screenNames = map[screen]string{
	scrHome: "홈", scrProblems: "문제", scrWorkbench: "작업", scrRun: "실행",
	scrDone: "기록", scrNew: "새 문제", scrRetry: "재도전", scrLog: "로그", scrStats: "통계",
	scrPlan: "플랜",
	scrStmt: "지문",
}

func (a *App) frame(body string) string {
	left := stAccent.Render("pstui") + stDim.Render(" · "+screenNames[a.cur])
	right := stDim.Render(a.cfg.Root + " · " + time.Now().Format("2006-01-02 15:04"))
	head := header(left, right, a.width)
	body = fitLines(body, a.height-2)
	return head + "\n" + body + "\n" + a.helpText()
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
		return a.runM.helpText()
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
	case scrPlan:
		return a.pickM.helpText()
	case scrStmt:
		return a.stmtM.helpText()
	}
	return ""
}
