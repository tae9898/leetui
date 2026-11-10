package pslog

import "time"

// WeekStat은 Home 화면의 "이번 주" 통계.
type WeekStat struct {
	Solved, Ok, Fail, Again, AvgMinutes int
}

// Recent는 파일 순서를 유지한 채 마지막 n개만 반환한다.
func Recent(entries []Entry, n int) []Entry {
	if n <= 0 {
		return nil
	}
	if n >= len(entries) {
		return entries
	}
	return entries[len(entries)-n:]
}

// Week는 today가 속한 주(월요일 시작)의 통계. AvgMinutes는 정수 나눗셈이며
// 주에 행이 없으면 0이다.
func Week(entries []Entry, today time.Time) WeekStat {
	y, m, d := today.Date()
	offset := (int(today.Weekday()) + 6) % 7 // 월=0 .. 일=6
	monday := time.Date(y, m, d-offset, 0, 0, 0, 0, today.Location())
	start := monday.Format("2006-01-02")
	end := time.Date(y, m, d-offset+7, 0, 0, 0, 0, today.Location()).Format("2006-01-02")
	var st WeekStat
	sum := 0
	for _, e := range entries {
		if e.Date < start || e.Date >= end {
			continue
		}
		st.Solved++
		sum += e.Minutes
		switch e.Result {
		case ResultOK:
			st.Ok++
		case ResultFail:
			st.Fail++
		case ResultAgain:
			st.Again++
		}
	}
	if st.Solved > 0 {
		st.AvgMinutes = sum / st.Solved
	}
	return st
}
