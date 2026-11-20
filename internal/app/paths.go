// Package app holds the OS glue: where files live, logging, notifications,
// autostart, opening files and links, and the loop that drives the guard.
package app

import (
	"os"
	"path/filepath"
	"runtime"
)

// Dir is where Karnize keeps config.yaml, the log and the state file.
func Dir() string {
	if d := os.Getenv("KARNIZE_DIR"); d != "" {
		return d // a separate profile for development and screenshots
	}
	base, err := os.UserConfigDir()
	if err != nil {
		base = os.TempDir()
	}
	name := "Karnize"
	if runtime.GOOS == "linux" {
		name = "karnize"
	}
	return filepath.Join(base, name)
}

func ConfigPath() string { return filepath.Join(Dir(), "config.yaml") }
func LogPath() string    { return filepath.Join(Dir(), "karnize.log") }
func StatePath() string  { return filepath.Join(Dir(), "state.json") }
func LockPath() string   { return filepath.Join(Dir(), "karnize.lock") }

const (
	RepoURL   = "https://github.com/CriptoRochy/Karnize"
	IssuesURL = RepoURL + "/issues"
	// UpstreamURL is the code Karnize is based on: Kordon 1.2.2 by Efimov Dev, GPL-3.0.
	UpstreamURL = "https://github.com/ef1m03/kordon/tree/v1.2.2"
)
