// Package linediff는 기대/실제 출력의 라인 diff를 외부 의존 없이 제공한다.
package linediff

import (
	"fmt"
	"strings"
)

// Kind는 diff 연산. Del은 기대에만, Add는 실제에만 있는 줄.
type Kind int

const (
	Keep Kind = iota
	Del
	Add
)

// Line은 diff 결과 1줄.
type Line struct {
	Kind Kind
	Text string
}

// maxLines: 이 줄 수를 초과하는 입력은 전체 diff 대신 요약만 반환.
const maxLines = 2000

// Diff는 LCS 기반 라인 diff. expected/actual 둘 중 하나라도 2000줄을
// 초과하면 "첫 불일치 줄 n"(같으면 0) 한 줄로 대체한다.
func Diff(expected, actual []string) []Line {
	if len(expected) > maxLines || len(actual) > maxLines {
		return []Line{{Kind: Add, Text: fmt.Sprintf("첫 불일치 줄 %d", firstMismatch(expected, actual))}}
	}
	n, m := len(expected), len(actual)
	// lcs[i][j] = expected[i:]와 actual[j:]의 LCS 길이
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case expected[i] == actual[j]:
				lcs[i][j] = lcs[i+1][j+1] + 1
			case lcs[i+1][j] >= lcs[i][j+1]:
				lcs[i][j] = lcs[i+1][j]
			default:
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}
	var out []Line
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case expected[i] == actual[j]:
			out = append(out, Line{Kind: Keep, Text: expected[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			out = append(out, Line{Kind: Del, Text: expected[i]})
			i++
		default:
			out = append(out, Line{Kind: Add, Text: actual[j]})
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, Line{Kind: Del, Text: expected[i]})
	}
	for ; j < m; j++ {
		out = append(out, Line{Kind: Add, Text: actual[j]})
	}
	return out
}

// firstMismatch: 1-based 첫 불일치 줄 번호. 완전히 같으면 0.
func firstMismatch(expected, actual []string) int {
	n := len(expected)
	if len(actual) < n {
		n = len(actual)
	}
	for i := 0; i < n; i++ {
		if expected[i] != actual[i] {
			return i + 1
		}
	}
	if len(expected) == len(actual) {
		return 0
	}
	return n + 1
}

// WhitespaceOnly: 두 바이트열이 줄 끝 공백/CR(끝 개행 포함) 차이만 나면 true.
func WhitespaceOnly(expected, actual []byte) bool {
	a := trimLineEnds(expected)
	b := trimLineEnds(actual)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func trimLineEnds(b []byte) []string {
	lines := strings.Split(string(b), "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1] // 끝 개행이 만든 빈 줄 제거
	}
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\r")
	}
	return lines
}
