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

func shellEscapeSingle(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

func (br *browseObj) runGrep() {
	if br.pattern == "" {
		br.printMessage("No search pattern", MSG_ORANGE)
		return
	}

	grepPath, err := exec.LookPath("grep")
	if err != nil || grepPath == "" {
		br.printMessage("Cannot find 'grep' in $PATH", MSG_ORANGE)
		return
	}

	brPath, err := exec.LookPath("browse")
	if err != nil || brPath == "" {
		br.printMessage("Cannot find 'browse' in $PATH", MSG_ORANGE)
		return
	}

	// case sensitivity
	grepOpts := grepDefaultOpts
	brOpts := ""
	if br.ignoreCase {
		grepOpts = grepIgnoreCase
		brOpts = browseIgnoreCase
	}

	title := fmt.Sprintf("grep %s -e \"%s\"", grepOpts, br.pattern)
	// Safely single-quote the pattern and title for shell
	patternArg := shellEscapeSingle(br.pattern)
	titleArg := shellEscapeSingle(title)

	cmd := fmt.Sprintf(
		"%s %s -e %s %s | %s %s -p %s -t %s",
		grepPath, grepOpts, patternArg, br.fileName,
		brPath, brOpts, patternArg, titleArg,
	)

	// Display command preview
	moveCursor(br.dispHeight, 1, true)
	fmt.Print("---\n", LINEWRAPON)
	fmt.Printf("$ %s\n", cmd)

	// Run command in a PTY
	fmt.Print(CURSAVE)
	resetScrRegion()
	fmt.Print(CURRESTORE)
	br.runInPty(cmd)
	br.resizeWindow()
	fmt.Print(LINEWRAPOFF)
}

// vim: set ts=4 sw=4 noet:
