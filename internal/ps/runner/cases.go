package runner

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Case는 .in 파일 하나와 짝이 되는 .out의 존재 여부.
type Case struct {
	In      string // .in 파일 경로
	OutPath string // .out 경로 (없으면 "")
	HasOut  bool
}

// Cases는 dir의 *.in을 자연 정렬해 반환한다(fish 글롭과 같이 숨김 파일 제외).
// 같은 이름의 .out이 있으면 HasOut=true가 된다. .in이 없으면 빈 슬라이스.
func Cases(dir string) ([]Case, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var cs []Case
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || isHidden(name) || !strings.HasSuffix(name, ".in") {
			continue
		}
		c := Case{In: filepath.Join(dir, name)}
		out := filepath.Join(dir, strings.TrimSuffix(name, ".in")+".out")
		if st, err := os.Stat(out); err == nil && !st.IsDir() {
			c.OutPath = out
			c.HasOut = true
		}
		cs = append(cs, c)
	}
	slices.SortFunc(cs, func(a, b Case) int {
		return naturalCompare(filepath.Base(a.In), filepath.Base(b.In))
	})
	return cs, nil
}

func isHidden(name string) bool { return strings.HasPrefix(name, ".") }

// naturalCompare는 자연 정렬 비교자: 숫자 덩어리를 수치로 비교해 "2" < "10".
func naturalCompare(a, b string) int {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if isDigit(a[i]) && isDigit(b[j]) {
			si, sj := i, j
			for i < len(a) && isDigit(a[i]) {
				i++
			}
			for j < len(b) && isDigit(b[j]) {
				j++
			}
			na, nb := strings.TrimLeft(a[si:i], "0"), strings.TrimLeft(b[sj:j], "0")
			if c := cmp.Compare(len(na), len(nb)); c != 0 {
				return c
			}
			if c := strings.Compare(na, nb); c != 0 {
				return c
			}
			continue
		}
		if a[i] != b[j] {
			return cmp.Compare(int(a[i]), int(b[j]))
		}
		i++
		j++
	}
	// 앞부분이 같으면 남은 길이가 짧은 쪽이 먼저. 자연 정렬상 동등한 쌍
	// (예: "02"와 "2")은 원문 사전순으로 순서를 고정한다.
	if c := cmp.Compare(len(a)-i, len(b)-j); c != 0 {
		return c
	}
	return strings.Compare(a, b)
}

func isDigit(c byte) bool { return '0' <= c && c <= '9' }
