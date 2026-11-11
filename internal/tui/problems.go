package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"pstui/internal/ps/problem"
	"pstui/internal/ps/pslog"
)

type problemsModel struct {
	a          *App
	cursor     int
	filterSite string // "" = 전체
	searching  bool
	search     string
	dueOnly    bool
}

func newProblems(a *App) *problemsModel { return &problemsModel{a: a} }

type probRow struct {
	prob       *problem.Problem
	site, name string
	last       *pslog.Entry
	ghost      bool // 로그에만 있고 디렉터리 없음
	inProgress bool
	typ        string
}

func (pm *problemsModel) rows() []probRow {
	now := time.Now()
	dueSet := map[string]bool{}
	for _, d := range pm.a.dueList() {
		dueSet[d.Key] = true
	}
	seen := map[string]bool{}
	var rows []probRow
	for i := range pm.a.problems {
		p := &pm.a.problems[i]
		seen[p.Site+"/"+p.Name] = true
		rows = append(rows, pm.makeRow(p, p.Site, p.Name, now, dueSet))
	}
	for i := range pm.a.entries {
		e := pm.a.entries[i]
		key := e.Site + "/" + e.Problem
		if !seen[key] {
			seen[key] = true
			rows = append(rows, pm.makeRow(nil, e.Site, e.Problem, now, dueSet))
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].site != rows[j].site {
			return rows[i].site < rows[j].site
		}
		return rows[i].name < rows[j].name
	})
	return rows
}

func (pm *problemsModel) makeRow(p *problem.Problem, site, name string, now time.Time, dueSet map[string]bool) probRow {
	r := probRow{prob: p, site: site, name: name}
	if le := pm.a.lastEntry(site, name); le != nil {
		r.last = le
	}
	if p != nil {
		if start, ok := problem.ReadStart(p.Dir); ok && now.Unix()-start < 86400 {
			r.inProgress = true
		}
		r.typ = problem.ParseType(p.Dir)
	}
	return r
}

