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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func (br *browseObj) manPage() {
	manPath, err := exec.LookPath("man")
	if err != nil || manPath == "" {
		br.printMessage("Cannot find 'man' in $PATH", MSG_ORANGE)
		return
	}

	brPath, err := exec.LookPath("browse")
	if err != nil || brPath == "" {
		br.printMessage("Cannot find 'browse' in $PATH", MSG_ORANGE)
		return
	}

	manpageFile := "/usr/local/man/man1/browse.1.gz"
	if _, err := os.Stat(manpageFile); os.IsNotExist(err) {
		br.printMessage(fmt.Sprintf("Manpage '%s' not found",
			filepath.Base(manpageFile)), MSG_ORANGE)
		return
	}

	cmd := manPath + " browse | " + brPath + " -t browse.1"

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
