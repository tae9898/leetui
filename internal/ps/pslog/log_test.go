package pslog

import (
	"os"
	"path/filepath"
	"testing"
)

// fixtureEntries: testdata/log.tsv 파싱 결과. 손상 행 2개(컬럼 부족, 분 정수 아님)를 포함.
func fixtureEntries(t *testing.T) []Entry {
	t.Helper()
	entries, skipped, err := Parse(filepath.Join("testdata", "log.tsv"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if skipped != 2 {
		t.Fatalf("skipped = %d, want 2", skipped)
	}
	return entries
}

func TestParse(t *testing.T) {
	entries := fixtureEntries(t)
	if len(entries) != 7 {
		t.Fatalf("len(entries) = %d, want 7", len(entries))
	}
	want := Entry{
		Date: "2026-11-01", Site: "lc", Problem: "0056-merge-intervals",
		Result: ResultOK, Minutes: 12, Due: "-", Note: "투 포인터 병합",
	}
	if entries[0] != want {
		t.Errorf("entries[0] = %+v, want %+v", entries[0], want)
	}
	// 메모에 탭이 있으면 6번째 탭 이후 전체가 Note
	if got := entries[3].Note; got != "메모에\t탭\t포함" {
		t.Errorf("tab note = %q, want %q", got, "메모에\t탭\t포함")
	}
	// 7번째 컬럼이 빈 행과 아예 없는 행 모두 Note ""
	if entries[4].Note != "" || entries[5].Note != "" {
		t.Errorf("빈 메모 = %q, %q, want 빈 문자열", entries[4].Note, entries[5].Note)
	}
}

func TestParseMissingFile(t *testing.T) {
	if _, _, err := Parse(filepath.Join(t.TempDir(), "없음.tsv")); err == nil {
		t.Error("없는 파일은 에러여야 한다")
	}
}

func TestParseNoTrailingNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log.tsv")
	row := "2026-11-01\tlc\tp\tok\t5\t-\t메모"
	if err := os.WriteFile(path, []byte(row), 0o644); err != nil {
		t.Fatal(err)
	}
	entries, skipped, err := Parse(path)
	if err != nil || skipped != 0 || len(entries) != 1 || entries[0].Note != "메모" {
		t.Fatalf("Parse = %+v, skipped %d, err %v", entries, skipped, err)
	}
}

func TestMarshalGolden(t *testing.T) {
	e := Entry{
		Date: "2026-11-11", Site: "lc", Problem: "0056-merge-intervals",
		Result: ResultOK, Minutes: 12, Due: "-", Note: "투 포인터 병합",
	}
	golden := "2026-11-11\tlc\t0056-merge-intervals\tok\t12\t-\t투 포인터 병합\n"
	if got := string(e.Marshal()); got != golden {
		t.Errorf("Marshal = %q, want %q", got, golden)
	}
}

func TestAppendNewFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log.tsv")
	e := Entry{
		Date: "2026-11-11", Site: "hr", Problem: "algo-1",
		Result: ResultFail, Minutes: 30, Due: "2026-11-14", Note: "메모",
	}
	if err := Append(path, e); err != nil {
		t.Fatal(err)
	}
	golden := "2026-11-11\thr\talgo-1\tfail\t30\t2026-11-14\t메모\n"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != golden {
		t.Errorf("파일 = %q, want %q", data, golden)
	}
}

func TestAppendMissingTrailingNewline(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log.tsv")
	old := "2026-11-01\tlc\told\tok\t1\t-\tnote" // 끝 개행 없음
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	e := Entry{Date: "2026-11-11", Site: "lc", Problem: "new", Result: ResultOK, Minutes: 0, Due: "-"}
	if err := Append(path, e); err != nil {
		t.Fatal(err)
	}
	golden := old + "\n2026-11-11\tlc\tnew\tok\t0\t-\t\n"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != golden {
		t.Errorf("파일 = %q, want %q", data, golden)
	}
}

func TestAppendAfterNewlineTerminatedLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log.tsv")
	first := "2026-11-01\tlc\told\tok\t1\t-\tnote\n"
	if err := os.WriteFile(path, []byte(first), 0o644); err != nil {
		t.Fatal(err)
	}
	e := Entry{Date: "2026-11-02", Site: "lc", Problem: "new", Result: ResultAgain, Minutes: 7, Due: "2026-11-09", Note: "n"}
	if err := Append(path, e); err != nil {
		t.Fatal(err)
	}
	// 이미 개행으로 끝나면 빈 행이 끼어들지 않는다
	golden := first + "2026-11-02\tlc\tnew\tagain\t7\t2026-11-09\tn\n"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != golden {
		t.Errorf("파일 = %q, want %q", data, golden)
	}
}

func TestAppendSanitizesNoteAndMinutes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "log.tsv")
	e := Entry{
		Date: "2026-11-11", Site: "lc", Problem: "p", Result: ResultAgain,
		Minutes: -3, Due: "2026-11-18", Note: "\t다양한\t메모\r\n다줄 ",
	}
	if err := Append(path, e); err != nil {
		t.Fatal(err)
	}
	// 탭/CR/LF는 공백으로, 양끝 공백은 trim, 음수 분은 0
	golden := "2026-11-11\tlc\tp\tagain\t0\t2026-11-18\t다양한 메모  다줄\n"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != golden {
		t.Errorf("파일 = %q, want %q", data, golden)
	}
}
