package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

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
	var b strings.Builder
	b.WriteString(stTitle.Render(fmt.Sprintf(" 재도전 · %s/%s", rm.prob.Site, rm.prob.Name)))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf(" main.cpp → %s 로 백업 후 템플릿으로 초기화하고 .start를 지금으로 갱신합니다.\n", rm.backup))
	if _, err := os.Stat(filepath.Join(rm.prob.Dir, rm.backup)); err == nil {
		b.WriteString(stWarn.Render(fmt.Sprintf(" ⚠ %s 이(가) 이미 있습니다. Enter를 누르면 덮어씁니다.", rm.backup)))
		b.WriteByte('\n')
	}
	if rm.err != "" {
		b.WriteString(stFail.Render("⚠ " + rm.err))
		b.WriteByte('\n')
	}
	b.WriteString(stDim.Render(" enter/y 실행 · esc/n 취소"))
	return b.String()
}

func (rm *retryModel) helpText() string {
	return "Retry: enter 실행 · esc 취소"
}
