package problem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeMain(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "main.cpp"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestParseType(t *testing.T) {
	dir := t.TempDir()
	writeMain(t, dir, "// 문제: 두 수 합\n// 유형: DP, 해시\n// 목표 복잡도: O(N)\nint main() {}\n")
	if got := ParseType(dir); got != "DP, 해시" {
		t.Errorf("ParseType = %q, want %q", got, "DP, 해시")
	}
}

func TestParseTypeTrimsValue(t *testing.T) {
	dir := t.TempDir()
	writeMain(t, dir, "  // 유형:   그리디 \nint main() {}\n")
	if got := ParseType(dir); got != "그리디" {
		t.Errorf("ParseType = %q, want %q", got, "그리디")
	}
}

func TestParseTypeOnTenthLine(t *testing.T) {
	dir := t.TempDir()
	body := strings.Repeat("// 채움\n", 9) + "// 유형: 스택\nint main() {}\n"
	writeMain(t, dir, body)
	if got := ParseType(dir); got != "스택" {
		t.Errorf("10번째 줄 마커 = %q, want %q", got, "스택")
	}
}

func TestParseTypeBeyondTenLinesIgnored(t *testing.T) {
	dir := t.TempDir()
	body := strings.Repeat("// 채움\n", 10) + "// 유형: 스택\n"
	writeMain(t, dir, body)
	if got := ParseType(dir); got != "" {
		t.Errorf("10줄 밖 마커는 무시해야 한다: %q", got)
	}
}

func TestParseTypeNoMarkerOrFile(t *testing.T) {
	dir := t.TempDir()
	writeMain(t, dir, "#include <bits/stdc++.h>\nint main() {}\n")
	if got := ParseType(dir); got != "" {
		t.Errorf("마커 없으면 빈 문자열: %q", got)
	}
	if got := ParseType(filepath.Join(t.TempDir(), "none")); got != "" {
		t.Errorf("파일이 없어도 빈 문자열이어야 한다: %q", got)
	}
}
