package problem

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// 2026-11-11 09:00 UTC — 어느 시간대에서도 월일이 같은 날(1111).
var retryNow = time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)

func TestRetryBackupAndReset(t *testing.T) {
	root := t.TempDir()
	tpl := writeTemplate(t, "fresh")
	dir := filepath.Join(root, "lc", "0001-a")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.cpp"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	backup, err := Retry(dir, tpl, retryNow)
	if err != nil {
		t.Fatal(err)
	}
	if backup != "main.1111.cpp" {
		t.Errorf("backup = %q, want %q", backup, "main.1111.cpp")
	}
	if got := mustRead(t, filepath.Join(dir, "main.1111.cpp")); got != "old" {
		t.Errorf("백업 내용 = %q, want %q", got, "old")
	}
	if got := mustRead(t, filepath.Join(dir, "main.cpp")); got != "fresh" {
		t.Errorf("main.cpp = %q, want 템플릿", got)
	}
	start, ok := ReadStart(dir)
	if !ok || start != retryNow.Unix() {
		t.Errorf("ReadStart = (%d, %v), want (%d, true)", start, ok, retryNow.Unix())
	}
}

func TestRetryOverwritesExistingBackup(t *testing.T) {
	root := t.TempDir()
	tpl := writeTemplate(t, "fresh")
	dir := filepath.Join(root, "lc", "0001-a")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.cpp"), []byte("newer"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.1111.cpp"), []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}

	backup, err := Retry(dir, tpl, retryNow)
	if err != nil {
		t.Fatal(err)
	}
	if backup != "main.1111.cpp" {
		t.Errorf("backup = %q, want %q", backup, "main.1111.cpp")
	}
	if got := mustRead(t, filepath.Join(dir, "main.1111.cpp")); got != "newer" {
		t.Errorf("백업 덮어쓰기 실패: %q", got)
	}
}

func TestRetryWithoutMain(t *testing.T) {
	root := t.TempDir()
	tpl := writeTemplate(t, "fresh")
	dir := filepath.Join(root, "lc", "0002-b")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	backup, err := Retry(dir, tpl, retryNow)
	if err != nil {
		t.Fatal(err)
	}
	if backup != "" {
		t.Errorf("main.cpp가 없으면 backup는 빈 문자열이어야 한다: %q", backup)
	}
	if got := mustRead(t, filepath.Join(dir, "main.cpp")); got != "fresh" {
		t.Errorf("main.cpp = %q, want 템플릿", got)
	}
	if _, ok := ReadStart(dir); !ok {
		t.Error(".start가 갱신되지 않았다")
	}
}
