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
	"regexp"
	"syscall"
)

var prevCommand string

func (x *browseObj) bashCommand() {
	// run a command with bash

	fmt.Printf("%s", LINEWRAPON)
	input := x.userInput("!")

	if len(input) > 0 {
		var wstat syscall.WaitStatus

		ttyRestore()
		resetScrRegion()

		// substitute ! with the previous command
		bangbuf := subCommandChars(input, "!", prevCommand)

		// save
		prevCommand = bangbuf

		// substitute % with the current file name
		cmdbuf := subCommandChars(bangbuf, "%", x.fileName)

		if len(cmdbuf) > 0 {
			// feedback
			movecursor(x.dispHeight, 1, true)
			fmt.Print("---\n")
			fmt.Printf("$ %s\n", cmdbuf)

			// set up env, run
			bashPath, err := exec.LookPath("bash")
			x.resetSignals()
			fmt.Printf("%s", LINEWRAPON) // again

			if err != nil {
				fmt.Printf("%v\n", err)
			} else {
				cmdArgs := []string{path.Base(bashPath), "-c", cmdbuf}
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
	}

	// cleanup
	x.catchSignals()
	x.userAnyKey(VIDMESSAGE + " Press any key to continue... " + VIDOFF)
	x.resizeWindow()
}

func subCommandChars(input, char, repl string) string {
	var rbuf1 string

	// negative lookbehind doesn't compile, e.g.
	// pattern := `(?<!\\)%`

	pattern := `(^|[^\\])` + char
	replstr := "${1}" + repl

	re, err := regexp.Compile(pattern)

	if err != nil {
		return ""
	}

	rbuf1 = input

	for {
		rbuf2 := re.ReplaceAllString(rbuf1, replstr)

		if rbuf2 == rbuf1 {
			break
		}

		rbuf1 = rbuf2
	}

	return rbuf1
}

// vim: set ts=4 sw=4 noet:
