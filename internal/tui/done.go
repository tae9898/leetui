package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"pstui/internal/ps/problem"
	"pstui/internal/ps/pslog"
)

type doneModel struct {
	a        *App
	prob     problem.Problem
	results  []string // {"ok","fail","again"}
	cur      int
	memo     textinput.Model
	mins     textinput.Model
	minsAuto bool
	field    int // 0 결과, 1 메모, 2 소요분
	err      string
}

func newDone(a *App) *doneModel {
	d := &doneModel{a: a, results: []string{"ok", "fail", "again"}}
	d.memo = textinput.New()
	d.memo.Prompt = ""
	d.memo.Placeholder = "막힌 지점 한 줄"
	d.memo.CharLimit = 200
	d.mins = textinput.New()
	d.mins.Prompt = ""
	d.mins.Placeholder = "분"
	d.mins.CharLimit = 5
	return d
}

func (d *doneModel) start(p problem.Problem, def string) {
	d.prob = p
	d.cur = 1
	if def == "ok" {
		d.cur = 0
	} else if def == "again" {
		d.cur = 2
	}
	d.memo.SetValue("")
	d.mins.SetValue("")
	d.field = 0
	d.err = ""
	if start, ok := problem.ReadStart(p.Dir); ok {
		d.minsAuto = true
		d.mins.SetValue(fmt.Sprintf("%d", problem.ElapsedMinutes(start, time.Now())))
	} else {
		d.minsAuto = false
	}
	d.focusActive()
}

// focusActive는 메모/소요분 필드만 포커스한다 (결과 라디오는 키 직접 처리).
func (d *doneModel) focusActive() {
	d.memo.Blur()
	d.mins.Blur()
	if d.field == 1 {
		d.memo.Focus()
	}
	if d.field == 2 {
		d.mins.Focus()
	}
}

func (d *doneModel) result() pslog.Result { return pslog.Result(d.results[d.cur]) }

func (d *doneModel) entry() (pslog.Entry, error) {
	e := pslog.Entry{
		Date:    time.Now().Format("2006-01-02"),
		Site:    d.prob.Site,
		Problem: d.prob.Name,
		Result:  d.result(),
		Due:     pslog.DueDate(d.result(), time.Now()),
		Note:    d.memo.Value(),
	}
	mins := 0
	if v := strings.TrimSpace(d.mins.Value()); v != "" {
		if _, err := fmt.Sscanf(v, "%d", &mins); err != nil {
			return e, fmt.Errorf("소요분이 숫자가 아님: %q", v)
		}
	}
	e.Minutes = mins
	return e, nil
}

func (d *doneModel) update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case appendDoneMsg:
		return d.updateAfterAppend(m)
	case tea.KeyMsg:
		switch m.String() {
		case "up", "k":
			if d.field > 0 {
				d.field--
				d.focusActive()
			}
		case "down", "j", "tab":
			if d.field < 2 {
				d.field++
				d.focusActive()
			}
		case "left", "h":
			if d.field == 0 && d.cur > 0 {
				d.cur--
			}
		case "right", "l":
			if d.field == 0 && d.cur < len(d.results)-1 {
				d.cur++
			}
		case "1", "2", "3":
			if d.field == 0 {
				d.cur = int(m.String()[0] - '1')
			}
		case "enter":
			if d.field == 2 {
				return d.save()
			}
			if d.field == 0 {
				d.field = 1
				d.focusActive()
				return nil
			}
			return d.save()
		case "esc":
			p := d.prob
			return func() tea.Msg { return openWorkbenchMsg{Prob: p} }
		}
		if d.field == 1 {
			return textinputBubbling(&d.memo, m)
		}
		if d.field == 2 {
			return textinputBubbling(&d.mins, m)
		}
	}
	return nil
}

func textinputBubbling(ti *textinput.Model, m tea.KeyMsg) tea.Cmd {
	mm, cmd := ti.Update(m)
	*ti = mm
	return cmd
}

func (d *doneModel) save() tea.Cmd {
	e, err := d.entry()
	if err != nil {
		d.err = err.Error()
		return nil
	}
	path := filepath.Join(d.a.cfg.Root, "log.tsv")
	return func() tea.Msg {
		return appendDoneMsg{Err: pslog.Append(path, e)}
	}
}

func (d *doneModel) updateAfterAppend(msg appendDoneMsg) tea.Cmd {
	if msg.Err != nil {
		d.err = msg.Err.Error()
		return nil
	}
	p := d.prob
	return tea.Batch(loadLogCmd(d.a.cfg.Root), func() tea.Msg { return openWorkbenchMsg{Prob: p} })
}

func (d *doneModel) view() string {
	a := d.a
	res := make([]string, 0, 3)
	for i, r := range d.results {
		if i == d.cur {
			res = append(res, stKey.Render("● "+r))
		} else {
			res = append(res, stDim.Render("○ "+r))
		}
	}
	var lines []string
	lines = append(lines, fieldRow(d.field == 0, "결과", strings.Join(res, "   ")+stDimmer.Render("   ←/→ 또는 1~3")))
	lines = append(lines, fieldRow(d.field == 1, "메모", d.memo.View()))
	minsLabel := "소요분"
	if d.minsAuto {
		minsLabel += " (자동)"
	} else {
		minsLabel += " (직접 입력)"
	}
	lines = append(lines, fieldRow(d.field == 2, minsLabel, d.mins.View()))
	if e, err := d.entry(); err == nil {
		preview := strings.ReplaceAll(strings.TrimSpace(string(e.Marshal())), "\t", stDimmer.Render("⇥"))
		lines = append(lines, "")
		lines = append(lines, stDim.Render("저장될 행  ")+stDimmer.Render(preview))
		lines = append(lines, stDim.Render("재도전일   ")+dueStyle(e.Due).Render(e.Due))
	} else {
		lines = append(lines, stFail.Render(err.Error()))
	}
	today := time.Now().Format("2006-01-02")
	for _, en := range d.a.problemEntries(d.prob.Site, d.prob.Name) {
		if en.Date == today {
			lines = append(lines, stWarn.Render("오늘 이미 기록 있음 — fish처럼 누적 허용"))
			break
		}
	}
	if d.err != "" {
		lines = append(lines, stFail.Render("⚠ "+d.err))
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
		paneBox("결과 기록 · "+d.prob.Site+"/"+d.prob.Name, content, w, h, true))
}

func dueStyle(due string) lipgloss.Style {
	if due == "-" {
		return stOK
	}
	return stWarn
}

func fieldRow(active bool, label, value string) string {
	mark := "  "
	lab := stDim.Render(padW(label, 12))
	if active {
		mark = stKey.Render("▸ ")
		lab = stTitle.Render(padW(label, 12))
	}
	return mark + lab + value
}

func (d *doneModel) helpText() string {
	return keyBar(
		[2]string{"h/l", "결과"}, [2]string{"j/k", "필드"}, [2]string{"1~3", "결과 선택"},
		[2]string{"⏎", "저장"}, [2]string{"esc", "취소"},
	)
}