func (pm *problemsModel) visible() []probRow {
	var out []probRow
	for _, r := range pm.rows() {
		if pm.filterSite != "" && r.site != pm.filterSite {
			continue
		}
		if pm.dueOnly {
			key := r.site + "/" + r.name
			last := pm.a.lastEntry(r.site, r.name)
			if last == nil || last.Due == "-" || last.Due > time.Now().Format("2006-01-02") {
				_ = key
				continue
			}
		}
		if pm.searching && pm.search != "" {
			if !strings.Contains(r.site+"/"+r.name, pm.search) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

func (pm *problemsModel) sites() []string {
	set := map[string]bool{}
	var out []string
	for _, p := range pm.a.problems {
		if !set[p.Site] {
			set[p.Site] = true
			out = append(out, p.Site)
		}
	}
	sort.Strings(out)
	return out
}

func (pm *problemsModel) update(msg tea.Msg) tea.Cmd {
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
			return nil
		}
		rows := pm.visible()
		switch m.String() {
		case "n":
			return func() tea.Msg { return openNewMsg{} }
		case "/":
			pm.searching, pm.search = true, ""
		case "d":
			pm.dueOnly = !pm.dueOnly
		case "tab", "l":
			ss := pm.sites()
			for i, s := range ss {
				if s == pm.filterSite {
					pm.filterSite = ss[(i+1)%len(ss)]
					return nil
				}
			}
			if len(ss) > 0 {
				pm.filterSite = ss[0]
			}
		case "h":
			ss := pm.sites()
			if pm.filterSite == "" {
				if len(ss) > 0 {
					pm.filterSite = ss[len(ss)-1]
				}
			} else {
				for i, s := range ss {
					if s == pm.filterSite {
						pm.filterSite = ss[(i+len(ss)-1)%len(ss)]
						return nil
					}
				}
			}
		case "up", "k":
			if pm.cursor > 0 {
				pm.cursor--
			}
		case "down", "j":
			if pm.cursor < len(rows)-1 {
				pm.cursor++
			}
		case "enter":
			if pm.cursor < len(rows) {
				r := rows[pm.cursor]
				if r.prob != nil {
					p := *r.prob
					return func() tea.Msg { return openWorkbenchMsg{Prob: p} }
				}
			}
		case "esc":
			return func() tea.Msg { return openHomeMsg{} }
		}
		if n, ok := siteDigitKey(m.String(), pm.sites()); ok {
			pm.filterSite = n
		}
	}
	return nil
}

func siteDigitKey(s string, sites []string) (string, bool) {
	if len(s) != 1 || s[0] < '1' || s[0] > '4' {
		return "", false
	}
	i := int(s[0]-'1') - 1 // 1=전체, 2~4=사이트
	if i == -1 {
		return "", true
	}
	if i < len(sites) {
		return sites[i], true
	}
	return "", false
}

func (pm *problemsModel) view() string {
	rows := pm.visible()
	if pm.cursor >= len(rows) {
		pm.cursor = maxInt(len(rows)-1, 0)
	}
	a := pm.a
	paneH := a.height - 2
	if paneH < 5 {
		paneH = 5
	}
	filter := pm.filterSite
	if filter == "" {
		filter = "전체"
	}
	title := fmt.Sprintf("문제 · %d · %s", len(rows), filter)
	if pm.dueOnly {
		title += " · due만"
	}
	if pm.searching {
		title += " · 검색: " + pm.search + "_"
	}
	innerW := a.width - 4
	return paneBox(title, pm.rowsText(rows, innerW), a.width, paneH, true)
}

func (pm *problemsModel) rowsText(rows []probRow, innerW int) string {
	if innerW < 40 {
		return ""
	}
	siteW, lastW, dueW, typW := 6, 17, 6, 10
	nameW := innerW - siteW - lastW - dueW - typW - 1
	var b strings.Builder
	b.WriteString(stDimmer.Render(
		padW("사이트", siteW) + padW("문제", nameW) + padW("마지막", lastW) +
			padW("due", dueW) + "유형"))
	b.WriteByte('\n')
	for i, r := range rows {
		lastStr, dueStr := stDimmer.Render(padW("-", lastW-1)), "-"
		if r.last != nil {
			res := resultStyle(r.last.Result).Render(padW(string(r.last.Result), 5))
			lastStr = padW(shortDate(r.last.Date), 11) + res
			dueStr = shortDate(r.last.Due)
		}
		progress := " "
		if r.inProgress {
			progress = stWarn.Render("▶")
		}
		name := truncateStr(r.name, nameW-2)
		if r.ghost {
			name = stDim.Render(name) + stDimmer.Render("†")
		}
		row := progress + padW(r.site, siteW-1) + padW(name, nameW) +
			lastStr + padW(dueStr, dueW) + truncateStr(r.typ, typW)
		b.WriteString(selRow(row, innerW, i == pm.cursor))
		if i < len(rows)-1 {
			b.WriteByte('\n')
		}
	}
	if len(rows) == 0 {
		b.WriteString(stDim.Render(" 해당 문제 없음"))
	}
	return b.String()
}

func (pm *problemsModel) helpText() string {
	return keyBar(
		[2]string{"j/k", "이동"}, [2]string{"⏎", "작업 열기"}, [2]string{"/", "검색"},
		[2]string{"h/l", "사이트"}, [2]string{"d", "due만"}, [2]string{"n", "새 문제"},
		[2]string{"esc", "홈"},
	)
}

func boolMark(b bool) string {
	if b {
		return "켬"
	}
	return "끔"
}

func searchMark(on bool, q string) string {
	if on {
		return " · 검색어: " + q + "_"
	}
	return ""
}

func ghostMark(g bool) string {
	if g {
		return " (디렉터리 없음)"
	}
	return ""
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
