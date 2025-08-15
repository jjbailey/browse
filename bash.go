// bash.go
// run a command with bash
//
// Copyright (c) 2024-2025 jjb
// All rights reserved.
//
// This source code is licensed under the MIT license found
// in the root directory of this source tree.

package main

import (
	"fmt"
	"strings"
)

var prevCommand string

func (br *browseObj) bashCommand() {
	// Run a command with bash

	moveCursor(br.dispHeight, 1, true)

	input, cancelled := userBashComp()
	if cancelled {
		br.pageCurrent()
		return
	}

	input = strings.TrimSpace(input)
	if input == "" {
		br.pageCurrent()
		return
	}

	// limit input length to prevent buffer overflows
	if len(input) > READBUFSIZ {
		br.printMessage("Command too long", MSG_RED)
		return
	}

	// substitute ! with the previous command
	bangbuf := subCommandChars(input, "!", prevCommand)
	prevCommand = bangbuf

	// substitute % with the current file name
	cmdbuf := subCommandChars(bangbuf, "%", `'`+br.fileName+`'`)

	if br.pattern != "" {
		// substitute & with the current search pattern
		cmdbuf = subCommandChars(cmdbuf, "&", `'`+br.pattern+`'`)
	}

	if cmdbuf == "" {
		br.pageCurrent()
		return
	}

	// Save command to history
	history := loadHistory(commHistory)
	history = append(history, cmdbuf)
	saveHistory(history, commHistory)

	// Display command preview
	fmt.Print(CURUP + CLEARLINE + CURUP + CLEARLINE)
	fmt.Print("---\n")
	fmt.Print(LINEWRAPON)
	fmt.Printf("$ %s\n", cmdbuf)

	// Run command in a PTY
	fmt.Print(CURSAVE)
	resetScrRegion()
	fmt.Print(CURRESTORE)
	br.runInPty(cmdbuf)
	br.resizeWindow()
	fmt.Print(LINEWRAPOFF)
}

// vim: set ts=4 sw=4 noet:
