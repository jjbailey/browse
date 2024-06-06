// grep.go
// pipe the current search to grep -nP
//
// Copyright (c) 2024 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"os/exec"
)

func (x *browseObj) grep() {
	if len(x.pattern) == 0 {
		x.warnMessage("No search pattern")
		return
	}

	brPath, err := exec.LookPath("browse")
	if len(brPath) == 0 || err != nil {
		x.warnMessage("Cannot find browse in $PATH")
		return
	}

	// run grep and browse in current search case mode
	grepOpts := "-nP"
	brOpts := ""

	if x.ignoreCase {
		grepOpts = "-inP"
		brOpts = "-i"
	}

	title := fmt.Sprintf("grep %s -e \"%s\"", grepOpts, x.pattern)

	cmdbuf := fmt.Sprintf("grep %s -e '%s' %s | %s %s -p '%s' -t '%s'",
		grepOpts, x.pattern, x.fileName, brPath, brOpts, x.pattern, title)

	fmt.Print(LINEWRAPON)

	// feedback
	moveCursor(x.dispHeight, 1, true)
	fmt.Print("---\n")
	fmt.Printf("$ %s\n", cmdbuf)

	// set up env, run
	resetScrRegion()
	x.runInPty(cmdbuf)
	x.resizeWindow()
}

// vim: set ts=4 sw=4 noet:
