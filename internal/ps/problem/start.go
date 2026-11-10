package problem

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const startFile = ".start"

// ReadStart는 .start의 epoch 초를 읽는다. 없거나 파싱에 실패하면 (0, false).
// 공백은 trim해 허용한다.
func ReadStart(dir string) (int64, bool) {
	data, err := os.ReadFile(filepath.Join(dir, startFile))
	if err != nil {
		return 0, false
	}
	n, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// WriteStart는 epoch 초 + "\n"을 .start에 기록한다(fish `date +%s >`와 동일).
func WriteStart(dir string, now time.Time) error {
	return os.WriteFile(filepath.Join(dir, startFile), []byte(strconv.FormatInt(now.Unix(), 10)+"\n"), 0o644)
}

// ElapsedMinutes는 (now-start)/60 정수 나눗셈 결과. 음수면 0.
func ElapsedMinutes(start int64, now time.Time) int {
	m := (now.Unix() - start) / 60
	if m < 0 {
		return 0
	}
	return int(m)
}
