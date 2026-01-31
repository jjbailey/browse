// grep.go
// pipe the current search to grep -nP
//
// Copyright (c) 2024-2026 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// grep options for the external grep command and browse flags.
const (
	grepDefaultOpts  = "-nP"
	grepIgnoreCase   = "-inP"
	browseIgnoreCase = "-i"
)

// runGrep pipes the current search pattern to grep and opens the results.
func (br *browseObj) runGrep() {
	if br.pattern == "" {
		br.printMessage("No search pattern", MSG_ORANGE)
		return
	}

	// Check if file exists before attempting to grep
	_, err := os.Stat(br.fileName)
	if err != nil {
		br.printMessage("File does not exist", MSG_ORANGE)
		return
	}

	grepPath, err := exec.LookPath("grep")
	if err != nil {
		br.printMessage("Cannot find 'grep' in $PATH", MSG_ORANGE)
		return
	}

	brPath, err := os.Executable()
	if err != nil {
		br.printMessage("Cannot find 'browse' executable", MSG_ORANGE)
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
	if !br.fromStdin {
		title += " " + filepath.Base(br.fileName)
	}
	patternArg := shellEscapeSingle(br.pattern)
	titleArg := shellEscapeSingle(title)
	fileNameArg := shellEscapeSingle(br.fileName)
	grepPathArg := shellEscapeSingle(grepPath)
	brPathArg := shellEscapeSingle(brPath)

	// Construct command - grep should pipe output to browse for display
	cmd := fmt.Sprintf(
		"%s %s -e %s %s | %s %s -p %s -t %s",
		grepPathArg, grepOpts, patternArg, fileNameArg,
		brPathArg, brOpts, patternArg, titleArg,
	)

	// Display command preview
	moveCursor(br.dispHeight, 1, true)
	fmt.Printf("---\n%s$ %s\n", LINEWRAPON, cmd)

	// Run command in a PTY
	resetScrRegion()
	br.runInPty(cmd)
	br.resizeWindow()
}

// vim: set ts=4 sw=4 noet:
