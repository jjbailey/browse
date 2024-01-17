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
	"path"
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
			cmdArgs := []string{path.Base(bashPath), "-c", rbuf}
			cmdEnv := os.Environ()
			cmdFiles := []uintptr{0, 1, 2}
			cmdAttr := &syscall.ProcAttr{
				Dir:   ".",
				Env:   cmdEnv,
				Files: cmdFiles,
			}

			pid, err := syscall.ForkExec(bashPath, cmdArgs, cmdAttr)

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
