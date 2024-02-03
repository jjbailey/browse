// grep.go
// pipe the current search to grep -nP
//
// Copyright (c) 2024 jjb
// All rights reserved.

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

	title := fmt.Sprintf("grep -nP \"%s\"", x.pattern)

	// browse colors, pattern set
	// cmdbuf := fmt.Sprintf("grep -nP '%s' %s | %s -p '%s' -t '%s'",
	//  x.pattern, x.fileName,
	//  brPath, x.pattern, title)

	// grep colors, pattern not set
	cmdbuf := fmt.Sprintf("grep --color=always -nP '%s' %s | %s -t '%s'",
		x.pattern, x.fileName,
		brPath, title)

	fmt.Print("%s", LINEWRAPON)

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
