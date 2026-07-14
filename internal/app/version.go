package app

import (
	"fmt"
	"io"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

func CurrentBuildInfo() BuildInfo {
	return BuildInfo{Version: Version, Commit: Commit, Date: Date}
}

func PrintVersion(w io.Writer, info BuildInfo) {
	fmt.Fprintf(w, "quickserve %s\ncommit: %s\nbuilt:  %s\n", info.Version, info.Commit, info.Date)
}
