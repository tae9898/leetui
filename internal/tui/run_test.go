package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"leetui/internal/ps/config"
	"leetui/internal/ps/problem"
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
