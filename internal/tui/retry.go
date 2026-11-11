package tui

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"pstui/internal/ps/problem"
)

type retryDoneMsg struct {
	Prob   problem.Problem
	Backup string
	Err    error
}

type retryModel struct {
	a      *App
	prob   problem.Problem
	err    string
	backup string
}

func newRetry(a *App) *retryModel { return &retryModel{a: a} }

func (rm *retryModel) start(p problem.Problem) {
	rm.prob = p
	rm.err = ""
	if !rm.a.cfg.TemplateOK() {
		rm.err = "template.cpp 없음 — 재도전 불가"
	}
	// 오늘 백업명 미리보기 (이미 있으면 덮어쓰기 경고)
	rm.backup = "main." + time.Now().Format("0102") + ".cpp"
}

func (rm *retryModel) update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case retryDoneMsg:
		if m.Err != nil {
			rm.err = m.Err.Error()
			return nil
		}
		p := m.Prob
		return tea.Batch(
			func() tea.Msg { return openWorkbenchMsg{Prob: p} },
			rm.a.editorCmd(p, "main.cpp"),
		)

	case tea.KeyMsg:
		switch m.String() {
		case "enter", "y":
			if rm.err != "" {
				return nil
			}
			p := rm.prob
			template := rm.a.cfg.Template
			dir := p.Dir
			return func() tea.Msg {
				backup, err := problem.Retry(dir, template, time.Now())
				return retryDoneMsg{Prob: p, Backup: backup, Err: err}
			}
		case "esc", "n":
			p := rm.prob
			return func() tea.Msg { return openWorkbenchMsg{Prob: p} }
		}
	}
	return nil
}

func (rm *retryModel) view() string {
	a := rm.a
	var lines []string
	lines = append(lines, stTitle.Render(rm.backup)+" "+stDim.Render("로 백업 후 템플릿으로 초기화"))
	lines = append(lines, stDim.Render(".start를 지금으로 갱신하고 편집기를 엽니다"))
	lines = append(lines, "")
	if _, err := os.Stat(filepath.Join(rm.prob.Dir, rm.backup)); err == nil {
		lines = append(lines, stWarn.Render("⚠ "+rm.backup+" 이(가) 이미 있습니다 — ⏎ 시 덮어씁니다"))
	} else {
		lines = append(lines, stDim.Render("기존 풀이는 사라지지 않고 백업 파일로 남습니다"))
	}
	if rm.err != "" {
		lines = append(lines, stFail.Render("⚠ "+rm.err))
	}
	content := strings.Join(lines, "\n")
	w := 56
	if a.width-10 < w {
		w = a.width - 10
	}
	h := len(lines) + 5
	if a.height-4 < h {
		h = a.height - 4
	}
	return lipgloss.Place(a.width, a.height-2, lipgloss.Center, lipgloss.Center,
		paneBox("재도전 · "+rm.prob.Site+"/"+rm.prob.Name, content, w, h, true))
}

func (rm *retryModel) helpText() string {
	return keyBar(
		[2]string{"⏎/y", "실행"}, [2]string{"esc/n", "취소"},
	)
}
