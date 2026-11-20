package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// The .exe is built as a GUI program (no console window flashing at login), so
// when it's run from a terminal with a command, it borrows that terminal for output.
// Output that was redirected (`karnize status > status.txt`, a pipe) already has a
// handle and is left alone: writing to CONOUT$ instead would leave the file empty.
func attachConsole() {
	if len(os.Args) < 2 {
		return
	}
	valid := func(f *os.File) bool {
		h := windows.Handle(f.Fd())
		return h != 0 && h != windows.InvalidHandle
	}
	outOK, errOK := valid(os.Stdout), valid(os.Stderr)
	if outOK && errOK {
		return
	}
	k := windows.NewLazySystemDLL("kernel32.dll")
	const attachParent = ^uintptr(0) // ATTACH_PARENT_PROCESS
	if r, _, _ := k.NewProc("AttachConsole").Call(attachParent); r == 0 {
		return
	}
	// Go writes UTF-8; without this, Cyrillic comes out garbled in PowerShell and cmd.
	const cpUTF8 = 65001
	_ = windows.SetConsoleOutputCP(cpUTF8)
	if f, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
		if !outOK {
			os.Stdout = f
		}
		if !errOK {
			os.Stderr = f
		}
	}
}

func processAlive(pid int) bool {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var code uint32
	return windows.GetExitCodeProcess(h, &code) == nil && code == 259 // STILL_ACTIVE
}

// stopProcess: there is no SIGTERM on Windows, and a forced kill would leave
// apps frozen, so the user restarts Karnize from the menu.
func stopProcess(int) bool { return false }
