// manpage.go
// browse the man page
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func (br *browseObj) manPage() {
	manPath, err := exec.LookPath("man")
	if err != nil || manPath == "" {
		br.printMessage("Cannot find 'man' in $PATH", MSG_ORANGE)
		return
	}

	// Check if man page for 'browse' exists
	cmdOut, err := exec.Command(manPath, "-w", "browse").CombinedOutput()

	if err != nil && !bytes.Contains(cmdOut, []byte{'/'}) {
		msg := string(bytes.TrimSpace(cmdOut))
		br.printMessage(msg, MSG_ORANGE)
		return
	}

	brPath, err := exec.LookPath("browse")
	if err != nil || brPath == "" {
		br.printMessage("Cannot find 'browse' in $PATH", MSG_ORANGE)
		return
	}

	cmdStr := manPath + " browse | " + brPath + " -t browse.1"

	// Run command in a PTY
	fmt.Print(LINEWRAPON)
	fmt.Print(CURSAVE)
	resetScrRegion()
	fmt.Print(CURRESTORE)
	br.runInPty(cmdStr)
	br.resizeWindow()
	fmt.Print(LINEWRAPOFF)
}

// vim: set ts=4 sw=4 noet:
