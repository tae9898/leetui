package pslog

import (
	"testing"
	"time"
)

func TestDueDate(t *testing.T) {
	now := time.Date(2026, 11, 11, 10, 30, 0, 0, time.Local)
	cases := []struct {
		result Result
		want   string
	}{
		{ResultOK, "-"},
		{ResultFail, "2026-11-14"},
		{ResultAgain, "2026-11-18"},
		{Result("typo"), "-"},
	}
	for _, c := range cases {
		if got := DueDate(c.result, now); got != c.want {
			t.Errorf("DueDate(%q) = %q, want %q", c.result, got, c.want)
		}
	}
	if got := DueDate(ResultFail, time.Date(2026, 11, 30, 23, 59, 59, 0, time.Local)); got != "2026-12-03" {
		t.Errorf("달 넘어감 = %q, want 2026-12-03", got)
	}
	if got := DueDate(ResultAgain, time.Date(2026, 12, 30, 0, 0, 0, 0, time.Local)); got != "2027-01-06" {
		t.Errorf("연 넘어감 = %q, want 2027-01-06", got)
	}
}

func TestDue(t *testing.T) {
	entries := fixtureEntries(t)
	// today=2026-11-08: due == today 경계 포함, lc/0056은 마지막 fail 행이 우선
	got := Due(entries, "2026-11-08")
	want := []DueItem{
		{"2026-11-05", "hr/algo-1", ResultFail},
		{"2026-11-07", "mock/contest-A", ResultFail},
		{"2026-11-08", "lc/0056-merge-intervals", ResultFail},
	}
	if len(got) != len(want) {
		t.Fatalf("len(Due) = %d, want %d (%+v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Due[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
	if got := Due(entries, "2026-11-07"); len(got) != 2 {
		t.Errorf("today=2026-11-07이면 2건이어야 한다, got %+v", got)
	}
	if got := Due(entries, "2026-11-04"); len(got) != 0 {
		t.Errorf("today=2026-11-04이면 없어야 한다, got %+v", got)
	}
}

func TestDueLastEntryWins(t *testing.T) {
	entries := []Entry{
		{Site: "lc", Problem: "p", Result: ResultFail, Due: "2026-11-01"},
		{Site: "lc", Problem: "p", Result: ResultOK, Due: "-"},
	}
	if got := Due(entries, "2026-11-05"); len(got) != 0 {
		t.Errorf("마지막 ok 행이 우선이므로 비어야 한다, got %+v", got)
	}
}

func TestDueSort(t *testing.T) {
	entries := []Entry{
		{Site: "zz", Problem: "b", Result: ResultFail, Due: "2026-11-03"},
		{Site: "aa", Problem: "c", Result: ResultFail, Due: "2026-11-05"},
		{Site: "aa", Problem: "a", Result: ResultFail, Due: "2026-11-05"},
		{Site: "mm", Problem: "x", Result: ResultOK, Due: "-"}, // due "-" 제외
	}
	got := Due(entries, "2026-12-01")
	want := []DueItem{
		{"2026-11-03", "zz/b", ResultFail},
		{"2026-11-05", "aa/a", ResultFail},
		{"2026-11-05", "aa/c", ResultFail},
	}
	if len(got) != len(want) {
		t.Fatalf("len(Due) = %d, want %d (%+v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Due[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestFormatDue(t *testing.T) {
	got := FormatDue(DueItem{Due: "2026-11-05", Key: "hr/algo-1", Result: ResultFail})
	if want := "2026-11-05 hr/algo-1 fail"; got != want {
		t.Errorf("FormatDue = %q, want %q", got, want)
	}
}
