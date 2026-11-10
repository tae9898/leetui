// Package tui는 bubbletea 화면 계층. 도메인(internal/ps)만 호출한다.
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"pstui/internal/ps/config"
)

type App struct {
	cfg config.Config
	width, height int
	ready  bool
}

func Run(cfg config.Config) error {
	app := &App{cfg: cfg}
	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (a *App) Init() tea.Cmd { return nil }

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = m.Width, m.Height
		a.ready = true
	case tea.KeyMsg:
		switch m.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}

func (a *App) View() string {
	if !a.ready {
		return "로딩 중..."
	}
	return fmt.Sprintf("pstui · %s · q 종료", a.cfg.Root)
}
