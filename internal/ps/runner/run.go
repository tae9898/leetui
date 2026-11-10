package runner

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const (
	CaseTimeout = 2 * time.Second // fish `timeout 2`와 동일
	MaxOutput   = 1 << 20         // stdout 캡처 상한 1MiB
)

// Status는 케이스 판정 결과.
type Status string

const (
	Pass  Status = "PASS"  // 기대출력과 바이트 단위 동일
	Fail  Status = "FAIL"  // 출력 불일치
	TLE   Status = "TLE"   // 컨텍스트 타임아웃
	RE    Status = "RE"    // 0이 아닌 종료/시그널 (sanitizer 포함)
	NoExp Status = "NOEXP" // .out 없음 — 실행은 하되 판정 제외
)

// CaseResult는 케이스 하나의 실행 결과.
type CaseResult struct {
	Case      Case
	Status    Status
	Expected  []byte // .out 전체 (HasOut=false면 nil)
	Actual    []byte // stdout (MaxOutput 상한, 잘리면 Truncated=true)
	Stderr    string
	Millis    int64
	Truncated bool
	ExitErr   string // 비정상 종료/실행 실패 요약 (RE 힌트)
}

// RunCase는 dir에서 ./sol을 c.In을 표준입력으로 실행해 판정한다.
func RunCase(ctx context.Context, dir string, c Case) CaseResult {
	return RunCaseWith(ctx, dir, "sol", c)
}

// RunCaseWith는 실행할 바이너리 이름(bin, 예: "sol"/"fast")만 다른 RunCase.
func RunCaseWith(ctx context.Context, dir, bin string, c Case) CaseResult {
	ctx, cancel := context.WithTimeout(ctx, CaseTimeout)
	defer cancel()

	res := CaseResult{Case: c}
	if c.HasOut {
		exp, err := os.ReadFile(c.OutPath)
		if err != nil {
			res.Status = Fail
			res.ExitErr = "기대출력 읽기 실패: " + err.Error()
			return res
		}
		res.Expected = exp
	}

	in, err := os.Open(c.In)
	if err != nil {
		res.Status = RE
		res.ExitErr = "입력 파일 열기 실패: " + err.Error()
		return res
	}
	defer in.Close()

	cmd := exec.CommandContext(ctx, "./"+bin)
	cmd.Dir = dir
	cmd.Stdin = in
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	// 타임아웃/취소 시 sol이 남긴 자식까지 죽인다(프로세스 그룹 kill).
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 100 * time.Millisecond
	stdout := &limitWriter{max: MaxOutput}
	var stderr bytes.Buffer
	cmd.Stdout = stdout
	cmd.Stderr = &stderr

	start := time.Now()
	runErr := cmd.Run()
	res.Millis = time.Since(start).Milliseconds()
	res.Actual = stdout.buf
	res.Truncated = stdout.truncated
	res.Stderr = stderr.String()

	switch {
	case ctx.Err() != nil: // 컨텍스트 타임아웃 또는 취소
		res.Status = TLE
		res.ExitErr = ctx.Err().Error()
	case runErr != nil: // 0이 아닌 종료/시그널 → RE (이때 Actual은 부분 출력)
		res.Status = RE
		res.ExitErr = runErr.Error()
	case !c.HasOut:
		res.Status = NoExp
	case bytes.Equal(res.Actual, res.Expected):
		res.Status = Pass
	default:
		res.Status = Fail
	}
	return res
}

// limitWriter는 max 바이트까지만 저장하고 이후는 버리되 쓰기를 계속
// 소비해 자식 프로세스가 출력 파이프에서 막히지 않게 한다.
type limitWriter struct {
	buf       []byte
	max       int
	truncated bool
}

func (w *limitWriter) Write(p []byte) (int, error) {
	if n := len(p); n > 0 {
		if room := w.max - len(w.buf); room > 0 {
			if room > n {
				room = n
			}
			w.buf = append(w.buf, p[:room]...)
			if room < n {
				w.truncated = true
			}
		} else {
			w.truncated = true
		}
	}
	return len(p), nil
}
