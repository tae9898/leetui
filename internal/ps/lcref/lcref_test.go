package lcref

import (
	"context"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		in             string
		site, id, name string
		ok             bool
	}{
		{"https://leetcode.com/problems/merge-intervals/", "lc", "", "merge-intervals", true},
		{"https://leetcode.cn/problems/two-sum/description", "lc", "", "two-sum", true},
		{"https://www.hackerrank.com/challenges/repeated-string/problem", "hr", "", "repeated-string", true},
		{"56. Merge Intervals", "lc", "0056", "merge-intervals", true},
		{"56 merge intervals", "lc", "0056", "merge-intervals", true},
		{"1) Two Sum", "lc", "0001", "two-sum", true},
		{"merge-intervals", "", "", "merge-intervals", true},
		{"", "", "", "", false},
		{"not a ref !!", "", "", "", false},
	}
	for _, c := range cases {
		got, ok := Parse(c.in)
		if ok != c.ok || got.Site != c.site || got.ID != c.id || got.Name != c.name {
			t.Errorf("Parse(%q) = %+v ok=%v, want site=%q id=%q name=%q ok=%v",
				c.in, got, ok, c.site, c.id, c.name, c.ok)
		}
	}
}

func TestSlugify(t *testing.T) {
	cases := [][2]string{
		{"Merge Intervals", "merge-intervals"},
		{"String to Integer (atoi)", "string-to-integer-atoi"},
		{"  Best  Time   ", "best-time"},
		{"3Sum", "3sum"},
	}
	for _, c := range cases {
		if got := Slugify(c[0]); got != c[1] {
			t.Errorf("Slugify(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestLookupNumber(t *testing.T) {
	if testing.Short() {
		t.Skip("네트워크 조회는 -short에서 스킵")
	}
	n, err := LookupNumber(context.Background(), "merge-intervals")
	if err != nil {
		t.Skipf("네트워크 실패: %v", err)
	}
	if PadNumber(n) != "0056" {
		t.Fatalf("merge-intervals 번호 = %q, want 56", n)
	}
}
