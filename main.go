package main

//go:generate go run ./internal/gen

import (
	"fmt"
	"os"

	"github.com/hasdata-com/hasdata-cli/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit, date)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(cmd.ExitCodeFor(err))
	}
}
