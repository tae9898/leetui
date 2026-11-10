package problem

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Validate는 psnew 입력 검증: 세 값 모두 비어 있지 않아야 하고
// '/', 공백, 선행 '.'을 쓸 수 없다.
func Validate(site, id, name string) error {
	for _, v := range []struct{ field, val string }{
		{"site", site}, {"id", id}, {"name", name},
	} {
		switch {
		case v.val == "":
			return fmt.Errorf("%s가 비어 있습니다", v.field)
		case strings.Contains(v.val, "/"):
			return fmt.Errorf("%s에 '/'를 쓸 수 없습니다: %q", v.field, v.val)
		case strings.Contains(v.val, " "):
			return fmt.Errorf("%s에 공백을 쓸 수 없습니다: %q", v.field, v.val)
		case strings.HasPrefix(v.val, "."):
			return fmt.Errorf("%s는 '.'로 시작할 수 없습니다: %q", v.field, v.val)
		}
	}
	return nil
}

// New는 root/site/id-name 디렉터리를 만들고 fish psnew와 동일한 산출물을 남긴다.
// main.cpp가 이미 있으면 덮어쓰지 않고(existedMain=true) 템플릿을 복사하지 않는다.
// 1.in/1.out은 기존 파일을 자르지 않고, .start에 now의 epoch 초를 기록한다.
func New(root, site, id, name, template string, now time.Time) (dir string, existedMain bool, err error) {
	dir = filepath.Join(root, site, id+"-"+name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", false, err
	}
	mainPath := filepath.Join(dir, "main.cpp")
	_, statErr := os.Stat(mainPath)
	switch {
	case statErr == nil:
		existedMain = true
	case errors.Is(statErr, fs.ErrNotExist):
		if err := copyFile(template, mainPath, 0o644); err != nil {
			return "", false, err
		}
	default:
		return "", false, statErr
	}
	for _, f := range []string{"1.in", "1.out"} {
		// O_CREATE만 쓰므로 이미 있는 파일 내용은 유지된다(fish touch와 동일).
		fh, err := os.OpenFile(filepath.Join(dir, f), os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return "", false, err
		}
		if err := fh.Close(); err != nil {
			return "", false, err
		}
	}
	if err := WriteStart(dir, now); err != nil {
		return "", false, err
	}
	return dir, existedMain, nil
}

// copyFile은 src 내용을 그대로 dst에 쓴다(새 파일 권한 perm).
func copyFile(src, dst string, perm fs.FileMode) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, perm)
}
