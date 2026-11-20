package app

import (
	"context"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/ef1m03/kordon/internal/config"
)

// A config read while an editor is still writing it must never be applied:
// an empty file parses as the defaults (Claude protected, nothing allowed).
func TestReloadIgnoresEmptyAndUnsettledConfig(t *testing.T) {
	t.Setenv("KORDON_DIR", t.TempDir())
	cfg := config.Default()
	cfg.Apps = []string{"chatgpt"}
	if err := config.Save(ConfigPath(), cfg); err != nil {
		t.Fatal(err)
	}
	r := NewRunner(context.Background(), cfg, false)
	apps := func() []string { return r.G.Snapshot(time.Now()).Config.Apps }
	write := func(data string, age time.Duration) time.Time {
		t.Helper()
		if err := os.WriteFile(ConfigPath(), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
		mod := time.Now().Add(-age)
		if err := os.Chtimes(ConfigPath(), mod, mod); err != nil {
			t.Fatal(err)
		}
		return mod
	}

	write("", 5*time.Second) // truncated by the editor, content not written yet
	r.reloadIfEdited(time.Now())
	if got := apps(); !slices.Equal(got, []string{"chatgpt"}) {
		t.Fatalf("empty config applied: apps = %v", got)
	}

	mod := write("apps: [cursor]\n", 0) // just written: may still be in progress
	r.reloadIfEdited(mod.Add(100 * time.Millisecond))
	if got := apps(); !slices.Equal(got, []string{"chatgpt"}) {
		t.Fatalf("config applied before it settled: apps = %v", got)
	}
	r.reloadIfEdited(mod.Add(2 * time.Second))
	if got := apps(); !slices.Equal(got, []string{"cursor"}) {
		t.Fatalf("settled config not applied: apps = %v", got)
	}
}
