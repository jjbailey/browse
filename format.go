// format.go
// pipe the current file to fmt
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

// runFormat pipes the current search pattern to fmt and opens the results.
func (br *browseObj) runFormat() {
	// Check if file exists before attempting to format
	_, err := os.Stat(br.fileName)
	if os.IsNotExist(err) {
		br.printMessage("File does not exist", MSG_ORANGE)
		return
	}

	formatPath, err := exec.LookPath("fmt")
	if err != nil {
		br.printMessage("Cannot find 'fmt' in $PATH", MSG_ORANGE)
		return
	}

	brPath, err := os.Executable()
	if err != nil {
		br.printMessage("Cannot find 'browse' executable", MSG_ORANGE)
		return
	}

	title := "fmt -s"
	if !br.fromStdin {
		// aesthetical
		if br.title == filepath.Base(br.fileName) {
			title += " " + filepath.Base(br.fileName)
		} else {
			title += " " + abbreviateFileName(br.fileName, br.dispWidth>>1)
		}
	}
	titleArg := shellEscapeSingle(title)
	fileNameArg := shellEscapeSingle(br.fileName)
	formatPathArg := shellEscapeSingle(formatPath)
	brPathArg := shellEscapeSingle(brPath)

	// Use correct fmt command options with width parameter
	formatOpts := fmt.Sprintf("-s -w %d", maximum(10, br.dispWidth-NUMCOLWIDTH-1))

	// Construct command - fmt should pipe output to browse for display
	cmd := fmt.Sprintf("%s %s %s | %s -t %s",
		formatPathArg, formatOpts, fileNameArg, brPathArg, titleArg)

	if br.pattern != "" {
		cmd += " -p " + shellEscapeSingle(br.pattern)
	}

	// Display command preview
	moveCursor(br.dispHeight, 1, true)
	fmt.Printf("---\n%s$ %s\n", LINEWRAPON, cmd)

	// Run command in a PTY
	resetScrRegion()
	br.runInPty(cmd)
	br.resizeWindow()
}

// vim: set ts=4 sw=4 noet:
