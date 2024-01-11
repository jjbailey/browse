// bash.go
// run a command with bash
// note: the process has no tty
//
// Copyright (c) 2024 jjb
// All rights reserved.

package main

import (
	"fmt"
	"os/exec"
	"strings"
)

var prevCommand string

func (x *browseObj) bashCommand() {
	// run a command with bash

	fmt.Printf("%s", LINEWRAPON)
	lbuf := x.userInput("!")

	if len(lbuf) > 0 {
		ttyRestore()
		resetScrRegion()

		// substitute ! with the previous command
		sbuf := strings.Replace(lbuf, "!", prevCommand, -1)
		prevCommand = sbuf

		// substitute % with the current file name
		rbuf := strings.Replace(sbuf, "%", x.fileName, -1)

		cmd := exec.Command("/bin/bash", "-c", rbuf)
		stdout, _ := cmd.Output()
		movecursor(x.dispHeight, 1, true)
		fmt.Print("---\n")
		fmt.Printf("$ %s\n", rbuf)
		fmt.Print(string(stdout))
	}

	x.userAnyKey(VIDMESSAGE + " Press any key to continue... " + VIDOFF)
}

// vim: set ts=4 sw=4 noet:
