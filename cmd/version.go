package main

import "fmt"

var (
	version = "unset"
	commit  = "unset"
	date    = "unset"
	builtBy = "unset"
)

// getVersion returns version information
func getVersion() string {
	return fmt.Sprintf("Version: %s, Commit: %s, Built date: %s, Built by: %s", version, commit, date, builtBy)
}
