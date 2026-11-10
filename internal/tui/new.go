package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

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
	n.site.Placeholder = "lc"
	n.site.CharLimit = 20
	n.id = textinput.New()
	n.id.Placeholder = "0056"
	n.id.CharLimit = 12
	n.name = textinput.New()
	n.name.Placeholder = "merge-intervals"
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
			return nil
		case "shift+tab":
			n.field = (n.field + 2) % 3
			return nil
		case "enter":
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
	var b strings.Builder
	b.WriteString(stTitle.Render(" 새 문제"))
	b.WriteString(stDim.Render(" · psnew lc 0056 merge-intervals 에 해당"))
	b.WriteByte('\n')
	fields := []struct {
		label string
		ti    *textinput.Model
	}{
		{"사이트", &n.site},
		{"번호", &n.id},
		{"이름", &n.name},
	}
	for i, f := range fields {
		b.WriteString(fieldLabel(n.field == i, f.label))
		if i == 0 {
			b.WriteString(stDim.Render(" (" + strings.Join(n.knownSites(), ", ") + ")"))
			b.WriteByte('\n')
			b.WriteString(strings.Repeat(" ", 14))
		}
		b.WriteString(f.ti.View())
		b.WriteByte('\n')
	}
	preview := n.dirPath()
	if n.site.Value() == "" || n.id.Value() == "" || n.name.Value() == "" {
		preview = "…"
	}
	b.WriteString(stDim.Render(" 경로: " + preview))
	if n.exists {
		b.WriteString(stWarn.Render("  (기존 문제 — Enter로 열기, main.cpp 보존)"))
	}
	b.WriteByte('\n')
	if n.err != "" {
		b.WriteString(stFail.Render("⚠ " + n.err))
		b.WriteByte('\n')
	}
	b.WriteString(stDim.Render(" tab 필드 이동 · enter 생성(후 편집기) · esc 취소"))
	return b.String()
}

func (n *newModel) helpText() string {
	return "New: tab 필드 · enter 생성 · esc 취소"
}
