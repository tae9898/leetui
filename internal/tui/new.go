package tui

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"leetui/internal/ps/lcref"
	"leetui/internal/ps/problem"
)

// newModel: 새 문제 폼. 첫 필드에 URL이나 "56. Merge Intervals"를 붙여넣으면
// 나머지 필드를 자동으로 채운다. 번호는 LeetCode GraphQL로 조회.
type newModel struct {
	a          *App
	paste      textinput.Model
	site       textinput.Model
	id         textinput.Model
	name       textinput.Model
	field      int
	err        string
	exists     bool   // 대상 디렉터리가 이미 있음
	lookupSlug string // 번호 조회 진행 중인 슬러그 (중복 발송 방지)
}

func newNew(a *App) *newModel {
	n := &newModel{a: a}
	n.paste = textinput.New()
	n.paste.Prompt = ""
	n.paste.Placeholder = "URL 또는 56. Merge Intervals 붙여넣기"
	n.paste.CharLimit = 200
	n.site = textinput.New()
	n.site.Prompt = ""
	n.site.Placeholder = "lc"
	n.site.CharLimit = 20
	n.id = textinput.New()
	n.id.Prompt = ""
	n.id.Placeholder = "0056"
	n.id.CharLimit = 12
	n.name = textinput.New()
	n.name.Prompt = ""
	n.name.Placeholder = "merge-intervals"
	n.name.CharLimit = 60
	return n
}

func (n *newModel) reset() {
	n.paste.SetValue("")
	n.site.SetValue("")
	n.id.SetValue("")
	n.name.SetValue("")
	n.field = 0
	n.err = ""
	n.exists = false
	n.lookupSlug = ""
	n.focusActive()
}

// focusActive는 현재 필드만 포커스한다 (포커스 없는 textinput은 입력을 무시한다).
func (n *newModel) focusActive() {
	n.paste.Blur()
	n.site.Blur()
	n.id.Blur()
	n.name.Blur()
	switch n.field {
	case 0:
		n.paste.Focus()
	case 1:
		n.site.Focus()
	case 2:
		n.id.Focus()
	case 3:
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
	case lcNumberMsg:
		if m.Slug == n.lookupSlug && n.id.Value() == "" && m.Err == nil {
			n.id.SetValue(lcref.PadNumber(m.Number))
		}
		if m.Slug == n.lookupSlug {
			n.lookupSlug = ""
		}
		return nil

	case tea.KeyMsg:
		switch m.String() {
		case "tab":
			n.field = (n.field + 1) % 4
			n.focusActive()
			return nil
		case "shift+tab":
			n.field = (n.field + 3) % 4
			n.focusActive()
			return nil
		case "enter":
			if n.field < 3 {
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
			active = &n.paste
		case 1:
			active = &n.site
		case 2:
			active = &n.id
		case 3:
			active = &n.name
		}
		if active != nil {
			cmd := textinputBubbling(active, m)
			n.refreshExists()
			if n.field == 0 {
				if apply := n.applyPaste(); apply != nil {
					return tea.Batch(cmd, apply)
				}
			}
			return cmd
		}
	}
	return nil
}

func (n *newModel) refreshExists() {
	_, statErr := os.Stat(n.dirPath())
	n.exists = statErr == nil && n.id.Value() != "" && n.name.Value() != ""
}

// applyPaste는 붙여넣기 필드를 파싱해 나머지 필드를 채운다.
// 번호를 모르는 LeetCode 참조면 조회 cmd를 반환한다.
func (n *newModel) applyPaste() tea.Cmd {
	ref, ok := lcref.Parse(n.paste.Value())
	if !ok {
		return nil
	}
	if ref.Site != "" {
		n.site.SetValue(ref.Site)
	}
	if ref.Name != "" {
		n.name.SetValue(ref.Name)
	}
	if ref.ID != "" {
		n.id.SetValue(ref.ID)
		n.lookupSlug = ""
		return nil
	}
	// 번호 없음: lc면 슬러그로 조회 (hr은 번호 관례가 없어 그대로)
	slug := n.name.Value()
	if n.site.Value() == "lc" && slug != "" && slug != n.lookupSlug {
		n.lookupSlug = slug
		return func() tea.Msg {
			num, err := lcref.LookupNumber(context.Background(), slug)
			return lcNumberMsg{Slug: slug, Number: num, Err: err}
		}
	}
	return nil
}

func (n *newModel) submit() tea.Cmd {
	site, id, name := strings.TrimSpace(n.site.Value()), strings.TrimSpace(n.id.Value()), strings.TrimSpace(n.name.Value())
	if site == "" || id == "" || name == "" {
		n.err = "사이트/번호/이름이 필요합니다 — 위 필드에 붙여넣으면 자동으로 채워집니다"
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
	var lines []string
	lines = append(lines, fieldRow(n.field == 0, "붙여넣기", n.paste.View()))
	lines = append(lines, stDimmer.Render("             URL · \"56. Merge Intervals\" · 슬러그 모두 인식"))
	lines = append(lines, fieldRow(n.field == 1, "사이트", n.site.View()))
	lines = append(lines, fieldRow(n.field == 2, "번호", n.id.View()+lookupMark(n)))
	lines = append(lines, fieldRow(n.field == 3, "이름", n.name.View()))
	lines = append(lines, "")
	preview := n.dirPath()
	if n.site.Value() == "" || n.id.Value() == "" || n.name.Value() == "" {
		preview = "…"
	}
	lines = append(lines, stDim.Render("생성 경로  ")+stDimmer.Render(preview))
	if n.exists {
		lines = append(lines, stWarn.Render("기존 문제 — ⏎ 로 열기 (main.cpp 보존)"))
	} else {
		lines = append(lines, stDim.Render("마지막 필드에서 ⏎ — 생성 후 편집기까지 열립니다"))
	}
	if n.err != "" {
		lines = append(lines, stFail.Render("⚠ "+n.err))
	}
	content := strings.Join(lines, "\n")
	w := 68
	if a.width-10 < w {
		w = a.width - 10
	}
	h := len(lines) + 5
	if a.height-4 < h {
		h = a.height - 4
	}
	content = fitWidth(content, w-6)
	return lipgloss.Place(a.width, a.height-2, lipgloss.Center, lipgloss.Center,
		paneBox("새 문제 · psnew", content, w, h, true))
}

func lookupMark(n *newModel) string {
	if n.lookupSlug != "" {
		return " " + stTitle.Render("(번호 조회 중…)")
	}
	return ""
}

func (n *newModel) helpText() string {
	return keyBar(
		[2]string{"⏎", "다음 필드/생성"}, [2]string{"붙여넣기", "자동 채움"}, [2]string{"esc", "취소"},
	)
}
