package pslog

import (
	"testing"
	"time"
)

func TestRecent(t *testing.T) {
	entries := []Entry{{Problem: "a"}, {Problem: "b"}, {Problem: "c"}, {Problem: "d"}, {Problem: "e"}}
	got := Recent(entries, 3)
	if len(got) != 3 || got[0].Problem != "c" || got[1].Problem != "d" || got[2].Problem != "e" {
		t.Errorf("Recent(3) = %+v, want c,d,e 순서", got)
	}
	if all := Recent(entries, 10); len(all) != 5 || all[0].Problem != "a" {
		t.Errorf("Recent(10) = %+v, want 전체 원순서", all)
	}
	if none := Recent(entries, 0); len(none) != 0 {
		t.Errorf("Recent(0) = %+v, want 빈 결과", none)
	}
}

func TestWeek(t *testing.T) {
	mk := func(date string, r Result, minutes int) Entry {
		return Entry{Date: date, Result: r, Minutes: minutes}
	}
	entries := []Entry{
		mk("2026-11-08", ResultOK, 60),   // 지난 주 일요일
		mk("2026-11-09", ResultOK, 30),   // 월요일
		mk("2026-11-10", ResultFail, 45), // 화요일
		mk("2026-11-11", ResultAgain, 20),
		mk("2026-11-11", ResultOK, 10),
		mk("2026-11-15", ResultOK, 24), // 이번 주 일요일(포함)
		mk("2026-11-16", ResultOK, 99), // 다음 주 월요일
		mk("garbage", ResultOK, 1),     // 날짜 아닌 값
	}
	// 2026-11-11은 수요일 → 주는 11-09 ~ 11-15
	want := WeekStat{Solved: 5, Ok: 3, Fail: 1, Again: 1, AvgMinutes: 25} // 129/5 = 25(정수 나눗셈)
	if got := Week(entries, time.Date(2026, 11, 11, 9, 0, 0, 0, time.Local)); got != want {
		t.Errorf("Week(수) = %+v, want %+v", got, want)
	}
	// 일요일 기준으로도 같은 주
	if got := Week(entries, time.Date(2026, 11, 15, 23, 0, 0, 0, time.Local)); got != want {
		t.Errorf("Week(일) = %+v, want %+v", got, want)
	}
	// 주에 행이 없으면 전부 0 (0 나누기 방지)
	if got := Week(entries, time.Date(2020, 1, 6, 0, 0, 0, 0, time.Local)); got != (WeekStat{}) {
		t.Errorf("Week(빈 주) = %+v, want 0", got)
	}
}
