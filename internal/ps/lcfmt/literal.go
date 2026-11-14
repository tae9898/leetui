package lcfmt

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// LC 입출력 리터럴은 JSON과 거의 같다. 차이는 json.Number 정밀도 유지와
// (드문) 작은따옴표 문자 표기뿐이라 예외만 처리한다.

var reSingleQuote = regexp.MustCompile(`'((?:[^'\\]|\\.)*)'`)

// parseLiteral은 LC 리터럴 한 개를 any(json.Number/string/bool/[]any/nil)로 파싱한다.
func parseLiteral(s string) (any, error) {
	s = strings.TrimSpace(s)
	if v, err := literalJSON(s); err == nil {
		return v, nil
	}
	if strings.Contains(s, "'") {
		if v, err := literalJSON(reSingleQuote.ReplaceAllString(s, `"$1"`)); err == nil {
			return v, nil
		}
	}
	return nil, fmt.Errorf("리터럴이 JSON 형식이 아님: %.40s", s)
}

func literalJSON(s string) (any, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("리터럴 뒤에 내용이 더 있음")
	}
	return v, nil
}

// ---- LC 리터럴 → 표준 stdin 라인 ----

// paramIn은 파라미터 리터럴 하나를 stdin 라인들로 바꾼다.
func paramIn(p Param, lit string) ([]string, error) {
	switch p.Type {
	case "integer", "long", "double":
		return []string{scalarToken(lit)}, nil
	case "boolean":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		b, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("boolean 리터럴이 아님: %.40s", lit)
		}
		return []string{strconv.FormatBool(b)}, nil
	case "string", "character":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("문자열 리터럴이 아님: %.40s", lit)
		}
		if p.Type == "character" && len([]rune(s)) != 1 {
			return nil, fmt.Errorf("character가 한 글자가 아님: %q", s)
		}
		return []string{s}, nil
	case "character[]":
		// LC는 ["h","e","l"] 표기. 붙여서 한 줄로.
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		s, ok := v.(string)
		if !ok {
			var b strings.Builder
			elems, err := charElems(v, lit)
			if err != nil {
				return nil, err
			}
			for _, e := range elems {
				b.WriteString(e)
			}
			s = b.String()
		}
		return []string{s}, nil
	case "integer[]", "long[]", "double[]", "ListNode":
		toks, err := numElems(lit)
		if err != nil {
			return nil, err
		}
		return []string{strconv.Itoa(len(toks)), strings.Join(toks, " ")}, nil
	case "string[]":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		arr, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("배열 리터럴이 아님: %.40s", lit)
		}
		out := []string{strconv.Itoa(len(arr))}
		for _, e := range arr {
			s, ok := e.(string)
			if !ok {
				return nil, fmt.Errorf("문자열 아닌 원소: %v", e)
			}
			out = append(out, s)
		}
		return out, nil
	case "integer[][]":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		rows, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("2차원 리터럴이 아님: %.40s", lit)
		}
		out := []string{strconv.Itoa(len(rows))}
		for _, row := range rows {
			toks, err := rowNums(row)
			if err != nil {
				return nil, err
			}
			out = append(out, strconv.Itoa(len(toks))+" "+strings.Join(toks, " "))
		}
		return out, nil
	case "character[][]":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		rows, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("2차원 리터럴이 아님: %.40s", lit)
		}
		var rowStrs []string
		for _, row := range rows {
			switch rv := row.(type) {
			case string:
				rowStrs = append(rowStrs, rv)
			case []any:
				elems, err := charElems(rv, "")
				if err != nil {
					return nil, err
				}
				rowStrs = append(rowStrs, strings.Join(elems, ""))
			default:
				return nil, fmt.Errorf("grid 행이 문자가 아님: %v", row)
			}
		}
		if len(rowStrs) == 0 {
			return []string{"0 0"}, nil
		}
		out := []string{fmt.Sprintf("%d %d", len(rowStrs), len([]rune(rowStrs[0])))}
		return append(out, rowStrs...), nil
	case "TreeNode":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		arr, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("배열 리터럴이 아님: %.40s", lit)
		}
		var toks []string
		for _, e := range arr {
			toks = append(toks, treeNodeToken(e))
		}
		return []string{strconv.Itoa(len(toks)), strings.Join(toks, " ")}, nil
	}
	return nil, fmt.Errorf("입력 변환 미지원 타입: %s", p.Type)
}

