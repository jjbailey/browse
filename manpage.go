// manpage.go
// browse the man page
//
// Copyright (c) 2024-2026 jjb
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

// manPage renders the browse man page inside the browse interface.
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

	cmd := manPath + " browse | " + brPath + " -t browse.1"

	// Display command preview
	moveCursor(br.dispHeight, 1, true)
	fmt.Printf("---\n%s$ %s\n", LINEWRAPON, cmd)

	// Run command in a PTY
	resetScrRegion()
	br.runInPty(cmd)
	br.resizeWindow()
}

// vim: set ts=4 sw=4 noet:
