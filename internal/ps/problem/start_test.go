package problem

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStartRoundTrip(t *testing.T) {
	dir := t.TempDir()
	now := time.Unix(1699999999, 0)

	if _, ok := ReadStart(dir); ok {
		t.Error(".start가 없으면 false여야 한다")
	}
	if err := WriteStart(dir, now); err != nil {
		t.Fatal(err)
	}
	got, ok := ReadStart(dir)
	if !ok {
		t.Fatal("쓴 직후 읽기 실패")
	}
	if got != now.Unix() {
		t.Errorf("ReadStart = %d, want %d", got, now.Unix())
	}
	if body := mustRead(t, filepath.Join(dir, ".start")); body != "1699999999\n" {
		t.Errorf(".start 바이트 = %q, want %q", body, "1699999999\n")
	}
}

func TestReadStartBroken(t *testing.T) {
	for _, body := range []string{"", "abc", "12\n34\n", "9.5", "0x10"} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, ".start"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, ok := ReadStart(dir); ok {
			t.Errorf(".start=%q: 깨진 파일은 false여야 한다", body)
		}
	}
}

func TestReadStartTrimsSpaces(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".start"), []byte(" 123 \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, ok := ReadStart(dir)
	if !ok || got != 123 {
		t.Errorf("ReadStart = (%d, %v), want (123, true)", got, ok)
	}
}

func TestElapsedMinutes(t *testing.T) {
	now := time.Unix(10_000, 0)
	cases := []struct {
		start int64
		want  int
	}{
		{10_000 - 0, 0},
		{10_000 - 30, 0},
		{10_000 - 60, 1},
		{10_000 - 119, 1},
		{10_000 - 120, 2},
		{10_000 + 30, 0},  // 미래 시작이면 음수 → 0
		{10_000 + 600, 0}, // -10이어도 0
	}
	for _, c := range cases {
		if got := ElapsedMinutes(c.start, now); got != c.want {
			t.Errorf("ElapsedMinutes(%d) = %d, want %d", c.start, got, c.want)
		}
	}
}
