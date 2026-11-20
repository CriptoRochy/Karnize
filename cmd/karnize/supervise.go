package main

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/CriptoRochy/Karnize/internal/app"
	"github.com/CriptoRochy/Karnize/internal/config"
	"github.com/CriptoRochy/Karnize/internal/procs"
	"github.com/CriptoRochy/Karnize/internal/ui"
)

const childEnv = "KARNIZE_CHILD"

// supervise runs Karnize as a child process and restarts it if it crashes, so a
// bug can't silently leave the apps unprotected. A clean exit (Quit in the menu,
// or "already running") ends the supervisor too. Returns false in the child.
func supervise() bool {
	if os.Getenv(childEnv) != "" {
		return false
	}
	exe, err := os.Executable()
	if err != nil {
		return false // can't re-exec: run unsupervised rather than not at all
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	var crashes []time.Time
	for {
		cmd := exec.Command(exe, os.Args[1:]...)
		cmd.Env = append(os.Environ(), childEnv+"=1")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Start(); err != nil {
			return false
		}
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()

		var err error
		select {
		case err = <-done:
		case s := <-sig:
			_ = cmd.Process.Signal(s) // let the child un-freeze apps and exit
			<-done
			return true
		}
		var exit *exec.ExitError
		if err == nil || !errors.As(err, &exit) {
			return true
		}

		now := time.Now()
		crashes = append(crashes, now)
		for len(crashes) > 0 && now.Sub(crashes[0]) > time.Minute {
			crashes = crashes[1:]
		}
		app.Logf("Karnize упал (%v), перезапускаю", err)
		if len(crashes) > 5 {
			app.Logf("Karnize падает слишком часто, останавливаюсь")
			releaseApps()
			app.Notify(ui.T("crash.t"), ui.T("crash.b", app.LogPath()))
			return true
		}
		time.Sleep(time.Second)
	}
}

// releaseApps un-freezes the protected apps when the supervisor gives up: the crashed
// child can't do it any more, and frozen apps (a browser, say) would otherwise stay
// stuck until a reboot and look like the internet is gone.
func releaseApps() {
	cfg := config.Default()
	if data, err := os.ReadFile(app.ConfigPath()); err == nil {
		if c, err := config.Parse(data); err == nil {
			cfg = c
		}
	}
	ctl := procs.New()
	all, err := ctl.List()
	if err != nil {
		return
	}
	for _, p := range procs.Filter(all, procs.Targets(cfg), os.Getpid()) {
		_ = ctl.Resume(p)
	}
}
