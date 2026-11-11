package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"leetui/internal/ps/lcref"
	"leetui/internal/ps/problem"
)

// stmtModel: 지문 뷰어. LeetCode는 GraphQL로 지문을 가져오고 문제 디렉터리에 캐시한다.
type stmtModel struct {
	a       *App
	prob    problem.Problem
	slug    string
	site    string
	loaded  bool
	loading bool
	err     string
	title   string
	diff    string
	text    string
	url     string
	ex      []lcref.Example
	filled  int // 마지막 a 키로 채운 케이스 수
	vp      viewport.Model
	vpOK    bool
}

func newStmt(a *App) *stmtModel { return &stmtModel{a: a} }

func (sm *stmtModel) open(p problem.Problem, site, slug string) tea.Cmd {
	sm.prob = p
	sm.site = site
	sm.slug = slug
	sm.loaded = false
	sm.loading = site == "lc"
	sm.err = ""
	sm.title, sm.diff, sm.text, sm.url = "", "", "", ""
	sm.ex = nil
	sm.filled = 0
	sm.vp = viewport.New(0, 0)
	sm.vpOK = false
	var cmd tea.Cmd
	switch site {
	case "lc":
		sm.url = "https://leetcode.com/problems/" + slug + "/"
		cmd = sm.loadCmd(false)
	case "hr":
		sm.loading = false
		sm.title = p.Name
		sm.url = "https://www.hackerrank.com/challenges/" + slug + "/problem"
		sm.text = "HackerRank 지문은 웹에서 보세요.\n\n" + sm.url + "\n\n(슬러그는 표시명 기준 추정 — 열리지 않으면 사이트에서 검색)"
		sm.loaded = true
	default: // task
		sm.loading = false
		sm.title, sm.text, sm.loaded = taskStatement(slug)
	}
	// 초기 WindowSizeMsg은 앱 시작(홈) 때 한 번만 오므로 진입 시점 크기로 뷰포트를 잡아야 스크롤이 동작한다
	sm.resizeVP()
	return cmd
}

func (sm *stmtModel) cachePath() string {
	return filepath.Join(sm.prob.Dir, ".stmt.html")
}

func (sm *stmtModel) loadCmd(refetch bool) tea.Cmd {
	slug, dir := sm.slug, sm.prob.Dir
	return func() tea.Msg {
		if !refetch {
			if data, err := os.ReadFile(filepath.Join(dir, ".stmt.html")); err == nil && len(data) > 0 {
				return stmtLoadedMsg{Slug: slug, BodyHTML: string(data), Cached: true}
			}
		}
		st, err := lcref.FetchStatement(context.Background(), slug)
		if err != nil {
			return stmtLoadedMsg{Slug: slug, Err: err}
		}
		_ = os.WriteFile(filepath.Join(dir, ".stmt.html"), []byte(st.BodyHTML), 0644)
		return stmtLoadedMsg{Slug: slug, Title: st.Title, Difficulty: st.Difficulty, BodyHTML: st.BodyHTML}
	}
}

func (sm *stmtModel) update(msg tea.Msg) tea.Cmd {
	if !sm.vpOK {
		sm.resizeVP() // 진입 시 크기 확보에 실패했으면(터미널이 잠시 작았을 때) 재시도
	}
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		sm.resizeVP()
		return nil

	case stmtLoadedMsg:
		if m.Slug != sm.slug {
			return nil
		}
		sm.loading = false
		if m.Err != nil {
			sm.err = m.Err.Error()
			sm.loaded = true
			return nil
		}
		if m.Title != "" {
			sm.title, sm.diff = m.Title, m.Difficulty
		}
		sm.ex = lcref.ParseExamples(m.BodyHTML)
		sm.text = lcref.HTMLToText(m.BodyHTML)
		sm.loaded = true
		sm.refresh()
		return nil

	case tea.KeyMsg:
		if sm.loading {
			return nil
		}
		switch m.String() {
		case "j":
			sm.vp.LineDown(1)
		case "k":
			sm.vp.LineUp(1)
		case "g":
			sm.vp.GotoTop()
		case "G":
			sm.vp.GotoBottom()
		case "a":
			return sm.fillCases()
		case "R":
			sm.loading = true
			sm.err = ""
			return sm.loadCmd(true)
		case "esc":
			p := sm.prob
			if p.Dir != "" && sm.a.bench.prob.Dir == p.Dir {
				return func() tea.Msg { return openWorkbenchMsg{Prob: p} }
			}
			return func() tea.Msg { return openHomeMsg{} }
		default:
			vm, cmd := sm.vp.Update(msg)
			sm.vp = vm
			return cmd
		}
	}
	return nil
}