func scalarToken(lit string) string {
	return strings.TrimSpace(lit)
}

// numElems는 숫자 배열 리터럴을 원본 토큰 문자열로 (long 정밀도 보존).
func numElems(lit string) ([]string, error) {
	v, err := parseLiteral(lit)
	if err != nil {
		return nil, err
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("배열 리터럴이 아님: %.40s", lit)
	}
	var toks []string
	for _, e := range arr {
		n, ok := e.(json.Number)
		if !ok {
			return nil, fmt.Errorf("숫자 아닌 원소: %v", e)
		}
		toks = append(toks, n.String())
	}
	return toks, nil
}

func rowNums(row any) ([]string, error) {
	arr, ok := row.([]any)
	if !ok {
		return nil, fmt.Errorf("행이 배열이 아님: %v", row)
	}
	var toks []string
	for _, e := range arr {
		n, ok := e.(json.Number)
		if !ok {
			return nil, fmt.Errorf("숫자 아닌 원소: %v", e)
		}
		toks = append(toks, n.String())
	}
	return toks, nil
}

// charElems는 한 글자 문자 원소들을 문자열로.
func charElems(v any, lit string) ([]string, error) {
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("배열 리터럴이 아님: %.40s", lit)
	}
	out := make([]string, len(arr))
	for i, e := range arr {
		s, ok := e.(string)
		if !ok || len([]rune(s)) != 1 {
			return nil, fmt.Errorf("한 글자가 아닌 원소: %v", e)
		}
		out[i] = s
	}
	return out, nil
}

func treeNodeToken(e any) string {
	if e == nil {
		return "null"
	}
	if n, ok := e.(json.Number); ok {
		return n.String()
	}
	return fmt.Sprintf("%v", e)
}

// ---- LC 정답 리터럴 → 표준 stdout 라인 ----

// outLines은 정답 리터럴을 stdout 라인들로 바꾼다. 개수 안내 줄 없이 값만.
func outLines(typ, lit string) ([]string, error) {
	if strings.TrimSpace(lit) == "" {
		return []string{""}, nil // 빈 정답 — 빈 줄 하나
	}
	switch typ {
	case "integer", "long", "double":
		return []string{scalarToken(lit)}, nil
	case "boolean":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		b, ok := v.(bool)
		if !ok {
			return nil, fmt.Errorf("boolean 리터럴이 아님: %.40s", lit)
		}
		return []string{strconv.FormatBool(b)}, nil
	case "string", "character":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("문자열 리터럴이 아님: %.40s", lit)
		}
		return []string{s}, nil
	case "character[]":
		return paramIn(Param{Type: "character[]"}, lit)
	case "integer[]", "long[]", "double[]", "ListNode":
		toks, err := numElems(lit)
		if err != nil {
			return nil, err
		}
		return []string{strings.Join(toks, " ")}, nil
	case "string[]":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		arr, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("배열 리터럴이 아님: %.40s", lit)
		}
		out := make([]string, len(arr))
		for i, e := range arr {
			s, ok := e.(string)
			if !ok {
				return nil, fmt.Errorf("문자열 아닌 원소: %v", e)
			}
			out[i] = s
		}
		return out, nil
	case "integer[][]":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		rows, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("2차원 리터럴이 아님: %.40s", lit)
		}
		out := make([]string, len(rows))
		for i, row := range rows {
			toks, err := rowNums(row)
			if err != nil {
				return nil, err
			}
			out[i] = strings.Join(toks, " ")
		}
		return out, nil
	case "character[][]":
		ls, err := paramIn(Param{Type: "character[][]"}, lit)
		if err != nil {
			return nil, err
		}
		return ls[1:], nil // 첫 줄 "r c"는 stdout에 없음
	case "TreeNode":
		v, err := parseLiteral(lit)
		if err != nil {
			return nil, err
		}
		arr, ok := v.([]any)
		if !ok {
			return nil, fmt.Errorf("배열 리터럴이 아님: %.40s", lit)
		}
		// 뒤쪽 null 제거
		for len(arr) > 0 && arr[len(arr)-1] == nil {
			arr = arr[:len(arr)-1]
		}
		var toks []string
		for _, e := range arr {
			toks = append(toks, treeNodeToken(e))
		}
		return []string{strings.Join(toks, " ")}, nil
	}
	return nil, fmt.Errorf("출력 변환 미지원 타입: %s", typ)
}

// ---- 표준 stdin → LC 리터럴 ----

// readParam은 파라미터 하나분의 stdin 라인을 읽어 LC 리터럴로 되돌린다.
func readParam(r *lineReader, p Param) (string, error) {
	switch p.Type {
	case "integer", "long", "double":
		tok, err := numLine(r, p.Type)
		if err != nil {
			return "", err
		}
		return tok, nil
	case "boolean":
		line, err := r.next()
		if err != nil {
			return "", fmt.Errorf("값 줄 없음")
		}
		t := strings.TrimSpace(line)
		if t != "true" && t != "false" {
			return "", fmt.Errorf("true/false가 아님: %q", line)
		}
		return t, nil
	case "string", "character", "character[]":
		line, err := r.next()
		if err != nil {
			return "", fmt.Errorf("값 줄 없음")
		}
		if p.Type == "string" {
			return quoteJSON(line), nil
		}
		if p.Type == "character" {
			if len([]rune(line)) != 1 {
				return "", fmt.Errorf("한 글자가 아님: %q", line)
			}
			return quoteJSON(line), nil
		}
		// character[]: 한 줄 원문을 한 글자 원소 배열로
		runes := []rune(line)
		elems := make([]string, len(runes))
		for i, ru := range runes {
			elems[i] = quoteJSON(string(ru))
		}
		return "[" + strings.Join(elems, ",") + "]", nil
	case "integer[]", "long[]", "double[]", "ListNode":
		n, err := r.count("크기")
		if err != nil {
			return "", err
		}
		toks, err := numTokens(r, n, p.Type)
		if err != nil {
			return "", err
		}
		return "[" + strings.Join(toks, ",") + "]", nil
	case "string[]":
		n, err := r.count("크기")
		if err != nil {
			return "", err
		}
		elems := make([]string, 0, n)
		for i := 0; i < n; i++ {
			line, err := r.next()
			if err != nil {
				return "", fmt.Errorf("문자열 %d번째 줄 없음", i+1)
			}
			elems = append(elems, quoteJSON(line))
		}
		return "[" + strings.Join(elems, ",") + "]", nil
	case "integer[][]":
		rn, err := r.count("행 수")
		if err != nil {
			return "", err
		}
		rows := make([]string, 0, rn)
		for i := 0; i < rn; i++ {
			toks, err := rowTokens(r)
			if err != nil {
				return "", fmt.Errorf("%d행: %w", i+1, err)
			}
			rows = append(rows, "["+strings.Join(toks, ",")+"]")
		}
		return "[" + strings.Join(rows, ",") + "]", nil
	case "character[][]":
		line, err := r.next()
		if err != nil {
			return "", fmt.Errorf("`r c` 줄 없음")
		}
		f := strings.Fields(line)
		if len(f) != 2 {
			return "", fmt.Errorf("`r c` 형식이 아님: %q", line)
		}
		rn, err1 := strconv.Atoi(f[0])
		cn, err2 := strconv.Atoi(f[1])
		if err1 != nil || err2 != nil || rn < 0 || cn < 0 {
			return "", fmt.Errorf("`r c`가 음수가 아닌 정수가 아님: %q", line)
		}
		rows := make([]string, 0, rn)
		for i := 0; i < rn; i++ {
			line, err := r.next()
			if err != nil {
				return "", fmt.Errorf("grid %d행 없음", i+1)
			}
			if len([]rune(line)) != cn {
				return "", fmt.Errorf("%d행 길이가 c=%d가 아님: %q", i+1, cn, line)
			}
			var b strings.Builder
			for _, ru := range line {
				b.WriteString(quoteJSON(string(ru)) + ",")
			}
			rows = append(rows, "["+strings.TrimSuffix(b.String(), ",")+"]")
		}
		return "[" + strings.Join(rows, ",") + "]", nil
	case "TreeNode":
		n, err := r.count("노드 수")
		if err != nil {
			return "", err
		}
		line, err := r.next()
		if err != nil {
			return "", fmt.Errorf("레벨순서 줄 없음")
		}
		f := strings.Fields(line)
		if len(f) != n {
			return "", fmt.Errorf("노드 수 %d와 토큰 %d개 불일치", n, len(f))
		}
		for _, tok := range f {
			if tok == "null" {
				continue
			}
			if _, err := strconv.ParseInt(tok, 10, 64); err != nil {
				return "", fmt.Errorf("정수/null이 아닌 토큰: %q", tok)
			}
		}
		return "[" + strings.Join(f, ",") + "]", nil
	}
	return "", fmt.Errorf("역변환 미지원 타입: %s", p.Type)
}

