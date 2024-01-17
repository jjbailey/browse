// bash.go
// run a command with bash
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var prevCommand string

func (x *browseObj) bashCommand() {
	// run a command with bash

	fmt.Printf("%s", LINEWRAPON)
	lbuf := x.userInput("!")

	if len(lbuf) > 0 {
		var err error
		var wstat syscall.WaitStatus

		ttyRestore()
		resetScrRegion()

		// substitute ! with the previous command
		sbuf := strings.Replace(lbuf, "!", prevCommand, -1)
		prevCommand = sbuf

		// substitute % with the current file name
		rbuf := strings.Replace(sbuf, "%", x.fileName, -1)

		// feedback
		movecursor(x.dispHeight, 1, true)
		fmt.Print("---\n")
		fmt.Printf("$ %s\n", rbuf)

		// set up env, run
		bashPath, err := exec.LookPath("bash")
		x.resetSignals()
		fmt.Printf("%s", LINEWRAPON) // again

		if err != nil {
			fmt.Printf("%v\n", err)
		} else {
			bashArgs := []string{"bash", "-c", rbuf}
			bashEnv := os.Environ()
			bashFiles := []uintptr{0, 1, 2}
			bashAttr := &syscall.ProcAttr{
				Dir:   ".",
				Env:   bashEnv,
				Files: bashFiles,
				Sys: &syscall.SysProcAttr{
					// fixme
					// Setsid: true,
					// Setpgid: true,
					// Setctty: true,
					Ctty: int(x.tty.Fd()),
					// Pgid: 0,
				},
			}

			pid, err := syscall.ForkExec(bashPath, bashArgs, bashAttr)

			if err != nil {
				fmt.Printf("%v\n", err)
			}

			syscall.Wait4(pid, &wstat, 0, nil)
		}
	}

	// cleanup
	x.catchSignals()
	x.userAnyKey(VIDMESSAGE + " Press any key to continue... " + VIDOFF)
	x.resizeWindow()
}

// vim: set ts=4 sw=4 noet:
