// Package config는 PS_ROOT와 외부 도구 경로 규칙을 담는다 (TUI-PLAN.md 2.3 config 표).
package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	Root     string // 문제 루트 ( --root > $PS_ROOT > ~/ps )
	Template string // $Root/template.cpp
	Editor   string // $EDITOR, 없으면 nvim
	CXX      string // $PS_CXX, 없으면 g++. 플래그는 runner가 고정
}

func Load(rootFlag string) Config {
	root := rootFlag
	if root == "" {
		root = os.Getenv("PS_ROOT")
	}
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, "ps")
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}
	cxx := os.Getenv("PS_CXX")
	if cxx == "" {
		cxx = "g++"
	}
	return Config{
		Root:     root,
		Template: filepath.Join(root, "template.cpp"),
		Editor:   editor,
		CXX:      cxx,
	}
}

// TemplateOK는 템플릿 존재 여부. 없으면 new/retry 비활성 + 홈 경고.
func (c Config) TemplateOK() bool {
	st, err := os.Stat(c.Template)
	return err == nil && !st.IsDir()
}
