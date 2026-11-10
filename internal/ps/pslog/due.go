package pslog

import (
	"sort"
	"time"
)

// DueDate는 결과별 재도전일. ok와 알 수 없는 결과는 "-".
func DueDate(result Result, now time.Time) string {
	switch result {
	case ResultFail:
		return now.AddDate(0, 0, 3).Format("2006-01-02")
	case ResultAgain:
		return now.AddDate(0, 0, 7).Format("2006-01-02")
	default:
		return "-"
	}
}

// DueItem은 psdue 출력 1행에 해당하는 데이터.
type DueItem struct {
	Due    string
	Key    string // "site/problem"
	Result Result
}

// Due는 psdue 규칙으로 오늘까지 도래한 재도전 항목을 뽑는다.
// 키(Site/Problem)마다 파일 순서상 마지막 Entry가 우선이고,
// Due != "-" && Due <= today (YYYY-MM-DD 문자열 비교)만 남긴 뒤
// (Due, Key, Result) 바이트 순으로 정렬한다.
func Due(entries []Entry, today string) []DueItem {
	last := make(map[string]DueItem)
	for _, e := range entries {
		key := e.Site + "/" + e.Problem
		last[key] = DueItem{Due: e.Due, Key: key, Result: e.Result}
	}
	items := make([]DueItem, 0, len(last))
	for _, it := range last {
		if it.Due != "-" && it.Due <= today {
			items = append(items, it)
		}
	}
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if a.Due != b.Due {
			return a.Due < b.Due
		}
		if a.Key != b.Key {
			return a.Key < b.Key
		}
		return string(a.Result) < string(b.Result)
	})
	return items
}

// FormatDue는 fish psdue의 1행 출력("due key result")과 동일한 문자열.
func FormatDue(it DueItem) string {
	return it.Due + " " + it.Key + " " + string(it.Result)
}
