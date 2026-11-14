package lcfmt

import (
	"strings"
	"testing"
)

// twoSum 메타 (녹화된 two-sum metaData와 같은 구조).
func meta(params []Param, ret Return) Meta {
	return Meta{Name: "f", Params: params, Return: ret}
}

func TestParseMeta(t *testing.T) {
	m, err := ParseMeta(`{
	  "name": "merge",
	  "params": [
	    {"name": "nums1", "type": "integer[]"},
	    {"name": "m", "type": "integer"},
	    {"name": "nums2", "type": "integer[]"},
	    {"name": "n", "type": "integer"}
	  ],
	  "return": {"type": "void"},
	  "output": {"paramindex": 0}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "merge" || len(m.Params) != 4 {
		t.Fatalf("파싱 결과: %+v", m)
	}
	if m.Return.Type != "void" || !m.Return.HasOutParam || m.Return.OutParam != 0 {
		t.Fatalf("return/output: %+v", m.Return)
	}
}

func TestParseMetaDesign(t *testing.T) {
	m, err := ParseMeta(`{"classname":"LRUCache","params":[{"name":"inputs","type":"integer[]"}],"return":{"type":"list<String>"}}`)
	if err != nil {
		t.Fatal(err)
	}
	if m.ClassName != "LRUCache" {
		t.Fatalf("classname: %+v", m)
	}
	ok, reason := Supported(m, "")
	if ok || !strings.Contains(reason, "design") {
		t.Fatalf("design 문제가 지원으로 판정됨: %q", reason)
	}
}

func TestSupported(t *testing.T) {
	intArr := meta([]Param{{Name: "a", Type: "integer[]"}}, Return{Type: "integer[]"})
	if ok, r := Supported(intArr, "plain"); !ok {
		t.Fatalf("integer[] 지원해야 함: %q", r)
	}
	// two-sum류: 고정 형태 답 + "any order" 문구 — 지원 유지 (LC 정답과 순서 일치)
	if ok, r := Supported(intArr, "You can return the answer in any order."); !ok {
		t.Fatalf("any order라도 답이 모음이 아니면 지원: %q", r)
	}
	// 3sum류: 답 자체가 모음이고 any order — 폴백
	twoD := meta([]Param{{Name: "a", Type: "integer[]"}}, Return{Type: "integer[][]"})
	if ok, r := Supported(twoD, "return the triplets ... in any order"); ok || !strings.Contains(r, "any order") {
		t.Fatalf("모음 답 + any order는 폴백: %q", r)
	}
	// double 반환 폴백
	if ok, r := Supported(meta(nil, Return{Type: "double"}), ""); ok || !strings.Contains(r, "double") {
		t.Fatalf("double 반환 폴백: %q", r)
	}
	// 미지원 파라미터 타입
	if ok, r := Supported(meta([]Param{{Name: "x", Type: "vector<vector<string>>"}}, Return{Type: "integer"}), ""); ok || !strings.Contains(r, "미지원") {
		t.Fatalf("미지원 타입 폴백: %q", r)
	}
	// void인데 출력 파라미터 없음
	if ok, r := Supported(meta([]Param{{Name: "x", Type: "integer[]"}}, Return{Type: "void"}), ""); ok || !strings.Contains(r, "void") {
		t.Fatalf("void 폴백: %q", r)
	}
}

// 타입별 LC 리터럴 → 표준 stdin/stdout → LC 리터럴 왕복.
func TestConversions(t *testing.T) {
	cases := []struct {
		desc   string
		params []Param
		ret    Return
		lcIn   string // 파라미터 리터럴을 \n로 이음
		lcOut  string
		stdin  string
		stdout string
	}{
		{
			desc:   "스칼라 혼합 (two-sum)",
			params: []Param{{Name: "nums", Type: "integer[]"}, {Name: "target", Type: "integer"}},
			ret:    Return{Type: "integer[]"},
			lcIn:   "[2,7,11,15]\n9",
			lcOut:  "[0,1]",
			stdin:  "4\n2 7 11 15\n9\n",
			stdout: "0 1\n",
		},
		{
			desc:   "빈 배열 ListNode",
			params: []Param{{Name: "head", Type: "ListNode"}},
			ret:    Return{Type: "ListNode"},
			lcIn:   "[]",
			lcOut:  "[]",
			stdin:  "0\n\n",
			stdout: "\n",
		},
		{
			desc:   "ListNode 왕복",
			params: []Param{{Name: "head", Type: "ListNode"}},
			ret:    Return{Type: "ListNode"},
			lcIn:   "[1,2,3,4,5]",
			lcOut:  "[5,4,3,2,1]",
			stdin:  "5\n1 2 3 4 5\n",
			stdout: "5 4 3 2 1\n",
		},
		{
			desc:   "문자열 두 개 (valid-anagram)",
			params: []Param{{Name: "s", Type: "string"}, {Name: "t", Type: "string"}},
			ret:    Return{Type: "boolean"},
			lcIn:   "\"anagram\"\n\"nagaram\"",
			lcOut:  "true",
			stdin:  "anagram\nnagaram\n",
			stdout: "true\n",
		},
		{
			desc:   "2차원 (merge-intervals) 들쭉날쭉",
			params: []Param{{Name: "intervals", Type: "integer[][]"}},
			ret:    Return{Type: "integer[][]"},
			lcIn:   "[[1,3],[2,6],[8,10],[15,18]]",
			lcOut:  "[[1,6],[8,10],[15,18]]",
			stdin:  "4\n2 1 3\n2 2 6\n2 8 10\n2 15 18\n",
			stdout: "1 6\n8 10\n15 18\n",
		},
		{
			desc:   "grid (number-of-islands)",
			params: []Param{{Name: "grid", Type: "character[][]"}},
			ret:    Return{Type: "integer"},
			lcIn:   "[[\"1\",\"1\",\"0\"],[\"0\",\"1\",\"0\"],[\"0\",\"0\",\"1\"]]",
			lcOut:  "2",
			stdin:  "3 3\n110\n010\n001\n",
			stdout: "2\n",
		},
		{
			desc:   "TreeNode",
			params: []Param{{Name: "root", Type: "TreeNode"}},
			ret:    Return{Type: "integer"},
			lcIn:   "[3,9,20,null,null,15,7]",
			lcOut:  "3",
			stdin:  "7\n3 9 20 null null 15 7\n",
			stdout: "3\n",
		},
		{
			desc:   "TreeNode 뒤쪽 null 제거",
			params: []Param{{Name: "root", Type: "TreeNode"}},
			ret:    Return{Type: "TreeNode"},
			lcIn:   "[2,1,3,null,null,4]",
			lcOut:  "[2,1,3,null,null,4,null]",
			stdin:  "6\n2 1 3 null null 4\n",
			stdout: "2 1 3 null null 4\n",
		},
		{
			desc:   "string[] (group-anagrams 입력으로)",
			params: []Param{{Name: "strs", Type: "string[]"}},
			ret:    Return{Type: "integer"},
			lcIn:   "[\"eat\",\"tea\"]",
			lcOut:  "1",
			stdin:  "2\neat\ntea\n",
			stdout: "1\n",
		},
		{
			desc:   "void 제자리 (merge-sorted-array)",
			params: []Param{{Name: "nums1", Type: "integer[]"}, {Name: "m", Type: "integer"}, {Name: "nums2", Type: "integer[]"}, {Name: "n", Type: "integer"}},
			ret:    Return{Type: "void", OutParam: 0, HasOutParam: true},
			lcIn:   "[1,2,3,0,0,0]\n3\n[2,5,6]\n3",
			lcOut:  "[1,2,2,3,5,6]",
			stdin:  "6\n1 2 3 0 0 0\n3\n3\n2 5 6\n3\n",
			stdout: "1 2 2 3 5 6\n",
		},
		{
			desc:   "character[] (reverse-string)",
			params: []Param{{Name: "s", Type: "character[]"}},
			ret:    Return{Type: "void", OutParam: 0, HasOutParam: true},
			lcIn:   "[\"h\",\"e\",\"l\",\"l\",\"o\"]",
			lcOut:  "[\"o\",\"l\",\"l\",\"e\",\"h\"]",
			stdin:  "hello\n",
			stdout: "olleh\n",
		},
		{
			desc:   "character 한 글자",
			params: []Param{{Name: "c", Type: "character"}},
			ret:    Return{Type: "character"},
			lcIn:   "\"a\"",
			lcOut:  "\"b\"",
			stdin:  "a\n",
			stdout: "b\n",
		},
		{
			desc:   "long 정밀도 유지",
			params: []Param{{Name: "x", Type: "long[]"}},
			ret:    Return{Type: "long"},
			lcIn:   "[9007199254740993,-9007199254740993]",
			lcOut:  "9007199254740993",
			stdin:  "2\n9007199254740993 -9007199254740993\n",
			stdout: "9007199254740993\n",
		},
		{
			desc:   "이스케이프 문자열",
			params: []Param{{Name: "s", Type: "string"}},
			ret:    Return{Type: "string"},
			lcIn:   "\"a\\\"b\\\\c\"",
			lcOut:  "\"x\\ny\"",
			stdin:  "a\"b\\c\n",
			stdout: "x\ny\n", // \n은 실제 개행으로 해제 — 두 줄
		},
	}

	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			m := meta(c.params, c.ret)
			got, err := ToStdin(m, c.lcIn)
			if err != nil {
				t.Fatalf("ToStdin: %v", err)
			}
			if got != c.stdin {
				t.Fatalf("ToStdin 불일치:\n got %q\nwant %q", got, c.stdin)
			}
			out, err := ToStdout(m, c.lcOut)
			if err != nil {
				t.Fatalf("ToStdout: %v", err)
			}
			if out != c.stdout {
				t.Fatalf("ToStdout 불일치:\n got %q\nwant %q", out, c.stdout)
			}
			// 왕복: 표준 stdin → LC 리터럴
			back, err := FromStdin(m, c.stdin)
			if err != nil {
				t.Fatalf("FromStdin: %v", err)
			}
			want := strings.Split(c.lcIn, "\n")
			if strings.Join(back, "\n") != strings.Join(want, "\n") {
				t.Fatalf("왕복 불일치:\n got  %q\nwant  %q", back, want)
			}
		})
	}
}

func TestToStdinParamCountMismatch(t *testing.T) {
	m := meta([]Param{{Name: "a", Type: "integer"}, {Name: "b", Type: "integer"}}, Return{Type: "integer"})
	if _, err := ToStdin(m, "1"); err == nil {
		t.Fatal("파라미터 수 불일치인데 에러가 없음")
	}
}

func TestFromStdinErrors(t *testing.T) {
	m := meta([]Param{{Name: "a", Type: "integer[]"}}, Return{Type: "integer"})
	cases := []string{
		"3\n1 2\n", // 크기/개수 불일치
		"x\n1\n",   // 크기가 정수가 아님
		"",         // 줄 없음
		"1\nabc\n", // 값이 정수가 아님
	}
	for _, in := range cases {
		if _, err := FromStdin(m, in); err == nil {
			t.Fatalf("잘못된 stdin %q이 에러 없이 통과", in)
		}
	}
}

func TestFromStdinTrailingNewlineOnly(t *testing.T) {
	// 파일 끝 개행/빈 줄 관대하게 — 마지막 파라미터가 빈 배열이어도 읽힌다
	m := meta([]Param{{Name: "a", Type: "integer[]"}, {Name: "b", Type: "integer"}}, Return{Type: "integer"})
	lits, err := FromStdin(m, "0\n\n7\n")
	if err != nil {
		t.Fatal(err)
	}
	if lits[0] != "[]" || lits[1] != "7" {
		t.Fatalf("빈 배열 파싱: %q", lits)
	}
}

func TestFormatSpec(t *testing.T) {
	m := meta([]Param{{Name: "nums", Type: "integer[]"}, {Name: "target", Type: "integer"}}, Return{Type: "integer[]"})
	spec := FormatSpec(m)
	for _, want := range []string{"## 입력 형식 (leetui)", "nums", "target", "출력", "값 한 줄"} {
		if !strings.Contains(spec, want) {
			t.Fatalf("FormatSpec에 %q 없음:\n%s", want, spec)
		}
	}
}