func numLine(r *lineReader, typ string) (string, error) {
	line, err := r.next()
	if err != nil {
		return "", fmt.Errorf("값 줄 없음")
	}
	tok := strings.TrimSpace(line)
	if typ == "double" {
		if _, err := strconv.ParseFloat(tok, 64); err != nil {
			return "", fmt.Errorf("실수가 아님: %q", line)
		}
	} else if _, err := strconv.ParseInt(tok, 10, 64); err != nil {
		return "", fmt.Errorf("정수가 아님: %q", line)
	}
	return tok, nil
}

// numTokens는 n개 숫자(또는 빈 줄 n=0)를 읽어 원본 토큰으로.
func numTokens(r *lineReader, n int, typ string) ([]string, error) {
	line, err := r.next()
	if err != nil && !(n == 0 && err == io.EOF) {
		return nil, fmt.Errorf("값 줄 없음")
	}
	if n == 0 && line == "" {
		return nil, nil
	}
	f := strings.Fields(line)
	if len(f) != n {
		return nil, fmt.Errorf("크기 %d와 값 %d개 불일치", n, len(f))
	}
	for _, tok := range f {
		if typ == "double" {
			if _, err := strconv.ParseFloat(tok, 64); err != nil {
				return nil, fmt.Errorf("실수가 아님: %q", tok)
			}
		} else if _, err := strconv.ParseInt(tok, 10, 64); err != nil {
			return nil, fmt.Errorf("정수가 아님: %q", tok)
		}
	}
	return f, nil
}

func rowTokens(r *lineReader) ([]string, error) {
	line, err := r.next()
	if err != nil {
		return nil, fmt.Errorf("행 줄 없음")
	}
	f := strings.Fields(line)
	if len(f) == 0 {
		return nil, fmt.Errorf("빈 행")
	}
	k, err := strconv.Atoi(f[0])
	if err != nil || k < 0 {
		return nil, fmt.Errorf("행 앞의 k가 음수가 아닌 정수가 아님: %q", f[0])
	}
	if len(f)-1 != k {
		return nil, fmt.Errorf("k=%d와 값 %d개 불일치", k, len(f)-1)
	}
	for _, tok := range f[1:] {
		if _, err := strconv.ParseInt(tok, 10, 64); err != nil {
			return nil, fmt.Errorf("정수가 아님: %q", tok)
		}
	}
	return f[1:], nil
}

func quoteJSON(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
