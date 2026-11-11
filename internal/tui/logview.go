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
		case "l":
			lm.filter = nextFilter(lm.filter)
			lm.cursor = 0
		case "h":
			lm.filter = prevFilter(lm.filter)
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
	a := lm.a
	paneH := a.height - 2
	if paneH < 5 {
		paneH = 5
	}
	filter := "전체"
	if lm.filter != 0 {
		filter = string(lm.filter)
	}
	title := fmt.Sprintf("로그 · %d행 · %s · 손상 %d", len(rows), filter, a.skipped)
	if lm.searching {
		title += " · 검색: " + lm.search + "_"
	} else if lm.search != "" {
		title += " · 필터: " + lm.search
	}
	return paneBox(title, lm.rowsText(rows, a.width-4), a.width, paneH, true)
}

func (lm *logModel) helpText() string {
	return keyBar(
		[2]string{"j/k", "이동"}, [2]string{"h/l", "결과 필터"},
		[2]string{"/", "검색"}, [2]string{"esc", "홈"},
	)
}

func (lm *logModel) rowsText(rows []int, innerW int) string {
	if innerW < 60 {
		return ""
	}
	dateW, siteW, resW, minW, dueW := 6, 5, 6, 4, 6
	probW := innerW - dateW - siteW - resW - minW - dueW - 24
	var b strings.Builder
	b.WriteString(stDimmer.Render(padW("날짜", dateW) + padW("사이트", siteW) +
		padW("문제", probW) + padW("결과", resW) + padW("분", minW+1) + padW("due", dueW) + "메모"))
	b.WriteByte('\n')
	end := lm.offsetTop + lm.visibleRows()
	if end > len(rows) {
		end = len(rows)
	}
	for vi := lm.offsetTop; vi < end; vi++ {
		e := lm.a.entries[rows[vi]]
		res := resultStyle(e.Result).Render(padW(string(e.Result), resW))
		note := truncateStr(e.Note, 24)
		row := padW(shortDate(e.Date), dateW) + padW(e.Site, siteW) +
			padW(truncateStr(e.Problem, probW-1), probW) + res +
			padW(fmt.Sprintf("%d", e.Minutes), minW) + padW(shortDate(e.Due), dueW) + note
		b.WriteString(selRow(row, innerW, vi == lm.cursor))
		if vi < end-1 {
			b.WriteByte('\n')
		}
	}
	if len(rows) == 0 {
		b.WriteString(stDim.Render(" 기록 없음"))
	}
	return b.String()
}

// 결과 필터 순환: 없음 → ok → fail → again → 없음
func nextFilter(cur byte) byte {
	switch cur {
	case 0:
		return 'o'
	case 'o':
		return 'f'
	case 'f':
		return 'a'
	default:
		return 0
	}
}

func prevFilter(cur byte) byte {
	switch cur {
	case 0:
		return 'a'
	case 'a':
		return 'f'
	case 'f':
		return 'o'
	default:
		return 0
	}
}
