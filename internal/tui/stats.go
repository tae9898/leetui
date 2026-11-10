package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"pstui/internal/ps/problem"
)

// statsModel: 유형별(헤더 "// 유형:") 통계 + 주차별 푼 수 (TUI-PLAN 3.2 Stats).
type statsModel struct {
	a *App
}

func newStats(a *App) *statsModel { return &statsModel{a: a} }

type typeStat struct {
	typ      string
	tries    int
	fails    int
	totalMin int
	avgMin   int
}

func (sm *statsModel) byType() []typeStat {
	m := map[string]*typeStat{}
	for _, e := range sm.a.entries {
		typ := "미분류"
		if p := sm.a.findProblem(e.Site, e.Problem); p != nil {
			if t := problem.ParseType(p.Dir); t != "" {
				typ = t
			}
		}
		st := m[typ]
		if st == nil {
			st = &typeStat{typ: typ}
			m[typ] = st
		}
		st.tries++
		st.totalMin += e.Minutes
		if e.Result == "fail" {
			st.fails++
		}
	}
	out := make([]typeStat, 0, len(m))
	for _, st := range m {
		st.avgMin = 0
		if st.tries > 0 {
			st.avgMin = st.totalMin / st.tries
		}
		out = append(out, *st)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].tries != out[j].tries {
			return out[i].tries > out[j].tries
		}
		return out[i].typ < out[j].typ
	})
	return out
}

// byWeek는 ISO 주 단위 시도/ok 수 (최근 5주).
func (sm *statsModel) byWeek() [][2]string {
	// 날짜 문자열의 주 계산: 월요일 시작. 간단히 날짜 → ISO week.
	m := map[string]*[3]int{}
	for _, e := range sm.a.entries {
		wk := isoWeekKey(e.Date)
		st := m[wk]
		if st == nil {
			st = &[3]int{}
			m[wk] = st
		}
		st[0]++
		if e.Result == "ok" || e.Result == "again" {
			st[1]++
		}
		st[2] += e.Minutes
	}
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) > 5 {
		keys = keys[len(keys)-5:]
	}
	out := make([][2]string, 0, len(keys))
	for _, k := range keys {
		st := m[k]
		out = append(out, [2]string{k, fmt.Sprintf("시도 %d · 해결 %d · %d분", st[0], st[1], st[2])})
	}
	return out
}

func (sm *statsModel) update(msg tea.Msg) tea.Cmd {
	if km, ok := msg.(tea.KeyMsg); ok {
		switch km.String() {
		case "esc":
			return func() tea.Msg { return openHomeMsg{} }
		}
	}
	return nil
}

func (sm *statsModel) view() string {
	var b strings.Builder
	b.WriteString(stTitle.Render(" 통계 · 유형별 (main.cpp 헤더 // 유형: 기준)"))
	b.WriteByte('\n')
	b.WriteString(stDim.Render("   유형                 시도  fail%  평균분"))
	b.WriteByte('\n')
	rows := sm.byType()
	if len(rows) == 0 {
		b.WriteString(stDim.Render(" 기록 없음"))
		b.WriteByte('\n')
	}
	for _, r := range rows {
		failPct := 0
		if r.tries > 0 {
			failPct = r.fails * 100 / r.tries
		}
		b.WriteString(fmt.Sprintf("   %-18s %4d  %4d%%  %5d\n", truncateStr(r.typ, 18), r.tries, failPct, r.avgMin))
	}
	b.WriteString(stTitle.Render(" 주차별 (최근 5주)"))
	b.WriteByte('\n')
	for _, w := range sm.byWeek() {
		b.WriteString(fmt.Sprintf("   %s  %s\n", w[0], w[1]))
	}
	b.WriteString(stDim.Render(" esc 홈"))
	b.WriteByte('\n')
	return b.String()
}

func (sm *statsModel) helpText() string {
	return "Stats: esc 홈"
}
