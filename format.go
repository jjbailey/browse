// format.go
// pipe the current file to fmt
//
// Copyright (c) 2024-2025 jjb
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

func (br *browseObj) runFormat() {
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

	formatOpts := fmt.Sprintf("-s -w %d", maximum(10, br.dispWidth-NUMCOLWIDTH-1))

	title := "fmt -s"
	if !br.fromStdin {
		title += " " + filepath.Base(br.fileName)
	}
	titleArg := shellEscapeSingle(title)
	fileNameArg := shellEscapeSingle(br.fileName)

	cmd := fmt.Sprintf(
		"%s %s %s | %s -t %s",
		formatPath, formatOpts, fileNameArg,
		brPath, titleArg,
	)

	// Display command preview
	moveCursor(br.dispHeight, 1, true)
	fmt.Printf("---\n%s$ %s\n", LINEWRAPON, cmd)

	// Run command in a PTY
	fmt.Print(LINEWRAPON)
	fmt.Print(CURSAVE)
	resetScrRegion()
	fmt.Print(CURRESTORE)
	br.runInPty(cmd)
	br.resizeWindow()
	fmt.Print(LINEWRAPOFF)
}

// vim: set ts=4 sw=4 noet:
