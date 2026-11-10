package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"pstui/internal/ps/problem"
)

type FileEntry struct {
	Name string
	Size int64
	Kind int // 0 main, 1 in, 2 out, 3 backup
	Case string // in/out 쌍의 케이스 번호 문자열
}

type tickMsg struct{}

type workbenchModel struct {
	a       *App
	prob    problem.Problem
	files   []FileEntry
	cursor  int
	err     string
	summary string // 마지막 실행 결과 (세션 내 유지)
	typ     string
	header  [3]string
}

func newWorkbench(a *App) *workbenchModel { return &workbenchModel{a: a} }

func (w *workbenchModel) setProblem(p problem.Problem) {
	w.prob = p
	w.cursor = 0
	w.err = ""
	w.typ = problem.ParseType(p.Dir)
	w.header = readHeaderLines(p.Dir)
}

func readHeaderLines(dir string) [3]string {
	var h [3]string
	data, err := os.ReadFile(filepath.Join(dir, "main.cpp"))
	if err != nil {
		return h
	}
	prefixes := []string{"// 문제:", "// 유형:", "// 목표 복잡도:"}
	for _, line := range strings.Split(string(data), "\n") {
		for i, p := range prefixes {
			if h[i] == "" && strings.HasPrefix(line, p) {
				h[i] = strings.TrimSpace(strings.TrimPrefix(line, p))
			}
		}
	}
	return h
}

func (w *workbenchModel) reload(p problem.Problem) tea.Cmd {
	w.prob = p
	dir := p.Dir
	return func() tea.Msg {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return filesLoadedMsg{Err: err}
		}
		var files []FileEntry
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			info, _ := e.Info()
			n := e.Name()
			f := FileEntry{Name: n}
			if info != nil {
				f.Size = info.Size()
			}
			switch {
			case n == "main.cpp":
				f.Kind = 0
			case strings.HasSuffix(n, ".in"):
				f.Kind = 1
				f.Case = strings.TrimSuffix(n, ".in")
			case strings.HasSuffix(n, ".out"):
				f.Kind = 2
				f.Case = strings.TrimSuffix(n, ".out")
			case strings.HasPrefix(n, "main.") && strings.HasSuffix(n, ".cpp"):
				f.Kind = 3
			default:
				continue
			}
			files = append(files, f)
		}
		sort.Slice(files, func(i, j int) bool {
			ki, kj := fileSortKey(files[i]), fileSortKey(files[j])
			if ki != kj {
				return ki < kj
			}
			return natLess(files[i].Case, files[j].Case)
		})
		return filesLoadedMsg{Files: files}
	}
}

func fileSortKey(f FileEntry) int {
	if f.Kind == 0 {
		return 0
	}
	if f.Kind == 1 {
		return 1
	}
	if f.Kind == 2 {
		return 2
	}
	return 3
}

// natLess는 자연 정렬 비교 ("2" < "10").
func natLess(a, b string) bool {
	for len(a) > 0 && len(b) > 0 {
		na, da := leadNum(a)
		nb, db := leadNum(b)
		if na != "" && nb != "" {
			x, _ := strconv.Atoi(na)
			y, _ := strconv.Atoi(nb)
			if x != y {
				return x < y
			}
			a, b = da, db
			continue
		}
		if a[0] != b[0] {
			return a[0] < b[0]
		}
		a, b = a[1:], b[1:]
	}
	return len(a) < len(b)
}

func leadNum(s string) (num, rest string) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	return s[:i], s[i:]
}

func (w *workbenchModel) update(msg tea.Msg) tea.Cmd {
	switch m := msg.(type) {
	case filesLoadedMsg:
		if m.Err != nil {
			w.err = m.Err.Error()
		} else {
			w.files = m.Files
			w.err = ""
		}
		return w.tickCmd()
	case tickMsg:
		return w.tickCmd()
	case startResetMsg:
		if m.Err != nil {
			w.err = m.Err.Error()
		}
		return w.reload(w.prob)
	case appendDoneMsg:
		return w.reload(w.prob)

	case tea.KeyMsg:
		switch m.String() {
		case "up", "k":
			if w.cursor > 0 {
				w.cursor--
			}
		case "down", "j":
			if w.cursor < len(w.files)-1 {
				w.cursor++
			}
		case "e":
			return w.a.editorCmd(w.prob, "main.cpp")
		case "r":
			p := w.prob
			return func() tea.Msg { return openRunMsg{Prob: p} }
		case "F":
			p := w.prob
			return func() tea.Msg { return openRunMsg{Prob: p, Fast: true} }
		case "d":
			p := w.prob
			return func() tea.Msg { return openDoneMsg{Prob: p} }
		case "t":
			p := w.prob
			return func() tea.Msg { return openRetryMsg{Prob: p} }
		case "a":
			return w.addCase()
		case "o":
			if f := w.selected(); f != nil && f.Kind == 2 {
				return w.a.editorCmd(w.prob, f.Name)
			}
		case "S":
			if _, ok := problem.ReadStart(w.prob.Dir); !ok {
				return w.resetStart()
			}
		case "esc":
			return func() tea.Msg { return openHomeMsg{} }
		}
	}
	return nil
}

