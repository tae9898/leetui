package main

import (
	"flag"
	"fmt"
	"os"

	"leetui/internal/ps/config"
	"leetui/internal/tui"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "버전 출력")
	root := flag.String("root", "", "PS_ROOT (기본: $PS_ROOT 또는 ~/ps)")
	flag.Parse()

	if *showVersion {
		fmt.Println("leetui", version)
		return
	}

	cfg := config.Load(*root)
	if err := tui.Run(cfg); err != nil {
		fmt.Fprintln(os.Stderr, "leetui:", err)
		os.Exit(1)
	}
}
