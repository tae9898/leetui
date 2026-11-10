package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

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
	d.memo.Placeholder = "막힌 지점 한 줄"
	d.memo.CharLimit = 200
	d.mins = textinput.New()
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
			}
		case "down", "j", "tab":
			if d.field < 2 {
				d.field++
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
				d.memo.Focus()
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
	var b strings.Builder
	b.WriteString(stTitle.Render(fmt.Sprintf(" 결과 기록 · %s/%s", d.prob.Site, d.prob.Name)))
	b.WriteByte('\n')

	b.WriteString(fieldLabel(d.field == 0, "결과"))
	for i, r := range d.results {
		mark := " "
		if i == d.cur {
			mark = "●"
		}
		style := stNormal
		if i == d.cur {
			style = stOK
		}
		b.WriteString(" " + style.Render(mark+" "+r))
	}
	b.WriteString(stDim.Render("  ←/→ 또는 1~3"))
	b.WriteByte('\n')

	b.WriteString(fieldLabel(d.field == 1, "메모") + d.memo.View())
	b.WriteByte('\n')

	minsLabel := "소요분"
	if d.minsAuto {
		minsLabel += " (.start 자동)"
	} else {
		minsLabel += " (.start 없음 — 직접 입력)"
	}
	b.WriteString(fieldLabel(d.field == 2, minsLabel) + d.mins.View())
	b.WriteByte('\n')

	if e, err := d.entry(); err == nil {
		preview := strings.ReplaceAll(string(e.Marshal()), "\t", "⇥")
		b.WriteString(stDim.Render(" 미리보기: " + strings.TrimSpace(preview)))
		b.WriteByte('\n')
		b.WriteString(stDim.Render(" 재도전일: " + e.Due))
		b.WriteByte('\n')
	} else {
		b.WriteString(stFail.Render(" " + err.Error()))
		b.WriteByte('\n')
	}
	if d.err != "" {
		b.WriteString(stFail.Render("⚠ " + d.err))
		b.WriteByte('\n')
	}
	today := time.Now().Format("2006-01-02")
	for _, en := range d.a.problemEntries(d.prob.Site, d.prob.Name) {
		if en.Date == today {
			b.WriteString(stWarn.Render(" 오늘 이미 기록 있음 (fish처럼 누적 허용)"))
			b.WriteByte('\n')
			break
		}
	}
	b.WriteString(stDim.Render(" enter 저장 · esc 취소"))
	return b.String()
}

func fieldLabel(active bool, name string) string {
	if active {
		return stSel.Render(fmt.Sprintf(" %-12s", name))
	}
	return fmt.Sprintf(" %-12s", name)
}

func (d *doneModel) helpText() string {
	return "Done: ←/→ 결과 · j/k 필드 이동 · enter 저장 · esc 취소"
}
