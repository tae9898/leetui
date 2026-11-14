package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"leetui/internal/ps/config"
	"leetui/internal/ps/problem"
	"leetui/internal/ps/runner"
)

// 컴파일 실패 시 로그 탭으로 전환되고 로그가 보여야 한다 (Tab 눌러도 안 보였던 회귀).
func TestRunScreenCompileFailShowsLog(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.cpp"), []byte("int main(){ int x = ; }"), 0644); err != nil {
		t.Fatal(err)
	}
	a := newApp(config.Config{Root: dir, Editor: "true", CXX: "g++", Template: filepath.Join(dir, "template.cpp")})
	a.width, a.height = 120, 40

	p := problem.Problem{Site: "lc", Name: "0001-x", Dir: dir}
	a.open(scrRun)
	a.runM.start(p, false, "")

	msg := a.runM.compileCmd()() // 실제 g++ 실행
	if cd, ok := msg.(compileDoneMsg); !ok {
		t.Fatalf("compileDoneMsg 아님: %T", msg)
	} else if cd.Err == nil {
		t.Fatal("깨진 코드인데 컴파일 성공?")
	}

	if cmd := a.runM.update(msg); cmd != nil {
		t.Fatalf("실패 분기에서 cmd 반환하면 안 됨: %v", cmd)
	}
	if a.runM.tab != 2 {
		t.Fatalf("tab=%d, 로그 탭(2)으로 전환되어야 함", a.runM.tab)
	}
	view := a.runM.view()
	for _, want := range []string{"컴파일 실패", "컴파일 로그", "error:"} {
		if !strings.Contains(view, want) {
			t.Errorf("뷰에 %q 없음:\n%s", want, view)
		}
	}
}

// diff가 비어 있을 때(출력 0바이트 / .out 빈 파일) 원인 힌트가 보여야 한다.
func TestDiffPaneEmptyOutputHints(t *testing.T) {
	a := newApp(config.Config{Root: t.TempDir(), Editor: "true", CXX: "g++"})
	a.width, a.height = 120, 40
	a.open(scrRun)
	rm := a.runM
	rm.start(problem.Problem{Dir: "/tmp"}, false, "")
	rm.tab = 0
	rm.vpOK = false // detailContent 직접 검증

	rm.results = []runner.CaseResult{{Status: runner.Fail, Expected: []byte("1\n")}}
	if s := rm.detailContent(); !strings.Contains(s, "아무것도 출력 안 함") {
		t.Errorf("출력 없음 힌트 없음: %q", s)
	}

	rm.results = []runner.CaseResult{{Status: runner.Pass}}
	if s := rm.detailContent(); !strings.Contains(s, ".out)이 비어 있음") {
		t.Errorf("빈 .out 힌트 없음: %q", s)
	}
}
