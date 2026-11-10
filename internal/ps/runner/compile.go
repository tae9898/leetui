// Package runner는 컴파일, 케이스 실행, 타임아웃, 판정을 담는다
// (TUI-PLAN.md 2.3 runner 표). 컴파일 플래그는 CLI-ENV.md 4절과 문자 단위로 동일하다.
package runner

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
	"time"
)

// CLI-ENV.md 4절의 플래그. 코테 함정 탐지용이므로 변경 금지 — 테스트로 고정.
const (
	SanitizeFlags = "-std=c++17 -O2 -g -Wall -Wextra -Wshadow -fsanitize=address,undefined -D_GLIBCXX_DEBUG"
	FastFlags     = "-std=c++17 -O2"
)

// Compile은 dir에서 판정용(sanitizer) 빌드를 실행해 sol을 만든다.
// stderr 전체를 반환하고 종료 코드가 0이 아니면 err를 돌려준다.
func Compile(dir, cxx string) (stderr string, err error) {
	return build(dir, cxx, SanitizeFlags, "sol")
}

// FastBuild는 시간 초과 확인용 fast 바이너리를 만들고 빌드 경과를 함께 반환한다.
func FastBuild(dir, cxx string) (stderr string, elapsed time.Duration, err error) {
	start := time.Now()
	stderr, err = build(dir, cxx, FastFlags, "fast")
	return stderr, time.Since(start), err
}

func build(dir, cxx, flags, out string) (string, error) {
	cmd := exec.Command(cxx, append(strings.Fields(flags), "main.cpp", "-o", out)...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stdout = io.Discard
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stderr.String(), err
}
