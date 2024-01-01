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

func (x *browseObj) bashCommand() {
	lbuf := x.userInput("!")
	ttyRestore()

	rbuf := strings.Replace(lbuf, "%", x.fileName, -1)
	cmd := exec.Command("/bin/bash", "-c", rbuf)
	stdout, _ := cmd.Output()

	movecursor(x.dispHeight, 1, true)
	fmt.Print(string(stdout))
	x.userAnyKey(VIDBOLDREV + " Press any key to continue... " + VIDOFF)
}

// vim: set ts=4 sw=4 noet:
