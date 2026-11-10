// Package pslog는 log.tsv의 파싱·append·재도전일(due)·통계를 담당한다
// (TUI-PLAN.md 2.3 pslog 표). 쓰기는 fish psdone과 바이트 단위로 동일해야 한다.
package pslog

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Result는 psdone의 결과값.
type Result string

const (
	ResultOK    Result = "ok"
	ResultFail  Result = "fail"
	ResultAgain Result = "again"
)

// Entry는 log.tsv 1행. 필드 순서가 컬럼 순서와 같다.
type Entry struct {
	Date    string // YYYY-MM-DD
	Site    string
	Problem string
	Result  Result
	Minutes int
	Due     string // YYYY-MM-DD 또는 "-"
	Note    string
}

// Marshal은 fish의 printf '%s\t...\t%s\n' 7컬럼 행과 바이트 단위로 동일하다.
func (e Entry) Marshal() []byte {
	return []byte(fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%s\t%s\n",
		e.Date, e.Site, e.Problem, e.Result, e.Minutes, e.Due, e.Note))
}

// Parse는 log.tsv를 읽어 유효 행만 반환한다. 컬럼 6개 미만이거나 Minutes가
// 정수가 아니면 그 행을 건너뛰고 skipped로 세며, 빈 행은 무시한다.
func Parse(path string) (entries []Entry, skipped int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 7)
		if len(parts) < 6 {
			skipped++
			continue
		}
		minutes, err := strconv.Atoi(parts[4])
		if err != nil {
			skipped++
			continue
		}
		note := ""
		if len(parts) == 7 {
			note = parts[6] // 메모에 탭이 섞였으면 나머지 전체
		}
		entries = append(entries, Entry{
			Date:    parts[0],
			Site:    parts[1],
			Problem: parts[2],
			Result:  Result(parts[3]),
			Minutes: minutes,
			Due:     parts[5],
			Note:    note,
		})
	}
	return entries, skipped, nil
}

// Append는 행 1개를 추가한다. 파일이 비어 있지 않은데 마지막 바이트가
// '\n'이 아니면 먼저 '\n'을 보정하고, 메모의 탭/CR/LF는 치환·trim한다.
func Append(path string, e Entry) error {
	e.Note = sanitizeNote(e.Note)
	if e.Minutes < 0 {
		e.Minutes = 0
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if fi, err := f.Stat(); err == nil && fi.Size() > 0 && !endsWithNewline(path, fi.Size()) {
		if _, err := f.Write([]byte{'\n'}); err != nil {
			return err
		}
	}
	_, err = f.Write(e.Marshal())
	return err
}

// sanitizeNote: 탭/CR/LF를 공백으로 치환한 뒤 양끝 공백을 제거.
func sanitizeNote(note string) string {
	r := strings.NewReplacer("\t", " ", "\r", " ", "\n", " ")
	return strings.TrimSpace(r.Replace(note))
}

// endsWithNewline: 마지막 바이트 확인. 확인에 실패하면 보정하지 않는다(true).
func endsWithNewline(path string, size int64) bool {
	rf, err := os.Open(path)
	if err != nil {
		return true
	}
	defer rf.Close()
	b := make([]byte, 1)
	if _, err := rf.ReadAt(b, size-1); err != nil {
		return true
	}
	return b[0] == '\n'
}
