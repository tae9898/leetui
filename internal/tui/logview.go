package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// logModel: log.tsv 전체 뷰어. 최신이 위, 읽기 전용.
type logModel struct {
	a         *App
	cursor    int
	searching bool
	search    string
	filter    byte // 0 전체, 'o'/'f'/'a'
	offsetTop int  // 스크롤 상단 행
}

func newLog(a *App) *logModel { return &logModel{a: a} }

func (lm *logModel) reset() {
	lm.cursor = 0
	lm.searching = false
	lm.search = ""
	lm.filter = 0
	lm.offsetTop = 0
}

func (lm *logModel) rows() []int { // entries 인덱스 (최신=마지막 행부터)
	var out []int
	for i := len(lm.a.entries) - 1; i >= 0; i-- {
		e := lm.a.entries[i]
		if lm.filter == 'o' && e.Result != "ok" {
			continue
		}
		if lm.filter == 'f' && e.Result != "fail" {
			continue
		}
		if lm.filter == 'a' && e.Result != "again" {
			continue
		}
		if lm.search != "" && !strings.Contains(e.Site+"/"+e.Problem+" "+e.Note, lm.search) {
			continue
		}
		out = append(out, i)
	}
	return out
}

func (lm *logModel) visibleRows() int { return lm.a.height - 6 }

func (lm *logModel) update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case tea.KeyMsg:
		if lm.searching {
			switch m.String() {
			case "esc":
				lm.searching, lm.search = false, ""
			case "enter":
				lm.searching = false
			case "backspace":
				if n := len(lm.search); n > 0 {
					lm.search = lm.search[:n-1]
				}
			default:
				if len(m.Runes) > 0 {
					lm.search += string(m.Runes)
				}
			}
			return nil
		}
		rows := lm.rows()
		switch m.String() {
		case "/":
			lm.searching, lm.search = true, ""
		case "o", "f", "a":
			if lm.filter == m.String()[0] {
				lm.filter = 0
			} else {
				lm.filter = m.String()[0]
			}
			lm.cursor = 0
		case "up", "k":
			if lm.cursor > 0 {
				lm.cursor--
			}
		case "down", "j":
			if lm.cursor < len(rows)-1 {
				lm.cursor++
			}
		case "esc":
			return func() tea.Msg { return openHomeMsg{} }
		}
		lm.fixScroll(len(rows))
	}
	return nil
}

func (lm *logModel) fixScroll(n int) {
	h := lm.visibleRows()
	if h < 1 {
		h = 1
	}
	if lm.cursor < lm.offsetTop {
		lm.offsetTop = lm.cursor
	}
	if lm.cursor >= lm.offsetTop+h {
		lm.offsetTop = lm.cursor - h + 1
	}
	if lm.offsetTop+h > n {
		lm.offsetTop = n - h
		if lm.offsetTop < 0 {
			lm.offsetTop = 0
		}
	}
}

func (lm *logModel) view() string {
	rows := lm.rows()
	if lm.cursor >= len(rows) {
		lm.cursor = maxInt(len(rows)-1, 0)
	}
	lm.fixScroll(len(rows))

	var b strings.Builder
	filter := "전체"
	if lm.filter != 0 {
		filter = string(lm.filter)
	}
	b.WriteString(stTitle.Render(fmt.Sprintf(" 로그 (%d행) · 필터 %s · o/f/a 결과 · / 검색%s · 손상 행 %d",
		len(rows), filter, searchMark(lm.searching, lm.search), lm.a.skipped)))
	b.WriteByte('\n')
	b.WriteString(stDim.Render("   날짜       사이트  문제                 결과   분  재도전일    메모"))
	b.WriteByte('\n')
	end := lm.offsetTop + lm.visibleRows()
	if end > len(rows) {
		end = len(rows)
	}
	for vi := lm.offsetTop; vi < end; vi++ {
		e := lm.a.entries[rows[vi]]
		marker, style := " ", stNormal
		if vi == lm.cursor {
			marker, style = ">", stSel
		}
		b.WriteString(style.Render(fmt.Sprintf("%s %s %-6s %-19s %-6s %3d %-11s %s",
			marker, e.Date, e.Site, e.Problem, e.Result, e.Minutes, e.Due, truncateStr(e.Note, 30))))
		b.WriteByte('\n')
	}
	return b.String()
}

func (lm *logModel) helpText() string {
	return "Log: j/k 이동 · o/f/a 필터 · / 검색 · esc 홈 (읽기 전용)"
}
