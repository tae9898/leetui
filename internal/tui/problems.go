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
		case "tab":
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
	var b strings.Builder
	filter := pm.filterSite
	if filter == "" {
		filter = "전체"
	}
	b.WriteString(stTitle.Render(fmt.Sprintf(" 문제 목록 (%d) · 사이트: %s · 1~4/tab 필터 · d: due만 (%s) · /: 검색%s",
		len(rows), filter, boolMark(pm.dueOnly), searchMark(pm.searching, pm.search))))
	b.WriteByte('\n')
	b.WriteString(stDim.Render("   사이트  문제                 마지막        재도전     유형"))
	b.WriteByte('\n')
	for i, r := range rows {
		marker, style := " ", stNormal
		if i == pm.cursor {
			marker, style = ">", stSel
		}
		lastStr, dueStr := "-", "-"
		if r.last != nil {
			lastStr = fmt.Sprintf("%s %s", r.last.Date, r.last.Result)
			dueStr = r.last.Due
		}
		progress := " "
		if r.inProgress {
			progress = "▶"
		}
		if r.ghost {
			style = stDim
		}
		line := fmt.Sprintf("%s %s%-6s %-19s %-12s %-10s %s%s",
			marker, progress, r.site, r.name, lastStr, dueStr, r.typ, ghostMark(r.ghost))
		b.WriteString(style.Render(line))
		b.WriteByte('\n')
	}
	return b.String()
}

func (pm *problemsModel) helpText() string {
	return "Problems: j/k 이동 · Enter 작업 · / 검색 · tab/1~4 사이트 필터 · d due만 · n 새 문제 · esc 홈"
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
