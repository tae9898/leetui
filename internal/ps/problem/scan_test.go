package problem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanHiddenExcludedAndNaturalSort(t *testing.T) {
	root := t.TempDir()
	mk := func(rel string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(root, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mk("lc/10-ten")
	mk("lc/2-two")
	mk("lc/.hidden-prob")
	mk("zz/a")
	mk(".git/obj")
	if err := os.WriteFile(filepath.Join(root, "plain.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []Problem{
		{Site: "lc", Name: "2-two", Dir: filepath.Join(root, "lc", "2-two")},
		{Site: "lc", Name: "10-ten", Dir: filepath.Join(root, "lc", "10-ten")},
		{Site: "zz", Name: "a", Dir: filepath.Join(root, "zz", "a")},
	}
	if len(got) != len(want) {
		t.Fatalf("문제 수 = %d, want %d (got %v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("항목 %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestScanMissingRoot(t *testing.T) {
	if _, err := Scan(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("존재하지 않는 root는 에러여야 한다")
	}
}

func TestNaturalCompare(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"2", "10", -1},
		{"10", "2", 1},
		{"2", "2", 0},
		{"1", "10", -1},
		{"a2b", "a10b", -1},
		{"a", "b", -1},
		{"b", "a", 1},
		{"a", "a1", -1},
		{"0056-merge", "1000-x", -1},
		{"9", "010", -1}, // 수치 동등하지 않음: 9 < 10
	}
	for _, c := range cases {
		if got := sign(naturalCompare(c.a, c.b)); got != c.want {
			t.Errorf("naturalCompare(%q, %q) 부호 = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	}
	return 0
}
