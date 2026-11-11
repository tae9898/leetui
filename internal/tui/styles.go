package tui

import "github.com/charmbracelet/lipgloss"

// 색상 체계 — 터미널 코크핏.
// 포커스는 테두리 색으로, 상태는 텍스트 색으로 인코딩한다. 장식은 없다.
var (
	cBorderDim = lipgloss.Color("238") // 비포커스 패널 테두리
	cBorderOn  = lipgloss.Color("39")  // 포커스 패널 테두리 (청색)
	cTitle     = lipgloss.Color("81")  // 패널 제목 (하늘)
	cAccent    = lipgloss.Color("207") // 앱 이름 강조
	cKey       = lipgloss.Color("75")  // 키바의 키
	cDim       = lipgloss.Color("245")
	cDimmer    = lipgloss.Color("240")
	cOK        = lipgloss.Color("42")  // PASS / ok
	cFail      = lipgloss.Color("203") // FAIL / TLE / RE / fail
	cWarn      = lipgloss.Color("214") // due / 경고
	cSelBG     = lipgloss.Color("236") // 선택 행 배경
	cSelFG     = lipgloss.Color("15")

	stNormal = lipgloss.NewStyle()
	stTitle  = lipgloss.NewStyle().Bold(true).Foreground(cTitle)
	stAccent = lipgloss.NewStyle().Bold(true).Foreground(cAccent)
	stDim    = lipgloss.NewStyle().Foreground(cDim)
	stDimmer = lipgloss.NewStyle().Foreground(cDimmer)
	stOK     = lipgloss.NewStyle().Foreground(cOK)
	stFail   = lipgloss.NewStyle().Foreground(cFail)
	stWarn   = lipgloss.NewStyle().Foreground(cWarn)
	stBanner = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	stKey    = lipgloss.NewStyle().Bold(true).Foreground(cKey)

	stSelRow = lipgloss.NewStyle().Background(cSelBG).Foreground(cSelFG)
)
