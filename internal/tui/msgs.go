package tui

import (
	"leetui/internal/ps/problem"
	"leetui/internal/ps/pslog"
	"leetui/internal/ps/runner"
	"leetui/internal/ps/stmt"
)

// 화면 전환
type openHomeMsg struct{}
type openProblemsMsg struct{}
type openWorkbenchMsg struct{ Prob problem.Problem }
type openRunMsg struct {
	Prob problem.Problem
	Fast bool   // F 키: fast 빌드 후 시간 측정만
	Only string // 커스텀 입력 등 이 파일만 실행 (빈 값=전체)
}
type openDoneMsg struct {
	Prob    problem.Problem
	Default string // "ok" | "fail" — Run 결과 기본값
}
type openPlanMsg struct{}
type openNewMsg struct{}
type newCreatedMsg struct {
	Prob problem.Problem
	Err  error
}
type lcNumberMsg struct {
	Slug   string
	Number string
	Err    error
}
type openStmtMsg struct {
	Site string
	Slug string
}
type stmtReadyMsg struct {
	Path string
	Err  error
}
type stmtFilledMsg struct {
	Res  stmt.FillResult
	Prob problem.Problem
	Open bool // 피커에서 눌렀으면 채운 뒤 작업 화면으로
	Err  error
}
type oracleDoneMsg struct {
	Prob problem.Problem
	Res  stmt.OracleResult
	Err  error
}
type viewerDoneMsg struct{ Err error }
type openRetryMsg struct{ Prob problem.Problem }
type openLogMsg struct{}
type openStatsMsg struct{}
type editorDoneMsg struct {
	Prob problem.Problem
	File string // 편집한 파일 (custom.in 후처리용)
	Err  error
}

// 데이터 로드 (도메인 호출은 전부 tea.Cmd 안에서)
type logLoadedMsg struct {
	Entries []pslog.Entry
	Skipped int
	Err     error
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
	Files  []FileEntry
	Oracle map[string]bool // 오라클이 채운 케이스명 (.oracle)
	Err    error
}
type startResetMsg struct{ Err error }
