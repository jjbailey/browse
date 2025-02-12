// grep.go
// pipe the current search to grep -nP
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"os/exec"
)

func (br *browseObj) grep() {
	if len(br.pattern) == 0 {
		br.printMessage("No search pattern", MSG_ORANGE)
		return
	}

	brPath, err := exec.LookPath("browse")
	if len(brPath) == 0 || err != nil {
		br.printMessage("Cannot find browse in $PATH", MSG_ORANGE)
		return
	}

	// run grep and browse in current search case mode
	grepOpts := "-nP"
	brOpts := ""

	if br.ignoreCase {
		grepOpts = "-inP"
		brOpts = "-i"
	}

	title := fmt.Sprintf("grep %s -e \"%s\"", grepOpts, br.pattern)

	cmdbuf := fmt.Sprintf("grep %s -e '%s' %s | %s %s -p '%s' -t '%s'",
		grepOpts, br.pattern, br.fileName, brPath, brOpts, br.pattern, title)

	fmt.Print(LINEWRAPON)

	// feedback
	moveCursor(br.dispHeight, 1, true)
	fmt.Print("---\n")
	fmt.Printf("$ %s\n", cmdbuf)

	// set up env, run
	resetScrRegion()
	br.runInPty(cmdbuf)
	br.resizeWindow()
}

// vim: set ts=4 sw=4 noet:
