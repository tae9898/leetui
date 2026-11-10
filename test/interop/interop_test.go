// fish ps* 함수와 Go 도메인의 상호 운용 테스트 (TUI-PLAN.md 5.2절).
// fish가 없으면 전체 스킵. 임시 PS_ROOT에서 교차 실행해 바이트 단위 호환을 검증한다.
package interop

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"pstui/internal/ps/problem"
	"pstui/internal/ps/pslog"
)

const template = `// 문제:
// 유형:
// 목표 복잡도:
#include <bits/stdc++.h>
using namespace std;

int main() {
    ios::sync_with_stdio(false);
    cin.tie(nullptr);

    return 0;
}
`

func fishOK(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish 없음")
	}
	if _, err := os.Stat(os.ExpandEnv("$HOME/.config/fish/conf.d/ps.fish")); err != nil {
		t.Skip("ps.fish 없음")
	}
}

// runFish는 nvim을 무력화한 fish에서 ps* 함수를 실행하고 stdout을 반환한다.
func runFish(t *testing.T, psRoot, dir, script string) string {
	t.Helper()
	cmd := exec.Command("fish", "-c", "function nvim; end; "+script)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "PS_ROOT="+psRoot)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("fish 실행 실패 (%s): %v", script, err)
	}
	return stdout.String()
}

func makeRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "template.cpp"), []byte(template), 0644); err != nil {
		t.Fatal(err)
	}
	return root
}

func today() string { return time.Now().Format("2006-01-02") }

// 1. Go New → fish psrun(빈 템플릿+빈 1.in/1.out = PASS) → fish psdone fail
//    → Go가 그 행을 3일 뒤 due로 읽는다.
func TestGoNewFishRunFishDoneGoDue(t *testing.T) {
	fishOK(t)
	root := makeRoot(t)
	now := time.Now()
	dir, _, err := problem.New(root, "lc", "0001", "t-a", filepath.Join(root, "template.cpp"), now)
	if err != nil {
		t.Fatal(err)
	}
	if out := runFish(t, root, dir, "psrun"); !strings.Contains(out, "PASS 1.in") {
		t.Fatalf("fish psrun이 빈 케이스를 PASS하지 않음: %q", out)
	}
	runFish(t, root, dir, `psdone fail "interop 메모"`)

	entries, skipped, err := pslog.Parse(filepath.Join(root, "log.tsv"))
	if err != nil || skipped != 0 {
		t.Fatalf("Parse: %v skipped=%d", err, skipped)
	}
	if len(entries) != 1 {
		t.Fatalf("기록 1행이어야 함: %+v", entries)
	}
	e := entries[0]
	if e.Site != "lc" || e.Problem != "0001-t-a" || e.Result != pslog.Result("fail") {
		t.Fatalf("fish가 쓴 행을 Go가 다르게 읽음: %+v", e)
	}
	want := pslog.DueDate("fail", now)
	if e.Due != want {
		t.Fatalf("due 불일치: fish=%s go=%s", e.Due, want)
	}
	if due := pslog.Due(entries, want); len(due) != 1 || due[0].Key != "lc/0001-t-a" {
		t.Fatalf("3일 뒤 Due 조회 실패: %+v", due)
	}
}

// 2. fish psnew → Go Append(ok) → fish psdue 출력이 Go Due 포맷과 바이트 동일.
func TestFishPsdueMatchesGoDue(t *testing.T) {
	fishOK(t)
	root := makeRoot(t)
	runFish(t, root, root, "psnew lc 0002 t-b")
	logPath := filepath.Join(root, "log.tsv")
	appends := []pslog.Entry{
		{Date: "2026-01-01", Site: "lc", Problem: "0002-t-b", Result: "ok", Minutes: 10, Due: "-"},
		{Date: today(), Site: "lc", Problem: "0002-t-b", Result: "fail", Minutes: 20, Due: today(), Note: "backdated"},
	}
	for _, e := range appends {
		if err := pslog.Append(logPath, e); err != nil {
			t.Fatal(err)
		}
	}
	entries, _, err := pslog.Parse(logPath)
	if err != nil {
		t.Fatal(err)
	}
	var goOut bytes.Buffer
	for _, it := range pslog.Due(entries, today()) {
		goOut.WriteString(pslog.FormatDue(it))
		goOut.WriteByte('\n')
	}
	fishOut := runFish(t, root, root, "psdue")
	if fishOut != goOut.String() {
		t.Fatalf("psdue 출력 불일치:\nfish=%q\ngo  =%q", fishOut, goOut.String())
	}
	if !strings.Contains(fishOut, today()+" lc/0002-t-b fail") {
		t.Fatalf("마지막 행 우선 반영 안 됨: %q", fishOut)
	}
}

// 3. fish psdone과 Go Append가 같은 입력으로 만든 행이 바이트 단위 동일.
func TestAppendRowByteEqual(t *testing.T) {
	fishOK(t)
	root := makeRoot(t)
	now := time.Now()
	dir, _, err := problem.New(root, "lc", "0003", "t-c", filepath.Join(root, "template.cpp"), now)
	if err != nil {
		t.Fatal(err)
	}
	startEpoch := strconv.FormatInt(now.Add(-2*time.Minute).Unix(), 10)
	if err := os.WriteFile(filepath.Join(dir, ".start"), []byte(startEpoch+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runFish(t, root, dir, `psdone fail "메모 테스트"`)

	logPath := filepath.Join(root, "log.tsv")
	if err := pslog.Append(logPath, pslog.Entry{
		Date: today(), Site: "lc", Problem: "0003-t-c",
		Result: "fail", Minutes: 2, Due: pslog.DueDate("fail", now), Note: "메모 테스트",
	}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	rows := bytes.Split(bytes.TrimSuffix(data, []byte("\n")), []byte("\n"))
	if len(rows) != 2 {
		t.Fatalf("2행이어야 함: %q", data)
	}
	if !bytes.Equal(rows[0], rows[1]) {
		t.Fatalf("행 불일치:\nfish=%q\ngo  =%q", rows[0], rows[1])
	}
}

// 4. Go Retry 후 fish psrun 정상 동작 + 백업 파일명 main.MMDD.cpp.
func TestGoRetryFishRun(t *testing.T) {
	fishOK(t)
	root := makeRoot(t)
	now := time.Now()
	dir, _, err := problem.New(root, "lc", "0004", "t-d", filepath.Join(root, "template.cpp"), now)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.cpp"), []byte("// 풀이\n"), 0644); err != nil {
		t.Fatal(err)
	}
	backup, err := problem.Retry(dir, filepath.Join(root, "template.cpp"), now)
	if err != nil {
		t.Fatal(err)
	}
	wantBackup := "main." + now.Format("0102") + ".cpp"
	if backup != wantBackup {
		t.Fatalf("백업명 불일치: %s != %s", backup, wantBackup)
	}
	if _, err := os.Stat(filepath.Join(dir, wantBackup)); err != nil {
		t.Fatalf("백업 파일 없음: %v", err)
	}
	if out := runFish(t, root, dir, "psrun"); !strings.Contains(out, "PASS 1.in") {
		t.Fatalf("Go Retry 후 fish psrun 실패: %q", out)
	}
}

// 5. fish가 쓴 .start를 Go가 같은 경과 분으로 계산.
func TestFishStartGoElapsed(t *testing.T) {
	fishOK(t)
	root := makeRoot(t)
	runFish(t, root, root, "psnew lc 0005 t-e")
	dir := filepath.Join(root, "lc", "0005-t-e")
	start, ok := problem.ReadStart(dir)
	if !ok {
		t.Fatal("fish가 쓴 .start를 Go가 못 읽음")
	}
	now := time.Now()
	fishMins := runFish(t, root, dir,
		"math -s0 \"( "+strconv.FormatInt(now.Unix(), 10)+" - "+strconv.FormatInt(start, 10)+" ) / 60\"")
	if got := problem.ElapsedMinutes(start, now); strconv.FormatInt(int64(got), 10) != strings.TrimSpace(fishMins) {
		t.Fatalf("경과 분 불일치: go=%d fish=%s", got, fishMins)
	}
}
