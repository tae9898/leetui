package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"pstui/internal/ps/plan"
	"pstui/internal/ps/problem"
)

// pickerModel: plan.md 5절의 문제 목록에서 골라 바로 시작한다.
type pickerModel struct {
	a         *App
	cursor    int
	week      int // -1 전체, 0 여유, 1~4
	searching bool
	search    string
	unsolved  bool
}

func newPicker(a *App) *pickerModel { return &pickerModel{a: a, week: -1} }

type itemStatus struct {
	label string
	due   string
}

// status는 디렉터리와 로그로 항목 상태를 계산한다.
func (pm *pickerModel) status(it plan.Item) (itemStatus, bool) {
	key := it.ID + "-" + it.Name
	if le := pm.a.lastEntry(it.Site, key); le != nil {
		switch le.Result {
		case "ok":
			return itemStatus{"완료", ""}, true
		case "again":
			return itemStatus{"재확인 대기", le.Due}, true
		default:
			return itemStatus{"재도전 예정", le.Due}, true
		}
	}
	if pm.a.findProblem(it.Site, key) != nil {
		return itemStatus{"진행 중", ""}, true
	}
	return itemStatus{"미풀", ""}, false
}

func (pm *pickerModel) visible() []plan.Item {
	var out []plan.Item
	for _, it := range plan.All() {
		if pm.week >= 0 && it.Week != pm.week {
			continue
		}
		if pm.unsolved {
			if st, solved := pm.status(it); solved && st.label == "완료" {
				continue
			}
		}
		if pm.search != "" {
			q := pm.search
			if !strings.Contains(it.ID+" "+it.Name+" "+it.Title+" "+it.Type, q) {
				continue
			}
		}
		out = append(out, it)
	}
	return out
}

func (pm *pickerModel) update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if pm.searching {
			switch m.String() {
			case "esc":
				pm.searching, pm.search = false, ""
			case "enter":
				pm.searching = false
			case "backspace":
				if n := len(pm.search); n > 0 {
					pm.search = pm.search[:n-1]
				}
			default:
				if len(m.Runes) > 0 {
					pm.search += string(m.Runes)
				}
			}
			pm.cursor = 0
			return nil
		}
		items := pm.visible()
		switch m.String() {
		case "j", "down":
			if pm.cursor < len(items)-1 {
				pm.cursor++
			}
		case "k", "up":
			if pm.cursor > 0 {
				pm.cursor--
			}
		case "l", "tab":
			pm.week = nextWeek(pm.week)
			pm.cursor = 0
		case "h":
			pm.week = prevWeek(pm.week)
			pm.cursor = 0
		case "1", "2", "3", "4":
			if int(m.String()[0]-'0') == pm.week {
				pm.week = -1
			} else {
				pm.week = int(m.String()[0] - '0')
			}
			pm.cursor = 0
		case "0":
			if pm.week == 0 {
				pm.week = -1
			} else {
				pm.week = 0
			}
			pm.cursor = 0
		case "u":
			pm.unsolved = !pm.unsolved
			pm.cursor = 0
		case "/":
			pm.searching, pm.search = true, ""
		case "m":
			return func() tea.Msg { return openNewMsg{} }
		case "v":
			if pm.cursor < len(items) {
				it := items[pm.cursor]
				p := problem.Problem{Site: it.Site, Name: it.ID + "-" + it.Name,
					Dir: pm.a.cfg.Root + "/" + it.Site + "/" + it.ID + "-" + it.Name}
				return func() tea.Msg { return openStmtMsg{Prob: p, Site: it.Site, Slug: it.Name} }
			}
		case "enter":
			if pm.cursor < len(items) {
				return pm.start(items[pm.cursor])
			}
		case "esc":
			return func() tea.Msg { return openHomeMsg{} }
		}
	}
	return nil
}

