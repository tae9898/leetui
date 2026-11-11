package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"pstui/internal/ps/problem"
)

// newModel: 새 문제 폼. 사이트는 기존 목록 순환 + 직접 입력, id/name 입력.
type newModel struct {
	a      *App
	site   textinput.Model
	id     textinput.Model
	name   textinput.Model
	field  int
	err    string
	exists bool // 대상 디렉터리가 이미 있음
}

func newNew(a *App) *newModel {
	n := &newModel{a: a}
	n.site = textinput.New()
	n.site.Prompt = ""
	n.site.Placeholder = "lc (엔터로 다음)"
	n.site.CharLimit = 20
	n.id = textinput.New()
	n.id.Prompt = ""
	n.id.Placeholder = "0056 (엔터로 다음)"
	n.id.CharLimit = 12
	n.name = textinput.New()
	n.name.Prompt = ""
	n.name.Placeholder = "merge-intervals (엔터로 생성)"
	n.name.CharLimit = 60
	return n
}

func (n *newModel) reset() {
	n.site.SetValue("")
	n.id.SetValue("")
	n.name.SetValue("")
	n.field = 0
	n.err = ""
	n.exists = false
	n.focusActive()
}

// focusActive는 현재 필드만 포커스한다 (포커스 없는 textinput은 입력을 무시한다).
func (n *newModel) focusActive() {
	n.site.Blur()
	n.id.Blur()
	n.name.Blur()
	switch n.field {
	case 0:
		n.site.Focus()
	case 1:
		n.id.Focus()
	case 2:
		n.name.Focus()
	}
}

func (n *newModel) knownSites() []string {
	set := map[string]bool{}
	for _, p := range n.a.problems {
		set[p.Site] = true
	}
	if ents, err := os.ReadDir(n.a.cfg.Root); err == nil {
		for _, e := range ents {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				set[e.Name()] = true
			}
		}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func (n *newModel) dirPath() string {
	return filepath.Join(n.a.cfg.Root, n.site.Value(), n.id.Value()+"-"+n.name.Value())
}

func (n *newModel) update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "tab":
			n.field = (n.field + 1) % 3
			n.focusActive()
			return nil
		case "shift+tab":
			n.field = (n.field + 2) % 3
			n.focusActive()
			return nil
		case "enter":
			if n.field < 2 {
				n.field++
				n.focusActive()
				return nil
			}
			return n.submit()
		case "esc":
			return func() tea.Msg { return openHomeMsg{} }
		}
		var active *textinput.Model
		switch n.field {
		case 0:
			active = &n.site
		case 1:
			active = &n.id
		case 2:
			active = &n.name
		}
		if active != nil {
			cmd := textinputBubbling(active, m)
			_, statErr := os.Stat(n.dirPath())
			n.exists = statErr == nil && n.id.Value() != "" && n.name.Value() != ""
			return cmd
		}
	}
	return nil
}

func (n *newModel) submit() tea.Cmd {
	site, id, name := strings.TrimSpace(n.site.Value()), strings.TrimSpace(n.id.Value()), strings.TrimSpace(n.name.Value())
	if site == "" || id == "" || name == "" {
		n.err = "세 필드 모두 필요합니다"
		return nil
	}
	if err := problem.Validate(site, id, name); err != nil {
		n.err = err.Error()
		return nil
	}
	dir := filepath.Join(n.a.cfg.Root, site, id+"-"+name)
	if _, err := os.Stat(dir); err == nil {
		// 기존 문제 열기: main.cpp 덮어쓰지 않음
		p := problem.Problem{Site: site, Name: id + "-" + name, Dir: dir}
		return tea.Batch(func() tea.Msg { return openWorkbenchMsg{Prob: p} }, n.a.editorCmd(p, "main.cpp"))
	}
	if !n.a.cfg.TemplateOK() {
		n.err = "template.cpp 없음 — 새 문제를 만들 수 없습니다"
		return nil
	}
	root, template := n.a.cfg.Root, n.a.cfg.Template
	now := time.Now()
	p2 := problem.Problem{Site: site, Name: id + "-" + name, Dir: dir}
	return func() tea.Msg {
		if _, _, err := problem.New(root, site, id, name, template, now); err != nil {
			return newCreatedMsg{Prob: p2, Err: err}
		}
		return newCreatedMsg{Prob: p2}
	}
}

func (n *newModel) view() string {
	a := n.a
	sites := strings.Join(n.knownSites(), ", ")
	if sites == "" {
		sites = "lc hr task mock"
	}
	var lines []string
	lines = append(lines, fieldRow(n.field == 0, "사이트", n.site.View()))
	lines = append(lines, stDimmer.Render("             "+truncateStr(sites, 44)))
	lines = append(lines, fieldRow(n.field == 1, "번호", n.id.View()))
	lines = append(lines, fieldRow(n.field == 2, "이름", n.name.View()))
	lines = append(lines, "")
	if !n.exists {
		lines = append(lines, stDim.Render("세 필드를 입력하고 마지막에 ⏎ — 편집기까지 열립니다"))
	}
	preview := n.dirPath()
	if n.site.Value() == "" || n.id.Value() == "" || n.name.Value() == "" {
		preview = "…"
	}
	lines = append(lines, stDim.Render("생성 경로  ")+stDimmer.Render(preview))
	if n.exists {
		lines = append(lines, stWarn.Render("기존 문제 — ⏎ 로 열기 (main.cpp 보존)"))
	}
	if n.err != "" {
		lines = append(lines, stFail.Render("⚠ "+n.err))
	}
	content := strings.Join(lines, "\n")
	w := 58
	if a.width-10 < w {
		w = a.width - 10
	}
	h := len(lines) + 5
	if a.height-4 < h {
		h = a.height - 4
	}
	return lipgloss.Place(a.width, a.height-2, lipgloss.Center, lipgloss.Center,
		paneBox("새 문제 · psnew", content, w, h, true))
}

func (n *newModel) helpText() string {
	return keyBar(
		[2]string{"⏎", "다음 필드/생성"}, [2]string{"tab", "필드 이동"}, [2]string{"esc", "취소"},
	)
}
