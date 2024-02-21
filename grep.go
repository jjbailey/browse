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
		x.printMessage("No search pattern")
		return
	}

	brPath, err := exec.LookPath("browse")

	if len(brPath) == 0 || err != nil {
		x.printMessage("Cannot find browse")
		return
	}

	title := fmt.Sprintf("grep -nP -e \"%s\"", x.pattern)

	// browse colors, pattern set
	cmdbuf := fmt.Sprintf("grep -nP -e '%s' %s | %s -p '%s' -t '%s'",
		x.pattern, x.fileName,
		brPath, x.pattern, title)

	// grep colors, pattern not set
	// cute, but browse gets the line lengths wrong
	//	cmdbuf := fmt.Sprintf("grep --color=always -nP -e '%s' %s | %s -t '%s'",
	//		x.pattern, x.fileName,
	//		brPath, title)

	fmt.Print(LINEWRAPON)

	// feedback
	movecursor(x.dispHeight, 1, true)
	fmt.Print("---\n")
	fmt.Printf("$ %s\n", cmdbuf)

	// set up env, run
	resetScrRegion()
	x.runInPty(cmdbuf)
	x.resizeWindow()
}

// vim: set ts=4 sw=4 noet:
