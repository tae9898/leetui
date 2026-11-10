package runner

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func needGXX(t *testing.T) string {
	t.Helper()
	cxx, err := exec.LookPath("g++")
	if err != nil {
		t.Skip("PATH에 g++가 없음")
	}
	return cxx
}

// copyProblem은 testdata 문제를 임시 디렉터리로 복사해 빌드 산출물이
// 저장소를 오염시키지 않게 한다.
func copyProblem(t *testing.T, name string) string {
	t.Helper()
	src := filepath.Join("testdata", name)
	dst := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dst
}

// CLI-ENV.md 4절과 문자 단위 동일해야 한다.
func TestFlagConstants(t *testing.T) {
	if SanitizeFlags != "-std=c++17 -O2 -g -Wall -Wextra -Wshadow -fsanitize=address,undefined -D_GLIBCXX_DEBUG" {
		t.Errorf("SanitizeFlags = %q", SanitizeFlags)
	}
	if FastFlags != "-std=c++17 -O2" {
		t.Errorf("FastFlags = %q", FastFlags)
	}
	if CaseTimeout != 2*time.Second {
		t.Errorf("CaseTimeout = %v", CaseTimeout)
	}
	if MaxOutput != 1<<20 {
		t.Errorf("MaxOutput = %d", MaxOutput)
	}
}

func TestCompileCreatesSol(t *testing.T) {
	cxx := needGXX(t)
	dir := copyProblem(t, "pass")
	stderr, err := Compile(dir, cxx)
	if err != nil {
		t.Fatalf("컴파일 실패: %v (stderr: %q)", err, stderr)
	}
	if st, serr := os.Stat(filepath.Join(dir, "sol")); serr != nil || st.IsDir() {
		t.Errorf("sol 바이너리가 없음: %v", serr)
	}
}

func TestFastBuildCreatesFast(t *testing.T) {
	cxx := needGXX(t)
	dir := copyProblem(t, "pass")
	_, elapsed, err := FastBuild(dir, cxx)
	if err != nil {
		t.Fatal(err)
	}
	if elapsed <= 0 {
		t.Errorf("경과 시간 = %v, want > 0", elapsed)
	}
	if st, serr := os.Stat(filepath.Join(dir, "fast")); serr != nil || st.IsDir() {
		t.Errorf("fast 바이너리가 없음: %v", serr)
	}
}

func TestCompileErrorCapturesStderr(t *testing.T) {
	cxx := needGXX(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.cpp"), []byte("int main( {\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stderr, err := Compile(dir, cxx)
	if err == nil {
		t.Fatal("컴파일 실패 시 err가 있어야 한다")
	}
	if !strings.Contains(stderr, "error") {
		t.Errorf("stderr에 컴파일 오류가 없음: %q", stderr)
	}
}

func TestRunCasePass(t *testing.T) {
	cxx := needGXX(t)
	dir := copyProblem(t, "pass")
	if _, err := Compile(dir, cxx); err != nil {
		t.Fatal(err)
	}
	cs, err := Cases(dir)
	if err != nil || len(cs) != 1 || !cs[0].HasOut {
		t.Fatalf("케이스 로드 실패: %v %v", cs, err)
	}
	res := RunCase(context.Background(), dir, cs[0])
	if res.Status != Pass {
		t.Errorf("상태 = %s, want PASS (ExitErr=%q Stderr=%q)", res.Status, res.ExitErr, res.Stderr)
	}
	if string(res.Actual) != "5\n" || string(res.Expected) != "5\n" {
		t.Errorf("Actual=%q Expected=%q", res.Actual, res.Expected)
	}
	if res.ExitErr != "" || res.Truncated {
		t.Errorf("정상 케이스: ExitErr=%q Truncated=%v", res.ExitErr, res.Truncated)
	}
}

func TestRunCaseFail(t *testing.T) {
	cxx := needGXX(t)
	dir := copyProblem(t, "fail")
	if _, err := Compile(dir, cxx); err != nil {
		t.Fatal(err)
	}
	cs, _ := Cases(dir)
	res := RunCase(context.Background(), dir, cs[0])
	if res.Status != Fail {
		t.Fatalf("상태 = %s, want FAIL", res.Status)
	}
	if string(res.Actual) != "6\n" || string(res.Expected) != "5\n" {
		t.Errorf("Actual=%q Expected=%q", res.Actual, res.Expected)
	}
}

func TestRunCaseTLE(t *testing.T) {
	if testing.Short() {
		t.Skip("-short: 케이스당 2초 대기 생략")
	}
	cxx := needGXX(t)
	dir := copyProblem(t, "tle")
	if _, err := Compile(dir, cxx); err != nil {
		t.Fatal(err)
	}
	cs, _ := Cases(dir)
	res := RunCase(context.Background(), dir, cs[0])
	if res.Status != TLE {
		t.Fatalf("상태 = %s, want TLE (ExitErr=%q)", res.Status, res.ExitErr)
	}
	if res.Millis < 1500 {
		t.Errorf("Millis = %d, 타임아웃까지 살아 있어야 함", res.Millis)
	}
	if len(res.Actual) != 0 {
		t.Errorf("무한 루프 프로그램은 출력이 없어야 함: %q", res.Actual)
	}
}

func TestRunCaseRE(t *testing.T) {
	cxx := needGXX(t)
	dir := copyProblem(t, "re")
	if _, err := Compile(dir, cxx); err != nil {
		t.Fatal(err)
	}
	cs, _ := Cases(dir)
	res := RunCase(context.Background(), dir, cs[0])
	if res.Status != RE {
		t.Fatalf("상태 = %s, want RE", res.Status)
	}
	if res.ExitErr == "" {
		t.Error("RE는 ExitErr 요약이 있어야 한다")
	}
}

func TestRunCaseNoExpectedOutput(t *testing.T) {
	cxx := needGXX(t)
	dir := copyProblem(t, "pass")
	if _, err := Compile(dir, cxx); err != nil {
		t.Fatal(err)
	}
	c := Case{In: filepath.Join(dir, "1.in")} // HasOut=false
	res := RunCase(context.Background(), dir, c)
	if res.Status != NoExp {
		t.Errorf("상태 = %s, want NOEXP", res.Status)
	}
	if res.Expected != nil {
		t.Errorf("Expected = %q, want nil", res.Expected)
	}
	if string(res.Actual) != "5\n" {
		t.Errorf("실행은 하므로 Actual=%q, want %q", res.Actual, "5\n")
	}
}

func TestRunCaseOutputCap(t *testing.T) {
	cxx := needGXX(t)
	dir := t.TempDir()
	// MaxOutput 바이트를 넘게 출력하는 프로그램
	src := "#include <cstdio>\nint main(){for(long i=0;i<(1<<20)+4096;i++)putchar('x');return 0;}\n"
	if err := os.WriteFile(filepath.Join(dir, "main.cpp"), []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "1.in"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "1.out"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := FastBuild(dir, cxx); err != nil {
		t.Fatal(err)
	}
	cs, _ := Cases(dir)
	if len(cs) != 1 || !cs[0].HasOut {
		t.Fatalf("케이스 로드 실패: %v", cs)
	}
	res := RunCaseWith(context.Background(), dir, "fast", cs[0])
	if !res.Truncated {
		t.Error("1MiB 초과 출력은 Truncated=true")
	}
	if len(res.Actual) != MaxOutput {
		t.Errorf("Actual 크기 = %d, want %d", len(res.Actual), MaxOutput)
	}
	if res.Status != Fail {
		t.Errorf("상태 = %s, want FAIL", res.Status)
	}
}
