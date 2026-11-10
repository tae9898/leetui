// Package problem은 문제 디렉터리 스캔, new/retry, .start, 헤더 파싱 규칙을
// 담는다(TUI-PLAN.md 2.3 problem 표). fish(psnew/psretry)와 산출물이 동일하다.
package problem

import (
	"cmp"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Problem은 스캔으로 발견한 문제 디렉터리 하나.
type Problem struct {
	Site string // PS_ROOT 바로 아래 디렉터리명
	Name string // 문제 디렉터리명
	Dir  string // 전체 경로
}

// Scan은 root 바로 아래 숨김 아닌 디렉터리(사이트) 전부와 그 아래 숨김 아닌
// 디렉터리(문제)를 (site, name) 자연 정렬로 반환한다. 사이트는 특정 이름에 한정하지 않는다.
func Scan(root string) ([]Problem, error) {
	sites, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var ps []Problem
	for _, site := range sites {
		if !site.IsDir() || isHidden(site.Name()) {
			continue
		}
		probs, err := os.ReadDir(filepath.Join(root, site.Name()))
		if err != nil {
			continue // 읽기 실패한 사이트는 건너뛴다
		}
		for _, p := range probs {
			if !p.IsDir() || isHidden(p.Name()) {
				continue
			}
			ps = append(ps, Problem{
				Site: site.Name(),
				Name: p.Name(),
				Dir:  filepath.Join(root, site.Name(), p.Name()),
			})
		}
	}
	slices.SortFunc(ps, func(a, b Problem) int {
		if c := naturalCompare(a.Site, b.Site); c != 0 {
			return c
		}
		return naturalCompare(a.Name, b.Name)
	})
	return ps, nil
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