// fillCases는 예제를 .in/.out으로 쓴다. 비어 있는 기존 케이스 우선, 없으면 다음 번호 생성.
func (sm *stmtModel) fillCases() tea.Cmd {
	if sm.site != "lc" || len(sm.ex) == 0 || sm.prob.Dir == "" {
		return nil
	}
	if _, err := os.Stat(sm.prob.Dir); err != nil {
		// 아직 시작하지 않은 문제 — 이 자리에서 스캐폴딩 후 채운다
		if !sm.a.cfg.TemplateOK() {
			return nil
		}
		id, name := splitProbName(sm.prob.Name, sm.slug)
		if _, _, err := problem.New(sm.a.cfg.Root, sm.prob.Site, id, name, sm.a.cfg.Template, time.Now()); err != nil {
			return nil
		}
	}
	n := 1
	filled := 0
	for _, ex := range sm.ex {
		for {
			in := filepath.Join(sm.prob.Dir, strconv.Itoa(n)+".in")
			out := filepath.Join(sm.prob.Dir, strconv.Itoa(n)+".out")
			inEmpty := fileEmpty(in)
			outEmpty := fileEmpty(out)
			if inEmpty && outEmpty {
				os.WriteFile(in, []byte(ex.In+"\n"), 0644)
				os.WriteFile(out, []byte(ex.Out+"\n"), 0644)
				filled++
				n++
				break
			}
			if !inEmpty && !outEmpty {
				n++ // 이미 채워진 케이스 — 다음 번호로
				continue
			}
			n++ // 절반만 찬 케이스 — 건너뜀
			break
		}
	}
	sm.filled = filled
	p := sm.prob
	return tea.Batch(
		func() tea.Msg { return openWorkbenchMsg{Prob: p} },
		sm.a.bench.reload(p),
	)
}

// splitProbName은 "0001-two-sum" → ("0001", "two-sum").
func splitProbName(full, fallback string) (id, name string) {
	if i := strings.IndexByte(full, '-'); i > 0 {
		return full[:i], full[i+1:]
	}
	return "0000", fallback
}

func fileEmpty(p string) bool {
	st, err := os.Stat(p)
	return err != nil || st.Size() == 0
}

func (sm *stmtModel) resizeVP() {
	w := sm.a.width - 6
	h := sm.a.height - 2 - 1 - 3
	if w < 20 || h < 3 {
		return
	}
	sm.vp.Width = w
	sm.vp.Height = h
	sm.vpOK = true
	sm.refresh()
}

func (sm *stmtModel) refresh() {
	if !sm.vpOK {
		return
	}
	sm.vp.SetContent(sm.content())
	sm.vp.GotoTop()
}

func (sm *stmtModel) content() string {
	var b strings.Builder
	if sm.title != "" {
		b.WriteString(stTitle.Render(sm.title))
		if sm.diff != "" {
			d := stOK
			if sm.diff == "Medium" {
				d = stWarn
			} else if sm.diff == "Hard" {
				d = stFail
			}
			b.WriteString("  " + d.Render(sm.diff))
		}
		b.WriteString("\n")
	}
	if len(sm.ex) > 0 && sm.prob.Dir != "" {
		b.WriteString(stDim.Render(fmt.Sprintf("예제 %d개 · a 키로 .in/.out 자동 채움\n", len(sm.ex))))
	}
	b.WriteString("\n")
	b.WriteString(sm.text)
	if sm.filled > 0 {
		b.WriteString("\n\n" + stOK.Render(fmt.Sprintf("→ 예제 %d개를 케이스로 채웠습니다", sm.filled)))
	}
	return b.String()
}

func (sm *stmtModel) view() string {
	a := sm.a
	paneH := a.height - 2 - 1
	if paneH < 5 {
		paneH = 5
	}
	title := "지문 · " + sm.slug
	if sm.loading {
		title += " · 가져오는 중…"
	}
	if sm.err != "" {
		title += " · 실패 (R 재시도)"
	}
	body := sm.vp.View()
	if !sm.vpOK {
		body = sm.text
	}
	return paneBox(title, body, a.width, paneH, true)
}

func (sm *stmtModel) helpText() string {
	keys := [][2]string{
		[2]string{"j/k", "스크롤"}, [2]string{"u/d", "반 페이지"}, [2]string{"g/G", "처음/끝"},
	}
	if sm.site == "lc" && len(sm.ex) > 0 {
		keys = append(keys, [2]string{"a", "예제→케이스"})
	}
	keys = append(keys, [2]string{"R", "다시 받기"}, [2]string{"esc", "뒤로"})
	return keyBar(keys...)
}

// taskStatement는 plan.md 7절의 과제형 요구사항을 반환한다.
func taskStatement(slug string) (string, string, bool) {
	common := "\n\n규칙 (plan.md 7절)\n· 40분 제한, assert 기반 테스트를 같은 파일에 작성\n· sanitizer를 켠 상태로 통과\n· C와 C++ 두 버전으로 한 번씩"
	switch slug {
	case "ring-buffer":
		return "링버퍼", "요구사항\n· 고정 크기, push/pop\n· 가득 찼을 때 정책 선택: 거부 또는 덮어쓰기\n\n확인 포인트\n· full과 empty 구분\n· 인덱스 wrap\n· 단일 생산자-단일 소비자에서의 스레드 안전성 설명" + common, true
	case "memory-pool":
		return "고정 블록 메모리 풀", "요구사항\n· 정적 배열 위에서 alloc/free\n· 동적 할당 없음\n\n확인 포인트\n· free list 구성\n· 이중 해제와 범위 밖 포인터 검출\n· 정렬(alignment)" + common, true
	case "protocol-parser":
		return "프로토콜 파서", "요구사항\n· 헤더 · 길이 · 페이로드 · 체크섬 구조의 바이트 스트림 파싱\n\n확인 포인트\n· 상태머신 구성\n· 조각난 입력 처리\n· 길이 필드 검증\n· 엔디안" + common, true
	}
	return "과제", "plan.md 7절 참고", true
}