func (w *workbenchModel) selected() *FileEntry {
	if w.cursor < len(w.files) {
		return &w.files[w.cursor]
	}
	return nil
}

func (w *workbenchModel) addCase() tea.Cmd {
	n := 1
	for _, f := range w.files {
		if f.Kind == 1 {
			if v, err := strconv.Atoi(f.Case); err == nil && v >= n {
				n = v + 1
			}
		}
	}
	name := strconv.Itoa(n) + ".in"
	for _, s := range []string{name, strconv.Itoa(n) + ".out"} {
		if f, err := os.OpenFile(filepath.Join(w.prob.Dir, s), os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			f.Close()
		}
	}
	w.reloadSync()
	return w.a.editorCmd(w.prob, name)
}

func (w *workbenchModel) resetStart() tea.Cmd {
	dir := w.prob.Dir
	return func() tea.Msg {
		return startResetMsg{Err: problem.WriteStart(dir, time.Now())}
	}
}

func (w *workbenchModel) reloadSync() {
	cmd := w.reload(w.prob)
	if cmd != nil {
		if msg := cmd(); msg != nil {
			w.update(msg)
		}
	}
}

func (w *workbenchModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Minute, func(time.Time) tea.Msg { return tickMsg{} })
}

func (w *workbenchModel) elapsed() string {
	start, ok := problem.ReadStart(w.prob.Dir)
	if !ok {
		return "-"
	}
	return fmt.Sprintf("%d분", problem.ElapsedMinutes(start, time.Now()))
}

func (w *workbenchModel) view() string {
	var b strings.Builder
	b.WriteString(stTitle.Render(fmt.Sprintf(" %s/%s", w.prob.Site, w.prob.Name)))
	b.WriteString(stDim.Render(fmt.Sprintf("  · 경과 %s · %s", w.elapsed(), w.prob.Dir)))
	b.WriteByte('\n')
	b.WriteString(stDim.Render(fmt.Sprintf(" 문제: %s · 유형: %s · 목표: %s", w.header[0], w.header[1], w.header[2])))
	b.WriteByte('\n')
	if w.err != "" {
		b.WriteString(stFail.Render("⚠ " + w.err))
		b.WriteByte('\n')
	}
	if w.summary != "" {
		b.WriteString(stDim.Render(" 마지막 실행: " + w.summary))
		b.WriteByte('\n')
	}
	b.WriteString(w.filesPane())
	b.WriteByte('\n')
	b.WriteString(w.historyPane())
	return b.String()
}

func (w *workbenchModel) filesPane() string {
	var b strings.Builder
	b.WriteString(stTitle.Render(" 파일"))
	b.WriteByte('\n')
	for i, f := range w.files {
		marker, style := " ", stNormal
		if i == w.cursor {
			marker, style = ">", stSel
		}
		empty := ""
		if f.Size == 0 {
			empty = stWarn.Render("(빈 파일)")
		}
		b.WriteString(style.Render(fmt.Sprintf("%s %-22s %6dB %s", marker, f.Name, f.Size, empty)))
		b.WriteByte('\n')
	}
	return b.String()
}

func (w *workbenchModel) historyPane() string {
	var b strings.Builder
	b.WriteString(stTitle.Render(" 이력"))
	b.WriteByte('\n')
	hist := w.a.problemEntries(w.prob.Site, w.prob.Name)
	if len(hist) == 0 {
		b.WriteString(stDim.Render(" 기록 없음"))
		b.WriteByte('\n')
	}
	for i := len(hist) - 1; i >= 0; i-- {
		e := hist[i]
		b.WriteString(fmt.Sprintf("%s %-6s %3d분 due %s %s\n",
			e.Date, e.Result, e.Minutes, e.Due, stDim.Render(e.Note)))
	}
	return b.String()
}

func (w *workbenchModel) helpText() string {
	return "Workbench: e 편집 · r 실행 · F fast빌드 · d 기록 · t 재도전 · a 케이스추가 · o .out편집 · S .start재설정 · esc 홈"
}

// setSummary는 Run 화면 종료 후 요약을 남긴다.
func (w *workbenchModel) setSummary(binName string, all, pass, elapsedMin int) {
	w.summary = fmt.Sprintf("%s: %d/%d PASS · 경과 %d분", binName, pass, all, elapsedMin)
}