// start는 항목의 문제 디렉터리를 열거나 만들고 편집기를 띄운다.
func (pm *pickerModel) start(it plan.Item) tea.Cmd {
	a := pm.a
	key := it.ID + "-" + it.Name
	if p := a.findProblem(it.Site, key); p != nil {
		p2 := *p
		return func() tea.Msg { return openWorkbenchMsg{Prob: p2} }
	}
	if !a.cfg.TemplateOK() {
		return func() tea.Msg { return openNewMsg{} }
	}
	p2 := problem.Problem{Site: it.Site, Name: key, Dir: a.cfg.Root + "/" + it.Site + "/" + key}
	root, template := a.cfg.Root, a.cfg.Template
	return func() tea.Msg {
		if _, _, err := problem.New(root, it.Site, it.ID, it.Name, template, time.Now()); err != nil {
			return openNewMsg{}
		}
		return openWorkbenchMsg{Prob: p2}
	}
}

func nextWeek(w int) int {
	switch w {
	case -1:
		return 1
	case 4:
		return 0
	case 0:
		return -1
	default:
		return w + 1
	}
}

func prevWeek(w int) int {
	switch w {
	case -1:
		return 0
	case 1:
		return 4
	case 0:
		return 4
	default:
		return w - 1
	}
}

func (pm *pickerModel) weekLabel() string {
	if pm.week == -1 {
		return "전체"
	}
	return plan.WeekLabel(pm.week)
}

func (pm *pickerModel) view() string {
	a := pm.a
	paneH := a.height - 2
	if paneH < 5 {
		paneH = 5
	}
	items := pm.visible()
	if pm.cursor >= len(items) {
		pm.cursor = maxInt(len(items)-1, 0)
	}
	unsolved := 0
	solvedAll := 0
	for _, it := range plan.All() {
		st := func() itemStatus { s, _ := pm.status(it); return s }()
		if st.label == "완료" {
			solvedAll++
			continue
		}
		unsolved++
	}
	title := fmt.Sprintf("플랜 · %d/%d항 · %s", len(items), len(plan.All()), pm.weekLabel())
	if pm.unsolved {
		title += " · 미풀만"
	}
	if pm.searching {
		title += " · 검색: " + pm.search + "_"
	} else if pm.search != "" {
		title += " · '" + pm.search + "'"
	}
	title += fmt.Sprintf(" · 완료 %d · 남음 %d", solvedAll, unsolved)
	return paneBox(title, pm.rowsText(items, a.width-4), a.width, paneH, true)
}

func (pm *pickerModel) rowsText(items []plan.Item, innerW int) string {
	if innerW < 60 {
		return ""
	}
	wkW, typeW, idW, stW, noteW := 4, 10, 6, 12, 14
	titleW := innerW - wkW - typeW - idW - stW - noteW - 4
	var b strings.Builder
	b.WriteString(stDimmer.Render(padW("주", wkW) + padW("유형", typeW) +
		padW("번호", idW) + padW("문제", titleW) + padW("상태", stW) + "핵심"))
	b.WriteByte('\n')
	for i, it := range items {
		st, _ := pm.status(it)
		var stCell string
		switch {
		case st.label == "완료":
			stCell = stOK.Render(padW("✓ 완료", stW))
		case st.label == "진행 중":
			stCell = stTitle.Render(padW(st.label, stW))
		case st.label == "미풀":
			stCell = stDim.Render(padW(st.label, stW))
		default:
			stCell = stWarn.Render(padW(st.label, stW))
		}
		row := padW(plan.WeekLabel(it.Week), wkW) +
			padW(truncateStr(it.Type, typeW-1), typeW) +
			padW(it.ID, idW) +
			padW(truncateStr(it.Title, titleW-1), titleW) +
			stCell + stDimmer.Render(truncateStr(it.Note, noteW))
		b.WriteString(selRow(row, innerW, i == pm.cursor))
		if i < len(items)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (pm *pickerModel) helpText() string {
	return keyBar(
		[2]string{"j/k", "이동"}, [2]string{"h/l", "주차"}, [2]string{"u", "미풀만"},
		[2]string{"/", "검색"}, [2]string{"v", "지문"}, [2]string{"⏎", "시작"}, [2]string{"m", "수동 입력"}, [2]string{"esc", "홈"},
	)
}
