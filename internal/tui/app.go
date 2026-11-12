// Package tui는 bubbletea 화면 계층이다. 도메인(internal/ps) 호출은 전부 tea.Cmd 안에서 실행한다.
package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"leetui/internal/ps/config"
	"leetui/internal/ps/problem"
	"leetui/internal/ps/pslog"
	"leetui/internal/ps/stmt"
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

	showHelp bool
	notice   string // 하단 키 바 대신 잠깐 뜨는 알림 (아무 키나 누르면 사라짐)
	noticeOK bool   // 참이면 ✓ 스타일, 거짓이면 ⚠ 스타일
	busy     bool   // 지문/예제 비동기 작업 중 — 중복 발사 방지
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
		return editorDoneMsg{Prob: prob, File: filepath.Base(file), Err: err}
	})
}

// execViewer는 지문 파일을 편집기(nvim)로 연다. TUI가 잠시 내려가고 종료 시 복귀한다.
func (a *App) execViewer(path string) tea.Cmd {
	cmd := exec.Command(a.cfg.Editor, path)
	cmd.Dir = filepath.Dir(path)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return viewerDoneMsg{Err: err}
	})
}

func (a *App) setNotice(ok bool, s string) {
	a.notice, a.noticeOK = s, ok
}

// stmtOpenCmd는 지문을 캐시/fetch해 편집기로 연다.
func (a *App) stmtOpenCmd(m openStmtMsg) tea.Cmd {
	if a.busy {
		return nil
	}
	a.busy = true
	a.setNotice(false, "지문 가져오는 중… "+m.Site+"/"+m.Slug)
	root, site, slug := a.cfg.Root, m.Site, m.Slug
	return func() tea.Msg {
		path, err := stmt.Ensure(context.Background(), root, site, slug)
		return stmtReadyMsg{Path: path, Err: err}
	}
}

// stmtFillCmd는 예제를 .in/.out으로 채운다. open이면 채운 뒤 작업 화면으로 간다.
func (a *App) stmtFillCmd(p problem.Problem, slug string, open bool) tea.Cmd {
	if a.busy {
		return nil
	}
	a.busy = true
	a.setNotice(false, "예제 가져오는 중… "+p.Site+"/"+slug)
	root, template := a.cfg.Root, a.cfg.Template
	return func() tea.Msg {
		n, err := stmt.Fill(context.Background(), root, slug, p, template, time.Now())
		return stmtFilledMsg{N: n, Prob: p, Open: open, Err: err}
	}
}

// slugOf는 문제 이름 "0001-two-sum"에서 슬러그 "two-sum"을 뽑는다.
func slugOf(p problem.Problem) string {
	n := p.Name
	if len(n) > 5 && n[4] == '-' && n[:4] >= "0000" && n[:4] <= "9999" {
		return n[5:]
	}
	return n
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
		a.notice = "" // 알림은 아무 키나 누르면 사라진다
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
		if m.Err != nil {
			a.setNotice(false, "편집기 실패: "+m.Err.Error())
			return a, a.bench.reload(m.Prob)
		}
		if m.File == "custom.in" {
			// 커스텀 입력(HackerRank Run Custom Input 대응) — 그 파일만 fast 실행
			if _, err := os.Stat(filepath.Join(m.Prob.Dir, "custom.in")); err != nil {
				a.setNotice(false, "custom.in이 저장되지 않음")
				return a, a.bench.reload(m.Prob)
			}
			return a, tea.Batch(func() tea.Msg {
				return openRunMsg{Prob: m.Prob, Fast: true, Only: "custom.in"}
			}, a.bench.reload(m.Prob))
		}
		return a, tea.Batch(loadLogCmd(a.cfg.Root), a.bench.reload(m.Prob))

	case openHomeMsg:
		return a, a.open(scrHome)
	case openProblemsMsg:
		return a, a.open(scrProblems)
	case openWorkbenchMsg:
		a.bench.setProblem(m.Prob)
		return a, tea.Batch(a.open(scrWorkbench), a.bench.reload(m.Prob))
	case openRunMsg:
		a.runM.start(m.Prob, m.Fast, m.Only)
		return a, tea.Batch(a.open(scrRun), a.runM.compileCmd())
	case openDoneMsg:
		a.doneM.start(m.Prob, m.Default)
		return a, a.open(scrDone)
	case openPlanMsg:
		return a, a.open(scrPlan)
	case openStmtMsg:
		return a, a.stmtOpenCmd(m)
	case stmtReadyMsg:
		a.busy = false
		if m.Err != nil {
			a.setNotice(false, "지문 실패: "+m.Err.Error())
			return a, nil
		}
		a.notice = ""
		return a, a.execViewer(m.Path)
	case viewerDoneMsg:
		if m.Err != nil {
			a.setNotice(false, "지문 뷰어 실패: "+m.Err.Error())
		}
		return a, nil
	case stmtFilledMsg:
		a.busy = false
		if m.Err != nil {
			a.setNotice(false, "예제 채움 실패: "+m.Err.Error())
			return a, nil
		}
		if m.N == 0 {
			a.setNotice(true, "채울 예제 없음")
			return a, nil
		}
		a.setNotice(true, fmt.Sprintf("예제 %d개 스펙 저장 (.exN.in) — 실행 케이스는 자기 포맷으로", m.N))
		if m.Open {
			a.bench.setProblem(m.Prob)
			return a, tea.Batch(a.open(scrWorkbench), a.bench.reload(m.Prob))
		}
		return a, a.bench.reload(m.Prob)
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
}

func (a *App) frame(body string) string {
	left := stAccent.Render("leetui") + stDim.Render(" · "+screenNames[a.cur])
	right := stDim.Render(a.cfg.Root + " · " + time.Now().Format("2006-01-02 15:04"))
	head := header(left, right, a.width)
	body = fitLines(body, a.height-2)
	bottom := a.helpText()
	if a.notice != "" {
		if a.noticeOK {
			bottom = stOK.Render(" ✓ " + truncateStr(a.notice, a.width-4))
		} else {
			bottom = stFail.Render(" ⚠ " + truncateStr(a.notice, a.width-4))
		}
	}
	return head + "\n" + body + "\n" + bottom
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
	}
	return ""
}

