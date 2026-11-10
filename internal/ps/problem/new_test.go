package problem

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTemplate(t *testing.T, body string) string {
	t.Helper()
	tpl := filepath.Join(t.TempDir(), "template.cpp")
	if err := os.WriteFile(tpl, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return tpl
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestNewFreshDirectory(t *testing.T) {
	root := t.TempDir()
	tpl := writeTemplate(t, "// tpl\n")
	now := time.Unix(1700000000, 0)

	dir, existed, err := New(root, "lc", "0056", "merge", tpl, now)
	if err != nil {
		t.Fatal(err)
	}
	if existed {
		t.Error("새 디렉터리인데 existedMain=true")
	}
	if want := filepath.Join(root, "lc", "0056-merge"); dir != want {
		t.Errorf("dir = %q, want %q", dir, want)
	}
	if got := mustRead(t, filepath.Join(dir, "main.cpp")); got != "// tpl\n" {
		t.Errorf("main.cpp = %q, want 템플릿 그대로", got)
	}
	for _, f := range []string{"1.in", "1.out"} {
		st, err := os.Stat(filepath.Join(dir, f))
		if err != nil {
			t.Fatalf("%s 없음: %v", f, err)
		}
		if st.Size() != 0 {
			t.Errorf("%s는 빈 파일이어야 한다(크기 %d)", f, st.Size())
		}
	}
	if got := mustRead(t, filepath.Join(dir, ".start")); got != "1700000000\n" {
		t.Errorf(".start = %q, want %q", got, "1700000000\n")
	}
}

func TestNewPreservesExistingFiles(t *testing.T) {
	root := t.TempDir()
	tpl := writeTemplate(t, "tpl")
	now := time.Unix(1700000000, 0)
	dir, _, err := New(root, "hr", "1", "a", tpl, now)
	if err != nil {
		t.Fatal(err)
	}

	// 기존 산출물을 사용자가 수정한 상태
	if err := os.WriteFile(filepath.Join(dir, "main.cpp"), []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "1.in"), []byte("keep me\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	now2 := time.Unix(1700000123, 0)
	dir2, existed, err := New(root, "hr", "1", "a", tpl, now2)
	if err != nil {
		t.Fatal(err)
	}
	if dir2 != dir {
		t.Errorf("dir 변경됨: %q vs %q", dir2, dir)
	}
	if !existed {
		t.Error("main.cpp가 있는데 existedMain=false")
	}
	if got := mustRead(t, filepath.Join(dir, "main.cpp")); got != "mine" {
		t.Errorf("main.cpp 덮어씀: %q", got)
	}
	if got := mustRead(t, filepath.Join(dir, "1.in")); got != "keep me\n" {
		t.Errorf("1.in 잘림/수정됨: %q", got)
	}
	if got := mustRead(t, filepath.Join(dir, ".start")); got != "1700000123\n" {
		t.Errorf(".start = %q, want 갱신된 값", got)
	}
}

func TestNewMissingTemplate(t *testing.T) {
	root := t.TempDir()
	if _, _, err := New(root, "lc", "1", "a", filepath.Join(root, "absent.cpp"), time.Unix(1, 0)); err == nil {
		t.Fatal("템플릿이 없고 main.cpp도 없으면 에러여야 한다")
	}
}

func TestValidate(t *testing.T) {
	ok := []struct{ site, id, name string }{
		{"lc", "0056", "merge-intervals"},
		{"mock", "07", "kth"},
		{"a-b", "c_d", "e.f"},
	}
	for _, c := range ok {
		if err := Validate(c.site, c.id, c.name); err != nil {
			t.Errorf("Validate(%q,%q,%q) = %v, want nil", c.site, c.id, c.name, err)
		}
	}
	bad := []struct{ site, id, name string }{
		{"", "1", "a"}, {"lc", "", "a"}, {"lc", "1", ""},
		{"lc/..", "1", "a"}, {"lc", "1/2", "a"}, {"lc", "1", "a/b"},
		{"l c", "1", "a"}, {"lc", "1 2", "a"}, {"lc", "1", "a b"},
		{".hidden", "1", "a"}, {"lc", ".1", "a"}, {"lc", "1", ".a"},
	}
	for _, c := range bad {
		if err := Validate(c.site, c.id, c.name); err == nil {
			t.Errorf("Validate(%q,%q,%q) = nil, want 에러", c.site, c.id, c.name)
		}
	}
}
