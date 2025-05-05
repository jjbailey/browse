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
	"strings"
)

const (
	grepDefaultOpts  = "-nP"
	grepIgnoreCase   = "-inP"
	browseIgnoreCase = "-i"
)

func (br *browseObj) runGrep() {
	if br.pattern == "" {
		br.printMessage("No search pattern", MSG_ORANGE)
		return
	}

	grepPath, err := exec.LookPath("grep")
	if len(grepPath) == 0 || err != nil {
		br.printMessage("Cannot find grep in $PATH", MSG_ORANGE)
		return
	}

	brPath, err := exec.LookPath("browse")
	if len(brPath) == 0 || err != nil {
		br.printMessage("Cannot find browse in $PATH", MSG_ORANGE)
		return
	}

	// case sensitivity
	grepOpts := grepDefaultOpts
	brOpts := ""
	if br.ignoreCase {
		grepOpts = grepIgnoreCase
		brOpts = browseIgnoreCase
	}

	// build the command
	var cmd strings.Builder
	cmd.Grow(256)

	fmt.Fprintf(&cmd, "%s %s -e '%s' %s | %s %s",
		grepPath, grepOpts, br.pattern, br.fileName, brPath, brOpts)

	if br.pattern != "" {
		fmt.Fprintf(&cmd, " -p '%s'", br.pattern)
	}

	title := fmt.Sprintf("grep %s -e \"%s\"", grepOpts, br.pattern)
	fmt.Fprintf(&cmd, " -t '%s'", title)

	// feedback
	moveCursor(br.dispHeight, 1, true)
	fmt.Print("---\n", LINEWRAPON)
	fmt.Printf("$ %s\n", cmd.String())

	// set up env, run
	resetScrRegion()
	br.runInPty(cmd.String())
	br.resizeWindow()
	fmt.Print(LINEWRAPOFF)
}

// vim: set ts=4 sw=4 noet:
