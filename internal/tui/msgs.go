package tui

import (
	"pstui/internal/ps/problem"
	"pstui/internal/ps/pslog"
	"pstui/internal/ps/runner"
)

// 화면 전환
type openHomeMsg struct{}
type openProblemsMsg struct{}
type openWorkbenchMsg struct{ Prob problem.Problem }
type openRunMsg struct {
	Prob problem.Problem
	Fast bool // F 키: fast 빌드 후 시간 측정만
}
type openDoneMsg struct {
	Prob     problem.Problem
	Default  string // "ok" | "fail" — Run 결과 기본값
}
type openNewMsg struct{}
type newCreatedMsg struct {
	Prob problem.Problem
	Err  error
}
type openRetryMsg struct{ Prob problem.Problem }
type openLogMsg struct{}
type openStatsMsg struct{}
type editorDoneMsg struct {
	Prob problem.Problem
	Err  error
}

// 데이터 로드 (도메인 호출은 전부 tea.Cmd 안에서)
type logLoadedMsg struct {
	Entries  []pslog.Entry
	Skipped  int
	Err      error
}
type problemsLoadedMsg struct {
	Problems []problem.Problem
	Err      error
}
type appendDoneMsg struct{ Err error }

// Run 화면 진행
type compileDoneMsg struct {
	Stderr string
	Err    error
}
type casesLoadedMsg struct {
	Cases []runner.Case
	Err   error
}
type caseStartedMsg struct{ Index int }
type caseDoneMsg struct {
	Index  int
	Result runner.CaseResult
}
type runCancelledMsg struct{}

// Workbench 갱신
type filesLoadedMsg struct {
	Files []FileEntry
	Err   error
}
type startResetMsg struct{ Err error }
