package problem

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// Retry는 fish psretry와 동일하게 기존 풀이를 main.MMDD.cpp로 옮기고(이미
// 있으면 덮어쓴다) 템플릿으로 리셋한 뒤 .start를 갱신한다. main.cpp가 없으면
// 백업 없이 리셋만 하고 backup로 ""을 돌려준다.
func Retry(dir, template string, now time.Time) (backup string, err error) {
	mainPath := filepath.Join(dir, "main.cpp")
	_, statErr := os.Stat(mainPath)
	if statErr == nil {
		backup = "main." + now.Format("0102") + ".cpp" // MMDD, fish date +%m%d와 동일
		if err := os.Rename(mainPath, filepath.Join(dir, backup)); err != nil {
			return "", err
		}
	} else if !errors.Is(statErr, fs.ErrNotExist) {
		return "", statErr
	}
	if err := copyFile(template, mainPath, 0o644); err != nil {
		return "", err
	}
	if err := WriteStart(dir, now); err != nil {
		return "", err
	}
	return backup, nil
}
