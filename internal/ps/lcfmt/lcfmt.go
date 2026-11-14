// Package lcfmt은 LeetCode metaData와 입출력 리터럴을 leetui 표준
// stdin/stdout 포맷으로 서로 변환한다 (lc-cases.md 플랜 D1).
//
// 표준 포맷: metaData.params 순서대로 직렬화. 토큰은 공백, 구조는 줄로 구분.
// 지원 범위 밖(design, double 반환 등)은 Supported이 거짓이 되고 호출부가 폴백한다.
package lcfmt

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// SpecMarker는 지문 md에 삽입하는 형식 섹션의 제목 — 중복 삽입 방지용 식별자.
const SpecMarker = "## 입력 형식 (leetui)"

// Meta는 LC metaData에서 케이스 변환에 필요한 부분.
type Meta struct {
	Name      string // 함수명 (twoSum)
	ClassName string // design 문제만 비어 있지 않음
	Params    []Param
	Return    Return
}

type Param struct {
	Name string
	Type string // "integer[]", "character[][]", "ListNode" …
}

type Return struct {
	Type        string // "void"면 in-place 수정
	OutParam    int    // output.paramindex — 수정된 파라미터를 출력
	HasOutParam bool
}

func ParseMeta(s string) (Meta, error) {
	var raw struct {
		Name      string `json:"name"`
		ClassName string `json:"classname"`
		Params    []Param
		Return    *struct {
			Type string `json:"type"`
		} `json:"return"`
		Output *struct {
			ParamIndex int `json:"paramindex"`
		} `json:"output"`
	}
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return Meta{}, fmt.Errorf("metaData 파싱 실패: %w", err)
	}
	m := Meta{Name: raw.Name, ClassName: raw.ClassName, Params: raw.Params}
	if raw.Return != nil {
		m.Return.Type = raw.Return.Type
	}
	if raw.Output != nil {
		m.Return.OutParam = raw.Output.ParamIndex
		m.Return.HasOutParam = true
	}
	if m.Name == "" && m.ClassName == "" || len(m.Params) == 0 {
		return Meta{}, fmt.Errorf("metaData에 함수 시그니처 없음")
	}
	return m, nil
}

// paramTypes는 stdin으로 읽을 수 있는 파라미터 타입.
var paramTypes = map[string]bool{
	"integer": true, "long": true, "double": true, "boolean": true,
	"string": true, "character": true,
	"integer[]": true, "long[]": true, "double[]": true,
	"string[]": true, "character[]": true,
	"integer[][]": true, "character[][]": true,
	"ListNode": true, "TreeNode": true,
}

// returnTypes는 정확 비교로 판정 가능한 반환 타입. double/double[]은
// LC가 1e-5 오차를 허용하지만 runner는 바이트 비교라 제외.
var returnTypes = map[string]bool{
	"void": true, "integer": true, "long": true, "boolean": true,
	"string": true, "character": true,
	"integer[]": true, "long[]": true, "string[]": true,
	"integer[][]": true, "character[][]": true,
	"ListNode": true, "TreeNode": true,
}

// multiAnswerTypes는 답 전체가 모음이라 순서가 자유로운 반환 타입.
// 이때만 "any order" 지문을 폴백 사유로 삼는다 — two-sum처럼 고정 형태
// 답(integer[2] 등)은 표준 풀이가 LC 정답과 같은 순서를 낸다.
var multiAnswerTypes = map[string]bool{
	"integer[][]": true, "string[]": true, "character[][]": true,
}

// Supported는 이 문제를 표준 포맷으로 다룰 수 있는지 판정한다.
// 거짓이면 reason에 폴백 사유가 들어간다.
func Supported(m Meta, content string) (bool, string) {
	if m.ClassName != "" {
		return false, "design 문제(클래스 인터페이스)"
	}
	if !returnTypes[m.Return.Type] {
		return false, "미지원 반환 타입: " + m.Return.Type
	}
	if m.Return.Type == "void" && !m.Return.HasOutParam {
		return false, "출력 파라미터 없는 void"
	}
	if m.Return.HasOutParam && m.Return.OutParam >= len(m.Params) {
		return false, "output.paramindex 범위 초과"
	}
	for _, p := range m.Params {
		if !paramTypes[p.Type] {
			return false, "미지원 파라미터 타입: " + p.Type
		}
	}
	if multiAnswerTypes[outputType(m)] && strings.Contains(strings.ToLower(content), "any order") {
		return false, "정답이 여러 개(any order)"
	}
	return true, ""
}

// outputType은 판정 대상 타입 — void+paramindex면 수정되는 파라미터 타입.
func outputType(m Meta) string {
	if m.Return.Type == "void" {
		return m.Params[m.Return.OutParam].Type
	}
	return m.Return.Type
}

