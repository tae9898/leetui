package problem

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

const typePrefix = "// 유형:"

// ParseType는 main.cpp 앞 10줄에서 "// 유형:" 접두 뒤의 텍스트를 trim해 돌려준다.
// 마커가 없거나 파일이 없으면 빈 문자열.
func ParseType(dir string) string {
	f, err := os.Open(filepath.Join(dir, "main.cpp"))
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for n := 0; sc.Scan() && n < 10; n++ {
		rest, ok := strings.CutPrefix(strings.TrimSpace(sc.Text()), typePrefix)
		if ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}
