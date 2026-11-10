package linediff

import (
	"reflect"
	"testing"
)

func TestDiffIdentical(t *testing.T) {
	exp := []string{"a", "b", "c"}
	want := []Line{{Keep, "a"}, {Keep, "b"}, {Keep, "c"}}
	if got := Diff(exp, []string{"a", "b", "c"}); !reflect.DeepEqual(got, want) {
		t.Errorf("Diff = %+v, want %+v", got, want)
	}
}

func TestDiffEmpty(t *testing.T) {
	if got := Diff(nil, nil); len(got) != 0 {
		t.Errorf("Diff(nil, nil) = %+v", got)
	}
	if got := Diff(nil, []string{"a"}); !reflect.DeepEqual(got, []Line{{Add, "a"}}) {
		t.Errorf("Diff(nil, [a]) = %+v", got)
	}
	if got := Diff([]string{"a"}, nil); !reflect.DeepEqual(got, []Line{{Del, "a"}}) {
		t.Errorf("Diff([a], nil) = %+v", got)
	}
}

func TestDiffInsertion(t *testing.T) {
	got := Diff([]string{"a", "c"}, []string{"a", "b", "c"})
	want := []Line{{Keep, "a"}, {Add, "b"}, {Keep, "c"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Diff = %+v, want %+v", got, want)
	}
}

func TestDiffDeletion(t *testing.T) {
	got := Diff([]string{"a", "b", "c"}, []string{"a", "c"})
	want := []Line{{Keep, "a"}, {Del, "b"}, {Keep, "c"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Diff = %+v, want %+v", got, want)
	}
}

func TestDiffSubstitution(t *testing.T) {
	got := Diff([]string{"a", "b", "c", "d"}, []string{"a", "x", "c"})
	want := []Line{{Keep, "a"}, {Del, "b"}, {Add, "x"}, {Keep, "c"}, {Del, "d"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Diff = %+v, want %+v", got, want)
	}
}

func TestDiffAtLimit(t *testing.T) {
	// 정확히 2000줄이면 전체 diff를 한다
	exp := make([]string, 2000)
	for i := range exp {
		exp[i] = "x"
	}
	act := append([]string(nil), exp...)
	act[1999] = "y"
	got := Diff(exp, act)
	if len(got) != 2001 {
		t.Fatalf("len = %d, want 2001", len(got))
	}
	if got[1998] != (Line{Keep, "x"}) || got[1999] != (Line{Del, "x"}) || got[2000] != (Line{Add, "y"}) {
		t.Errorf("끝부분 = %+v", got[1998:])
	}
}

func TestDiffOverLimit(t *testing.T) {
	exp := make([]string, 2001)
	for i := range exp {
		exp[i] = "x"
	}
	act := append([]string(nil), exp...)
	act[1499] = "y"

	// 기대 초과: 첫 불일치 줄만 요약
	got := Diff(exp, act)
	want := []Line{{Add, "첫 불일치 줄 1500"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Diff = %+v, want %+v", got, want)
	}
	// 실제 초과 + 접두사 일치: 길이 차이가 첫 불일치
	shortAct := append([]string(nil), exp[:3]...)
	got = Diff(shortAct, act)
	want = []Line{{Add, "첫 불일치 줄 4"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Diff = %+v, want %+v", got, want)
	}
	// 2000줄 초과인데 완전히 같으면 0
	got = Diff(exp, exp)
	want = []Line{{Add, "첫 불일치 줄 0"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Diff = %+v, want %+v", got, want)
	}
}

func TestWhitespaceOnly(t *testing.T) {
	cases := []struct {
		exp, act string
		want     bool
	}{
		{"a\nb\n", "a\nb\n", true},
		{"a \nb\t\n", "a\nb\n", true},  // 줄 끝 공백
		{"a\nb\r\n", "a\nb\n", true},   // CR
		{"a\nb", "a\nb\n", true},       // 끝 개행 유무
		{"", "", true},                 // 둘 다 빈 입력
		{"", "\n", false},              // 빈 입력 vs 빈 줄 1개(줄 수 차이)
		{"a b\n", "a  b\n", false},     // 줄 내부 공백 차이
		{"a\nb\n", "a\nb\nc\n", false}, // 줄 수 차이
		{"a\n\nb\n", "a\nb\n", false},  // 빈 줄 추가
		{"a\r\nb\n", "ab\n", false},    // CR 위치가 줄 끝이 아님
	}
	for _, c := range cases {
		if got := WhitespaceOnly([]byte(c.exp), []byte(c.act)); got != c.want {
			t.Errorf("WhitespaceOnly(%q, %q) = %v, want %v", c.exp, c.act, got, c.want)
		}
	}
}