// ToStdin은 LC 케이스(exampleTestcaseList 원소 — 파라미터 리터럴을 \n로 이음)를
// 표준 stdin 텍스트로 바꾼다.
func ToStdin(m Meta, lcCase string) (string, error) {
	lines := strings.Split(lcCase, "\n")
	if len(lines) != len(m.Params) {
		return "", fmt.Errorf("파라미터 %d개인데 케이스 라인이 %d줄", len(m.Params), len(lines))
	}
	var out []string
	for i, p := range m.Params {
		ls, err := paramIn(p, lines[i])
		if err != nil {
			return "", fmt.Errorf("파라미터 %s: %w", p.Name, err)
		}
		out = append(out, ls...)
	}
	return strings.Join(out, "\n") + "\n", nil
}

// ToStdout은 LC 정답 리터럴(expected_code_answer 원소, 지문 Output 텍스트)을
// 표준 stdout 텍스트로 바꾼다.
func ToStdout(m Meta, lcOut string) (string, error) {
	typ := outputType(m)
	ls, err := outLines(typ, lcOut)
	if err != nil {
		return "", fmt.Errorf("출력 변환: %w", err)
	}
	return strings.Join(ls, "\n") + "\n", nil
}

// FromStdin은 표준 stdin 텍스트를 LC 케이스 리터럴(파라미터별 한 줄)로 되돌린다.
// 오라클(O 키)이 내 입력을 LC로 보낼 때 쓴다.
func FromStdin(m Meta, stdin string) ([]string, error) {
	r := &lineReader{lines: strings.Split(strings.TrimSuffix(stdin, "\n"), "\n")}
	var lits []string
	for i, p := range m.Params {
		lit, err := readParam(r, p)
		if err != nil {
			return nil, fmt.Errorf("파라미터 %d(%s): %w", i+1, p.Name, err)
		}
		lits = append(lits, lit)
	}
	return lits, nil
}

// FormatSpec은 지문 md 상단에 넣을 입력/출력 형식 설명 (HR Input Format 칸 대응).
func FormatSpec(m Meta) string {
	var b strings.Builder
	fmt.Fprintf(&b, "## 입력 형식 (leetui)\n\n함수 ")
	if m.Name != "" {
		b.WriteString(m.Name + "(")
		for i, p := range m.Params {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(p.Name + " " + p.Type)
		}
		b.WriteString(")")
	}
	if m.Return.Type == "void" {
		fmt.Fprintf(&b, " — %s을(를) 제자리에서 수정", m.Params[m.Return.OutParam].Name)
	} else {
		fmt.Fprintf(&b, " → %s", m.Return.Type)
	}
	b.WriteString("\n\n입력:\n\n")
	for _, p := range m.Params {
		fmt.Fprintf(&b, "- %s: %s\n", p.Name, typeInSpec(p.Type))
	}
	b.WriteString("\n출력:\n\n")
	if m.Return.Type == "void" {
		fmt.Fprintf(&b, "- 수정된 %s: %s\n", m.Params[m.Return.OutParam].Name, typeOutSpec(m.Params[m.Return.OutParam].Type))
	} else {
		fmt.Fprintf(&b, "- 반환값(%s): %s\n", m.Return.Type, typeOutSpec(m.Return.Type))
	}
	return b.String()
}

func typeInSpec(t string) string {
	switch t {
	case "integer", "long":
		return "정수 한 줄"
	case "double":
		return "실수 한 줄"
	case "boolean":
		return "true 또는 false 한 줄"
	case "string", "character", "character[]":
		return "한 줄 원문 (character[]는 문자를 이어붙임)"
	case "integer[]", "long[]", "double[]", "ListNode":
		return "두 줄 — 크기 n, 값(공백 구분). n=0이면 값 줄은 빈 줄"
	case "string[]":
		return "크기 n 한 줄 + 문자열 n줄"
	case "integer[][]":
		return "행 수 r 한 줄 + 행마다 `k v1 … vk` (들쭉날쭉해도 됨)"
	case "character[][]":
		return "`r c` 한 줄 + 행마다 길이 c 문자열"
	case "TreeNode":
		return "노드 수 n 한 줄 + 레벨순서 토큰 한 줄 (빈 자식은 null)"
	}
	return t
}

func typeOutSpec(t string) string {
	switch t {
	case "integer[]", "long[]", "double[]", "ListNode", "TreeNode":
		if t == "TreeNode" {
			return "레벨순서 토큰 한 줄 (뒤쪽 null은 생략)"
		}
		return "값 한 줄 (공백 구분, 빈 배열이면 빈 줄)"
	case "string[]":
		return "문자열 줄마다 하나"
	case "integer[][]":
		return "행마다 `v1 … vk` 한 줄"
	case "character[][]":
		return "행마다 문자열 한 줄"
	case "boolean":
		return "true 또는 false"
	}
	return "한 줄"
}

// lineReader는 FromStdin용 순차 라인 판독기.
type lineReader struct {
	lines []string
	i     int
}

func (r *lineReader) next() (string, error) {
	if r.i >= len(r.lines) {
		return "", io.EOF
	}
	s := strings.TrimSuffix(r.lines[r.i], "\r")
	r.i++
	return s, nil
}

func (r *lineReader) count(what string) (int, error) {
	line, err := r.next()
	if err != nil {
		return 0, fmt.Errorf("%s 줄 없음", what)
	}
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%s 줄이 음수가 아닌 정수가 아님: %q", what, line)
	}
	return n, nil
}
