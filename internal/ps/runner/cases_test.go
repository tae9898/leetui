package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCasesNaturalSortAndOutPairing(t *testing.T) {
	dir := t.TempDir()
	files := map[string]bool{
		"1.in": true, "1.out": true,
		"2.in":  true, // .out 없음
		"10.in": true, "10.out": true,
		"notes.txt": false, ".hidden.in": false,
	}
	for name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "d.in"), 0o755); err != nil {
		t.Fatal(err)
	}

	cs, err := Cases(dir)
	if err != nil {
		t.Fatal(err)
	}
	wantNames := []string{"1.in", "2.in", "10.in"}
	wantOut := []bool{true, false, true}
	if len(cs) != len(wantNames) {
		t.Fatalf("케이스 수 = %d, want %d (%v)", len(cs), len(wantNames), cs)
	}
	for i, c := range cs {
		if got := filepath.Base(c.In); got != wantNames[i] {
			t.Errorf("케이스 %d 이름 = %q, want %q", i, got, wantNames[i])
		}
		if c.HasOut != wantOut[i] {
			t.Errorf("케이스 %d HasOut = %v, want %v", i, c.HasOut, wantOut[i])
		}
		if c.HasOut != (c.OutPath != "") {
			t.Errorf("케이스 %d: HasOut과 OutPath 불일치", i)
		}
	}
}

func TestCasesEmptyAndMissingDir(t *testing.T) {
	dir := t.TempDir()
	cs, err := Cases(dir)
	if err != nil || len(cs) != 0 {
		t.Errorf("빈 디렉터리: %v %v", cs, err)
	}
	if _, err := Cases(filepath.Join(dir, "absent")); err == nil {
		t.Error("없는 디렉터리는 에러여야 한다")
	}
}
